//go:build viewer3d

package viewer3d

import (
	"encoding/binary"
	sysMath "math"

	"github.com/gogpu/gputypes"
	"github.com/gogpu/wgpu"

	"github.com/dkrutsko/oasis/game"
	"github.com/dkrutsko/oasis/geometry"
	"github.com/dkrutsko/oasis/maps"
	"github.com/dkrutsko/oasis/math"
)

////////////////////////////////////////////////////////////////////////////////

// Maximum entity vertex buffer size in bytes. Enough for
// 64 entities with ~6 lines each (64 * 6 * 2 * 24 = 18KB).
const entityBufSize = 65536

////////////////////////////////////////////////////////////////////////////////

// Renderer manages all WebGPU resources and draw calls
// for the 3D viewer.
type Renderer struct {
	dev   *wgpu.Device
	queue *wgpu.Queue

	mapPipe    *wgpu.RenderPipeline
	entityPipe *wgpu.RenderPipeline
	pipeLayout *wgpu.PipelineLayout
	bgLayout   *wgpu.BindGroupLayout

	uniformBuf *wgpu.Buffer
	bindGroup  *wgpu.BindGroup

	mapBuf      *wgpu.Buffer
	mapVertices uint32

	entityBuf      *wgpu.Buffer
	entityVertices uint32

	hudUniformBuf *wgpu.Buffer
	hudBindGroup  *wgpu.BindGroup
	textBuf       *wgpu.Buffer
	textVertices  uint32

	lastTeam int32
}

////////////////////////////////////////////////////////////////////////////////

func (r *Renderer) Init(dev *wgpu.Device, format gputypes.TextureFormat) error {

	//----------------------------------------------------------------------------//

	r.dev = dev
	r.queue = dev.Queue()

	//----------------------------------------------------------------------------//

	// Bind group layout: one uniform buffer (MVP matrix)
	var err error
	r.bgLayout, err = dev.CreateBindGroupLayout(&wgpu.BindGroupLayoutDescriptor{
		Entries: []wgpu.BindGroupLayoutEntry{{
			Binding:    0,
			Visibility: wgpu.ShaderStageVertex | wgpu.ShaderStageFragment,
			Buffer:     &gputypes.BufferBindingLayout{Type: gputypes.BufferBindingTypeUniform},
		}},
	})
	if err != nil {
		return err
	}

	r.pipeLayout, err = dev.CreatePipelineLayout(&wgpu.PipelineLayoutDescriptor{
		BindGroupLayouts: []*wgpu.BindGroupLayout{r.bgLayout},
	})
	if err != nil {
		return err
	}

	//----------------------------------------------------------------------------//

	// Uniform buffer (64 bytes for mat4x4)
	r.uniformBuf, err = dev.CreateBuffer(&wgpu.BufferDescriptor{
		Label: "viewer3d_uniform",
		Size:  64,
		Usage: wgpu.BufferUsageUniform | wgpu.BufferUsageCopyDst,
	})
	if err != nil {
		return err
	}

	r.bindGroup, err = dev.CreateBindGroup(&wgpu.BindGroupDescriptor{
		Layout: r.bgLayout,
		Entries: []wgpu.BindGroupEntry{{
			Binding: 0,
			Buffer:  r.uniformBuf,
			Size:    64,
		}},
	})
	if err != nil {
		return err
	}

	//----------------------------------------------------------------------------//

	// Map pipeline: solid triangles with alpha blending
	mapShader, err := dev.CreateShaderModule(&wgpu.ShaderModuleDescriptor{WGSL: mapShaderWGSL})
	if err != nil {
		return err
	}
	defer mapShader.Release()

	r.mapPipe, err = dev.CreateRenderPipeline(&wgpu.RenderPipelineDescriptor{
		Label:  "viewer3d_map",
		Layout: r.pipeLayout,
		Vertex: wgpu.VertexState{
			Module:     mapShader,
			EntryPoint: "vs_main",
			Buffers: []wgpu.VertexBufferLayout{{
				ArrayStride: 12,
				StepMode:    gputypes.VertexStepModeVertex,
				Attributes: []gputypes.VertexAttribute{{
					Format:         gputypes.VertexFormatFloat32x3,
					Offset:         0,
					ShaderLocation: 0,
				}},
			}},
		},
		Primitive: gputypes.PrimitiveState{
			Topology: gputypes.PrimitiveTopologyTriangleList,
		},
		Fragment: &wgpu.FragmentState{
			Module:     mapShader,
			EntryPoint: "fs_main",
			Targets: []gputypes.ColorTargetState{{
				Format:    format,
				WriteMask: gputypes.ColorWriteMaskAll,
				Blend: &gputypes.BlendState{
					Color: gputypes.BlendComponent{
						SrcFactor: gputypes.BlendFactorSrcAlpha,
						DstFactor: gputypes.BlendFactorOneMinusSrcAlpha,
						Operation: gputypes.BlendOperationAdd,
					},
					Alpha: gputypes.BlendComponent{
						SrcFactor: gputypes.BlendFactorOne,
						DstFactor: gputypes.BlendFactorOneMinusSrcAlpha,
						Operation: gputypes.BlendOperationAdd,
					},
				},
			}},
		},
	})
	if err != nil {
		return err
	}

	//----------------------------------------------------------------------------//

	// Entity pipeline: colored lines, opaque
	entShader, err := dev.CreateShaderModule(&wgpu.ShaderModuleDescriptor{WGSL: entityShaderWGSL})
	if err != nil {
		return err
	}
	defer entShader.Release()

	r.entityPipe, err = dev.CreateRenderPipeline(&wgpu.RenderPipelineDescriptor{
		Label:  "viewer3d_entity",
		Layout: r.pipeLayout,
		Vertex: wgpu.VertexState{
			Module:     entShader,
			EntryPoint: "vs_main",
			Buffers: []wgpu.VertexBufferLayout{{
				ArrayStride: 28,
				StepMode:    gputypes.VertexStepModeVertex,
				Attributes: []gputypes.VertexAttribute{
					{Format: gputypes.VertexFormatFloat32x3, Offset: 0, ShaderLocation: 0},
					{Format: gputypes.VertexFormatFloat32x4, Offset: 12, ShaderLocation: 1},
				},
			}},
		},
		Primitive: gputypes.PrimitiveState{
			Topology: gputypes.PrimitiveTopologyTriangleList,
		},
		Fragment: &wgpu.FragmentState{
			Module:     entShader,
			EntryPoint: "fs_main",
			Targets: []gputypes.ColorTargetState{{
				Format:    format,
				WriteMask: gputypes.ColorWriteMaskAll,
				Blend: &gputypes.BlendState{
					Color: gputypes.BlendComponent{
						SrcFactor: gputypes.BlendFactorSrcAlpha,
						DstFactor: gputypes.BlendFactorOneMinusSrcAlpha,
						Operation: gputypes.BlendOperationAdd,
					},
					Alpha: gputypes.BlendComponent{
						SrcFactor: gputypes.BlendFactorOne,
						DstFactor: gputypes.BlendFactorOneMinusSrcAlpha,
						Operation: gputypes.BlendOperationAdd,
					},
				},
			}},
		},
	})
	if err != nil {
		return err
	}

	//----------------------------------------------------------------------------//

	// Pre-allocate entity vertex buffer
	r.entityBuf, err = dev.CreateBuffer(&wgpu.BufferDescriptor{
		Label: "viewer3d_entities",
		Size:  entityBufSize,
		Usage: wgpu.BufferUsageVertex | wgpu.BufferUsageCopyDst,
	})
	if err != nil {
		return err
	}

	//----------------------------------------------------------------------------//

	// HUD: identity MVP for screen-space text
	r.hudUniformBuf, err = dev.CreateBuffer(&wgpu.BufferDescriptor{
		Label: "viewer3d_hud_uniform",
		Size:  64,
		Usage: wgpu.BufferUsageUniform | wgpu.BufferUsageCopyDst,
	})
	if err != nil {
		return err
	}

	identityData := math.Matrix4Identity.ToSlice32()
	identityBytes := make([]byte, 64)
	for i, v := range identityData {
		binary.LittleEndian.PutUint32(identityBytes[i*4:], sysMath.Float32bits(v))
	}
	r.queue.WriteBuffer(r.hudUniformBuf, 0, identityBytes)

	r.hudBindGroup, err = dev.CreateBindGroup(&wgpu.BindGroupDescriptor{
		Layout: r.bgLayout,
		Entries: []wgpu.BindGroupEntry{{
			Binding: 0,
			Buffer:  r.hudUniformBuf,
			Size:    64,
		}},
	})
	if err != nil {
		return err
	}

	r.textBuf, err = dev.CreateBuffer(&wgpu.BufferDescriptor{
		Label: "viewer3d_text",
		Size:  8192,
		Usage: wgpu.BufferUsageVertex | wgpu.BufferUsageCopyDst,
	})
	if err != nil {
		return err
	}

	//----------------------------------------------------------------------------//

	return nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// LoadMap uploads the map collision geometry to the GPU.
func (r *Renderer) LoadMap(m *maps.Map) error {

	triangles := m.GetTriangles()
	if len(triangles) == 0 {
		return nil
	}

	// Flatten to position-only float32 buffer
	data := make([]byte, len(triangles)*9*4)
	for i := range triangles {
		tri := &triangles[i]
		off := i * 36
		putVec3(data[off:], tri.V0)
		putVec3(data[off+12:], tri.V1)
		putVec3(data[off+24:], tri.V2)
	}

	r.mapVertices = uint32(len(triangles) * 3)

	// Release old buffer if any
	if r.mapBuf != nil {
		r.mapBuf.Release()
	}

	var err error
	r.mapBuf, err = r.dev.CreateBuffer(&wgpu.BufferDescriptor{
		Label: "viewer3d_map",
		Size:  uint64(len(data)),
		Usage: wgpu.BufferUsageVertex | wgpu.BufferUsageCopyDst,
	})
	if err != nil {
		return err
	}

	r.queue.WriteBuffer(r.mapBuf, 0, data)
	return nil
}

////////////////////////////////////////////////////////////////////////////////

// UnloadMap clears the map geometry from the GPU.
func (r *Renderer) UnloadMap() {

	if r.mapBuf != nil {
		r.mapBuf.Release()
		r.mapBuf = nil
	}
	r.mapVertices = 0
}

////////////////////////////////////////////////////////////////////////////////

// UpdateEntities rebuilds the entity line buffer from
// the current action state. The camera eye position is
// used to billboard line quads so they always face the
// viewer.
func (r *Renderer) UpdateEntities(action *game.ActionState, camEye math.Vector3, currentMap *maps.Map) {

	//----------------------------------------------------------------------------//

	if action == nil {
		r.entityVertices = 0
		return
	}

	var localTeam int32
	var playerZ float64
	hasPlayer := action.Player != nil

	if hasPlayer {
		localTeam = action.Player.Team
		playerZ = action.Player.Origin.Z
		r.lastTeam = localTeam
	} else {
		localTeam = r.lastTeam
	}

	// Build vertex data: [x y z r g b a] per vertex
	buf := make([]byte, 0, entityBufSize)

	//----------------------------------------------------------------------------//

	for i := range action.Entities {
		entity := &action.Entities[i]

		if !entity.Valid || entity.Team <= 1 {
			continue
		}

		isPlayer := hasPlayer && entity == action.Player
		isEnemy := localTeam != 0 && entity.Team != localTeam

		// Only show the local player and enemies
		if !isPlayer && !isEnemy {
			continue
		}

		// Pick color based on relationship
		var cr, cg, cb, ca float32

		if isPlayer {
			cr, cg, cb, ca = 0.3, 0.5, 1.0, 1.0
		} else if entity.Health > 70 {
			cr, cg, cb, ca = 0.196, 0.863, 0.196, 1.0
		} else if entity.Health > 30 {
			cr, cg, cb, ca = 1.0, 0.627, 0.118, 1.0
		} else {
			cr, cg, cb, ca = 1.0, 0.196, 0.196, 1.0
		}

		// Fade enemies based on height difference from
		// the local player. Same floor stays opaque,
		// then drops sharply past ~128 units (one floor).
		if isEnemy && hasPlayer {
			heightDiff := sysMath.Abs(entity.Origin.Z - playerZ)
			t := heightDiff / 128.0
			if t < 0.5 {
				// Same floor: fully opaque
			} else if t < 1.0 {
				// Sharp falloff from 1.0 to 0.15
				ca *= float32(1.0 - (t-0.5)*1.4)
			} else {
				ca *= 0.3
			}
		}

		// Use bone head position when available
		headPos := entity.Head
		if entity.Bones.Valid && !entity.Bones.Pos[game.BoneHead].IsZero() {
			headPos = entity.Bones.Pos[game.BoneHead]
		}

		yawRad := entity.Angles.Y * sysMath.Pi / 180.0
		pitchRad := entity.Angles.X * sysMath.Pi / 180.0
		sin90 := sysMath.Sin(yawRad + sysMath.Pi/2)
		cos90 := sysMath.Cos(yawRad + sysMath.Pi/2)

		feetZ := entity.Origin.Z

		// Spine line: feet to head
		buf = appendLineQuad(buf, camEye,
			headPos.X, headPos.Y, feetZ,
			headPos.X, headPos.Y, headPos.Z,
			cr, cg, cb, ca,
		)

		// Shoulder bar: perpendicular to facing at feet height
		buf = appendLineQuad(buf, camEye,
			headPos.X-20*cos90, headPos.Y-20*sin90, feetZ,
			headPos.X+20*cos90, headPos.Y+20*sin90, feetZ,
			cr, cg, cb, ca,
		)

		// View direction: from head
		dirX := sysMath.Cos(pitchRad) * sysMath.Cos(yawRad)
		dirY := sysMath.Cos(pitchRad) * sysMath.Sin(yawRad)
		dirZ := -sysMath.Sin(pitchRad)

		buf = appendLineQuad(buf, camEye,
			headPos.X, headPos.Y, headPos.Z,
			headPos.X+dirX*10, headPos.Y+dirY*10, headPos.Z+dirZ*10,
			cr, cg, cb, ca,
		)
	}

	//----------------------------------------------------------------------------//

	// Aim ray from the local player's eye position
	if hasPlayer {
		player := action.Player
		var eyePos math.Vector3
		if player.Bones.Valid && !player.Bones.Pos[game.BoneHead].IsZero() {
			eyePos = player.Bones.Pos[game.BoneHead]
		} else {
			eyePos = math.Vector3{
				X: player.Origin.X,
				Y: player.Origin.Y,
				Z: player.Origin.Z + 64.0,
			}
		}

		yaw := player.Angles.Y * sysMath.Pi / 180.0
		pitch := player.Angles.X * sysMath.Pi / 180.0
		aimDir := math.Vector3{
			X: sysMath.Cos(pitch) * sysMath.Cos(yaw),
			Y: sysMath.Cos(pitch) * sysMath.Sin(yaw),
			Z: -sysMath.Sin(pitch),
		}

		ray := geometry.Ray{Origin: eyePos, Direction: aimDir}
		const maxRayDist = 50000.0

		// Find where the ray hits map geometry
		hitDist, hasHit := float64(0), false
		if currentMap != nil {
			hitDist, hasHit = currentMap.Trace(ray, maxRayDist)
		}

		// Find where the ray exits the map bounds
		exitDist := maxRayDist
		if currentMap != nil {
			bounds := currentMap.GetBounds()
			if d, ok := ray.IntersectBox(bounds); ok && d > 0 {
				exitDist = d
			}
		}

		// Solid ray from eye to hit point (or exit if no hit)
		solidEnd := exitDist
		if hasHit {
			solidEnd = hitDist
		}

		hitPoint := ray.GetPoint(solidEnd)
		buf = appendLineQuad(buf, camEye,
			eyePos.X, eyePos.Y, eyePos.Z,
			hitPoint.X, hitPoint.Y, hitPoint.Z,
			1.0, 1.0, 0.3, 0.5,
		)

		// Dotted line from hit point to map bounds exit
		if hasHit && exitDist > hitDist {
			const dashLen = 40.0
			const gapLen = 40.0

			for d := hitDist + gapLen; d < exitDist; d += dashLen + gapLen {
				endD := d + dashLen
				if endD > exitDist {
					endD = exitDist
				}
				p1 := ray.GetPoint(d)
				p2 := ray.GetPoint(endD)
				buf = appendLineQuad(buf, camEye,
					p1.X, p1.Y, p1.Z,
					p2.X, p2.Y, p2.Z,
					1.0, 1.0, 0.3, 0.2,
				)
			}
		}
	}

	//----------------------------------------------------------------------------//

	r.entityVertices = uint32(len(buf) / 28)

	if r.entityVertices > 0 && len(buf) <= entityBufSize {
		r.queue.WriteBuffer(r.entityBuf, 0, buf)
	}

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// UpdateHud builds the HUD indicator. Shows a green
// circle in the top-right when auto mode is active.
// Viewport width/height are needed to correct for aspect
// ratio so the circle isn't stretched into an oval.
func (r *Renderer) UpdateHud(autoMode bool, vpW, vpH int) {

	if !autoMode {
		r.textVertices = 0
		return
	}

	// Filled circle in NDC. Top-right corner with margin.
	// Scale X radius by height/width to counteract the
	// non-square NDC coordinate space.
	const (
		cx       = 0.92
		cy       = 0.90
		radius   = 0.015
		segments = 24
	)

	aspect := 1.0
	if vpW > 0 && vpH > 0 {
		aspect = float64(vpH) / float64(vpW)
	}

	var cr, cg, cb, ca float32 = 0.2, 0.9, 0.2, 1.0
	buf := make([]byte, 0, segments*3*28)

	for i := 0; i < segments; i++ {
		a1 := float64(i) * 2 * sysMath.Pi / float64(segments)
		a2 := float64(i+1) * 2 * sysMath.Pi / float64(segments)

		buf = appendVertex(buf, cx, cy, 0, cr, cg, cb, ca)
		buf = appendVertex(buf, cx+radius*aspect*sysMath.Cos(a1), cy+radius*sysMath.Sin(a1), 0, cr, cg, cb, ca)
		buf = appendVertex(buf, cx+radius*aspect*sysMath.Cos(a2), cy+radius*sysMath.Sin(a2), 0, cr, cg, cb, ca)
	}

	r.textVertices = uint32(len(buf) / 28)
	if r.textVertices > 0 {
		r.queue.WriteBuffer(r.textBuf, 0, buf)
	}
}

////////////////////////////////////////////////////////////////////////////////

// Draw renders the map and entities using the given MVP.
func (r *Renderer) Draw(sv *wgpu.TextureView, mvp math.Matrix4) error {

	//----------------------------------------------------------------------------//

	// Upload MVP matrix
	mvpData := mvp.ToSlice32()
	mvpBytes := make([]byte, 64)
	for i, v := range mvpData {
		binary.LittleEndian.PutUint32(mvpBytes[i*4:], sysMath.Float32bits(v))
	}
	r.queue.WriteBuffer(r.uniformBuf, 0, mvpBytes)

	//----------------------------------------------------------------------------//

	enc, err := r.dev.CreateCommandEncoder(nil)
	if err != nil {
		return err
	}

	rp, err := enc.BeginRenderPass(&wgpu.RenderPassDescriptor{
		ColorAttachments: []wgpu.RenderPassColorAttachment{{
			View:       sv,
			LoadOp:     gputypes.LoadOpClear,
			StoreOp:    gputypes.StoreOpStore,
			ClearValue: gputypes.Color{R: 0.08, G: 0.08, B: 0.10, A: 1},
		}},
	})
	if err != nil {
		return err
	}

	//----------------------------------------------------------------------------//

	// Draw map (transparent solid triangles)
	if r.mapVertices > 0 && r.mapBuf != nil {
		rp.SetPipeline(r.mapPipe)
		rp.SetBindGroup(0, r.bindGroup, nil)
		rp.SetVertexBuffer(0, r.mapBuf, 0)
		rp.Draw(r.mapVertices, 1, 0, 0)
	}

	// Draw entities (opaque colored lines)
	if r.entityVertices > 0 {
		rp.SetPipeline(r.entityPipe)
		rp.SetBindGroup(0, r.bindGroup, nil)
		rp.SetVertexBuffer(0, r.entityBuf, 0)
		rp.Draw(r.entityVertices, 1, 0, 0)
	}

	// Draw HUD text (screen-space, identity MVP)
	if r.textVertices > 0 {
		rp.SetPipeline(r.entityPipe)
		rp.SetBindGroup(0, r.hudBindGroup, nil)
		rp.SetVertexBuffer(0, r.textBuf, 0)
		rp.Draw(r.textVertices, 1, 0, 0)
	}

	//----------------------------------------------------------------------------//

	rp.End()

	cmds, err := enc.Finish()
	if err != nil {
		return err
	}

	r.queue.Submit(cmds)
	return nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// Destroy releases all GPU resources.
func (r *Renderer) Destroy() {

	for _, res := range []interface{ Release() }{
		r.mapPipe, r.entityPipe, r.pipeLayout, r.bgLayout,
		r.uniformBuf, r.bindGroup, r.mapBuf, r.entityBuf,
		r.hudUniformBuf, r.hudBindGroup, r.textBuf,
	} {
		if res != nil {
			res.Release()
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

func putVec3(dst []byte, v math.Vector3) {

	binary.LittleEndian.PutUint32(dst[0:], sysMath.Float32bits(float32(v.X)))
	binary.LittleEndian.PutUint32(dst[4:], sysMath.Float32bits(float32(v.Y)))
	binary.LittleEndian.PutUint32(dst[8:], sysMath.Float32bits(float32(v.Z)))
}

////////////////////////////////////////////////////////////////////////////////

// lineWidth controls the world-space thickness of
// entity line quads.
const lineWidth = 8.0

// appendLineQuad builds a billboarded quad (2 triangles,
// 6 vertices) from a line segment. The quad always faces
// the camera so lines remain visible from any angle.
func appendLineQuad(buf []byte, camEye math.Vector3,
	x1, y1, z1, x2, y2, z2 float64,
	r, g, b, a float32,
) []byte {

	// Line midpoint to camera direction
	midX := (x1 + x2) * 0.5
	midY := (y1 + y2) * 0.5
	midZ := (z1 + z2) * 0.5
	toEyeX := camEye.X - midX
	toEyeY := camEye.Y - midY
	toEyeZ := camEye.Z - midZ

	// Line direction
	ldx := x2 - x1
	ldy := y2 - y1
	ldz := z2 - z1
	lineLen := sysMath.Sqrt(ldx*ldx + ldy*ldy + ldz*ldz)
	if lineLen < 1e-6 {
		return buf
	}

	// Perpendicular = cross(lineDir, toEye), then normalize.
	// This makes the quad always face the camera.
	px := ldy*toEyeZ - ldz*toEyeY
	py := ldz*toEyeX - ldx*toEyeZ
	pz := ldx*toEyeY - ldy*toEyeX
	pLen := sysMath.Sqrt(px*px + py*py + pz*pz)
	if pLen < 1e-6 {
		return buf
	}

	half := lineWidth / 2.0 / pLen
	px *= half
	py *= half
	pz *= half

	// Four corners of the billboard quad
	ax, ay, az := x1-px, y1-py, z1-pz
	bx, by, bz := x1+px, y1+py, z1+pz
	cx, cy, cz := x2-px, y2-py, z2-pz
	ex, ey, ez := x2+px, y2+py, z2+pz

	// Two triangles: (a, b, c) and (c, b, e)
	buf = appendVertex(buf, ax, ay, az, r, g, b, a)
	buf = appendVertex(buf, bx, by, bz, r, g, b, a)
	buf = appendVertex(buf, cx, cy, cz, r, g, b, a)
	buf = appendVertex(buf, cx, cy, cz, r, g, b, a)
	buf = appendVertex(buf, bx, by, bz, r, g, b, a)
	buf = appendVertex(buf, ex, ey, ez, r, g, b, a)

	return buf
}

////////////////////////////////////////////////////////////////////////////////

func appendVertex(buf []byte,
	x, y, z float64,
	r, g, b, a float32,
) []byte {

	var v [28]byte
	binary.LittleEndian.PutUint32(v[0:], sysMath.Float32bits(float32(x)))
	binary.LittleEndian.PutUint32(v[4:], sysMath.Float32bits(float32(y)))
	binary.LittleEndian.PutUint32(v[8:], sysMath.Float32bits(float32(z)))
	binary.LittleEndian.PutUint32(v[12:], sysMath.Float32bits(r))
	binary.LittleEndian.PutUint32(v[16:], sysMath.Float32bits(g))
	binary.LittleEndian.PutUint32(v[20:], sysMath.Float32bits(b))
	binary.LittleEndian.PutUint32(v[24:], sysMath.Float32bits(a))
	return append(buf, v[:]...)
}
