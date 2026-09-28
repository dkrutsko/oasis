package viewer3d

import (
	"encoding/binary"
	sysMath "math"
	"unsafe"

	"github.com/Zyko0/go-sdl3/sdl"

	"github.com/dkrutsko/oasis/errors"
	"github.com/dkrutsko/oasis/game"
	"github.com/dkrutsko/oasis/geometry"
	"github.com/dkrutsko/oasis/maps"
	"github.com/dkrutsko/oasis/math"
)

////////////////////////////////////////////////////////////////////////////////

// Maximum entity vertex buffer size in bytes. Each line is
// a camera-facing quad of 6 vertices at 28 bytes (168 bytes)
// so this holds 390 quads. 64 entities use 192 of them and
// the aim ray gets the rest.
const entityBufSize = 65536

// Maximum HUD vertex buffer size in bytes.
const textBufSize = 8192

////////////////////////////////////////////////////////////////////////////////

// Renderer manages all SDL GPU resources and draw calls
// for the 3D viewer.
type Renderer struct {
	device *sdl.GPUDevice
	window *sdl.Window

	mapPipe    *sdl.GPUGraphicsPipeline
	entityPipe *sdl.GPUGraphicsPipeline

	mapBuf      *sdl.GPUBuffer
	mapVertices uint32

	entityBuf      *sdl.GPUBuffer
	entityData     []byte
	entityVertices uint32

	textBuf      *sdl.GPUBuffer
	textData     []byte
	textVertices uint32

	// Stages the entity and HUD vertices, which are uploaded
	// at the start of every frame
	transferBuf *sdl.GPUTransferBuffer

	lastTeam int32
}

////////////////////////////////////////////////////////////////////////////////

func (r *Renderer) Init(device *sdl.GPUDevice, window *sdl.Window) error {

	//----------------------------------------------------------------------------//

	r.device = device
	r.window = window

	format := device.SwapchainTextureFormat(window)

	//----------------------------------------------------------------------------//

	// Map pipeline: solid triangles with alpha blending
	var err error
	r.mapPipe, err = createPipeline(device, mapShaderWGSL, format, 12, []sdl.GPUVertexAttribute{
		{Location: 0, Format: sdl.GPU_VERTEXELEMENTFORMAT_FLOAT3, Offset: 0},
	})
	if err != nil {
		return err
	}

	// Entity pipeline: colored triangles with alpha blending
	r.entityPipe, err = createPipeline(device, entityShaderWGSL, format, 28, []sdl.GPUVertexAttribute{
		{Location: 0, Format: sdl.GPU_VERTEXELEMENTFORMAT_FLOAT3, Offset: 0},
		{Location: 1, Format: sdl.GPU_VERTEXELEMENTFORMAT_FLOAT4, Offset: 12},
	})
	if err != nil {
		return err
	}

	//----------------------------------------------------------------------------//

	// Pre-allocate the entity and HUD vertex buffers
	r.entityBuf, err = device.CreateBuffer(&sdl.GPUBufferCreateInfo{
		Usage: sdl.GPU_BUFFERUSAGE_VERTEX,
		Size:  entityBufSize,
	})
	if err != nil {
		return errors.New(
			"failed to create entity buffer",
			errors.Error("error", err),
		)
	}

	r.textBuf, err = device.CreateBuffer(&sdl.GPUBufferCreateInfo{
		Usage: sdl.GPU_BUFFERUSAGE_VERTEX,
		Size:  textBufSize,
	})
	if err != nil {
		return errors.New(
			"failed to create hud buffer",
			errors.Error("error", err),
		)
	}

	r.transferBuf, err = device.CreateTransferBuffer(&sdl.GPUTransferBufferCreateInfo{
		Usage: sdl.GPU_TRANSFERBUFFERUSAGE_UPLOAD,
		Size:  entityBufSize + textBufSize,
	})
	if err != nil {
		return errors.New(
			"failed to create transfer buffer",
			errors.Error("error", err),
		)
	}

	r.entityData = make([]byte, 0, entityBufSize)
	r.textData = make([]byte, 0, textBufSize)

	//----------------------------------------------------------------------------//

	return nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// LoadMap uploads the map collision geometry to the GPU.
func (r *Renderer) LoadMap(m *maps.Map) error {

	//----------------------------------------------------------------------------//

	triangles := m.GetTriangles()
	if len(triangles) == 0 {
		return nil
	}

	r.UnloadMap()

	size := uint32(len(triangles) * 36)

	buf, err := r.device.CreateBuffer(&sdl.GPUBufferCreateInfo{
		Usage: sdl.GPU_BUFFERUSAGE_VERTEX,
		Size:  size,
	})
	if err != nil {
		return errors.New(
			"failed to create map buffer",
			errors.Int("size", int(size)),
			errors.Error("error", err),
		)
	}

	//----------------------------------------------------------------------------//

	// The positions are written straight into the transfer
	// buffer, so the map is not copied again in Go memory.
	// SDL frees the transfer buffer once the upload is done.
	transfer, err := r.device.CreateTransferBuffer(&sdl.GPUTransferBufferCreateInfo{
		Usage: sdl.GPU_TRANSFERBUFFERUSAGE_UPLOAD,
		Size:  size,
	})
	if err != nil {
		r.device.ReleaseBuffer(buf)
		return errors.New(
			"failed to create transfer buffer",
			errors.Int("size", int(size)),
			errors.Error("error", err),
		)
	}
	defer r.device.ReleaseTransferBuffer(transfer)

	ptr, err := r.device.MapTransferBuffer(transfer, false)
	if err != nil {
		r.device.ReleaseBuffer(buf)
		return errors.New(
			"failed to map transfer buffer",
			errors.Error("error", err),
		)
	}

	data := unsafe.Slice((*byte)(unsafe.Pointer(ptr)), size)
	for i := range triangles {
		tri := &triangles[i]
		off := i * 36
		putVec3(data[off:], tri.V0)
		putVec3(data[off+12:], tri.V1)
		putVec3(data[off+24:], tri.V2)
	}

	r.device.UnmapTransferBuffer(transfer)

	//----------------------------------------------------------------------------//

	cmd, err := r.device.AcquireCommandBuffer()
	if err != nil {
		r.device.ReleaseBuffer(buf)
		return errors.New(
			"failed to acquire command buffer",
			errors.Error("error", err),
		)
	}

	pass := cmd.BeginCopyPass()
	pass.UploadToGPUBuffer(
		&sdl.GPUTransferBufferLocation{TransferBuffer: transfer},
		&sdl.GPUBufferRegion{Buffer: buf, Size: size},
		false,
	)
	pass.End()

	if err := cmd.Submit(); err != nil {
		r.device.ReleaseBuffer(buf)
		return errors.New(
			"failed to upload map",
			errors.Error("error", err),
		)
	}

	r.mapBuf = buf
	r.mapVertices = uint32(len(triangles) * 3)

	//----------------------------------------------------------------------------//

	return nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// UnloadMap clears the map geometry from the GPU.
func (r *Renderer) UnloadMap() {

	if r.mapBuf != nil {
		r.device.ReleaseBuffer(r.mapBuf)
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
		r.entityData = r.entityData[:0]
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

	// Build vertex data: [x y z r g b a] per vertex. The
	// buffer is reused and uploaded when the frame is drawn.
	buf := r.entityData[:0]

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
		// the local player. Within 64 units stays opaque,
		// then fades to 0.3 at 128 units (one floor).
		if isEnemy && hasPlayer {
			heightDiff := sysMath.Abs(entity.Origin.Z - playerZ)
			t := heightDiff / 128.0
			if t < 0.5 {
				// Same floor: fully opaque
			} else if t < 1.0 {
				// Sharp falloff from 1.0 to 0.3
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

	// Drop whole quads that do not fit in the vertex buffer.
	// The dashed aim ray is appended last so only its far end
	// is lost.
	if len(buf) > entityBufSize {
		buf = buf[:entityBufSize-entityBufSize%(6*28)]
	}

	r.entityData = buf
	r.entityVertices = uint32(len(buf) / 28)

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// UpdateHud builds the HUD indicator. Shows a green
// circle in the top-right when auto mode is active.
// Viewport width/height are needed to correct for aspect
// ratio so the circle isn't stretched into an oval.
func (r *Renderer) UpdateHud(autoMode bool, vpW, vpH int) {

	if !autoMode {
		r.textData = r.textData[:0]
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
	buf := r.textData[:0]

	for i := 0; i < segments; i++ {
		a1 := float64(i) * 2 * sysMath.Pi / float64(segments)
		a2 := float64(i+1) * 2 * sysMath.Pi / float64(segments)

		buf = appendVertex(buf, cx, cy, 0, cr, cg, cb, ca)
		buf = appendVertex(buf, cx+radius*aspect*sysMath.Cos(a1), cy+radius*sysMath.Sin(a1), 0, cr, cg, cb, ca)
		buf = appendVertex(buf, cx+radius*aspect*sysMath.Cos(a2), cy+radius*sysMath.Sin(a2), 0, cr, cg, cb, ca)
	}

	r.textData = buf
	r.textVertices = uint32(len(buf) / 28)
}

////////////////////////////////////////////////////////////////////////////////

// Draw uploads the entity and HUD vertices of this frame,
// then renders the map, entities and HUD using the given
// MVP. Nothing is rendered while the window has no frame to
// draw into, such as when it is minimized.
func (r *Renderer) Draw(mvp math.Matrix4) error {

	//----------------------------------------------------------------------------//

	cmd, err := r.device.AcquireCommandBuffer()
	if err != nil {
		return errors.New(
			"failed to acquire command buffer",
			errors.Error("error", err),
		)
	}

	//----------------------------------------------------------------------------//

	// Cycling gives the transfer and vertex buffers a fresh
	// copy when earlier frames still read the last one
	if len(r.entityData) > 0 || len(r.textData) > 0 {
		ptr, err := r.device.MapTransferBuffer(r.transferBuf, true)
		if err != nil {
			cmd.Cancel()
			return errors.New(
				"failed to map transfer buffer",
				errors.Error("error", err),
			)
		}

		staging := unsafe.Slice((*byte)(unsafe.Pointer(ptr)), entityBufSize+textBufSize)
		copy(staging, r.entityData)
		copy(staging[entityBufSize:], r.textData)

		r.device.UnmapTransferBuffer(r.transferBuf)

		pass := cmd.BeginCopyPass()

		if len(r.entityData) > 0 {
			pass.UploadToGPUBuffer(
				&sdl.GPUTransferBufferLocation{TransferBuffer: r.transferBuf},
				&sdl.GPUBufferRegion{Buffer: r.entityBuf, Size: uint32(len(r.entityData))},
				true,
			)
		}

		if len(r.textData) > 0 {
			pass.UploadToGPUBuffer(
				&sdl.GPUTransferBufferLocation{TransferBuffer: r.transferBuf, Offset: entityBufSize},
				&sdl.GPUBufferRegion{Buffer: r.textBuf, Size: uint32(len(r.textData))},
				true,
			)
		}

		pass.End()
	}

	//----------------------------------------------------------------------------//

	// Waits for the window to take another frame, which paces
	// the viewer to the display refresh rate
	swapchain, err := cmd.WaitAndAcquireGPUSwapchainTexture(r.window)
	if err != nil {
		cmd.Cancel()
		return errors.New(
			"failed to acquire swapchain texture",
			errors.Error("error", err),
		)
	}

	// The command buffer is still submitted for the uploads
	if swapchain.Texture == nil {
		return cmd.Submit()
	}

	pass := cmd.BeginRenderPass([]sdl.GPUColorTargetInfo{{
		Texture:    swapchain.Texture,
		ClearColor: sdl.FColor{R: 0.08, G: 0.08, B: 0.10, A: 1},
		LoadOp:     sdl.GPU_LOADOP_CLEAR,
		StoreOp:    sdl.GPU_STOREOP_STORE,
	}}, nil)

	var uniforms [64]byte
	putMatrix(uniforms[:], mvp)
	cmd.PushVertexUniformData(0, uniforms[:])

	//----------------------------------------------------------------------------//

	// Draw map (transparent solid triangles)
	if r.mapVertices > 0 && r.mapBuf != nil {
		pass.BindGraphicsPipeline(r.mapPipe)
		pass.BindVertexBuffers([]sdl.GPUBufferBinding{{Buffer: r.mapBuf}})
		pass.DrawPrimitives(r.mapVertices, 1, 0, 0)
	}

	// Draw entities (alpha blended line quads)
	if r.entityVertices > 0 {
		pass.BindGraphicsPipeline(r.entityPipe)
		pass.BindVertexBuffers([]sdl.GPUBufferBinding{{Buffer: r.entityBuf}})
		pass.DrawPrimitives(r.entityVertices, 1, 0, 0)
	}

	// Draw HUD text (screen-space, identity MVP)
	if r.textVertices > 0 {
		putMatrix(uniforms[:], math.Matrix4Identity)
		cmd.PushVertexUniformData(0, uniforms[:])

		pass.BindGraphicsPipeline(r.entityPipe)
		pass.BindVertexBuffers([]sdl.GPUBufferBinding{{Buffer: r.textBuf}})
		pass.DrawPrimitives(r.textVertices, 1, 0, 0)
	}

	//----------------------------------------------------------------------------//

	pass.End()

	if err := cmd.Submit(); err != nil {
		return errors.New(
			"failed to submit command buffer",
			errors.Error("error", err),
		)
	}

	return nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// Destroy releases all GPU resources.
func (r *Renderer) Destroy() {

	if r.device == nil {
		return
	}

	r.UnloadMap()

	for _, pipe := range []*sdl.GPUGraphicsPipeline{r.mapPipe, r.entityPipe} {
		if pipe != nil {
			r.device.ReleaseGraphicsPipeline(pipe)
		}
	}

	for _, buf := range []*sdl.GPUBuffer{r.entityBuf, r.textBuf} {
		if buf != nil {
			r.device.ReleaseBuffer(buf)
		}
	}

	if r.transferBuf != nil {
		r.device.ReleaseTransferBuffer(r.transferBuf)
	}
}

////////////////////////////////////////////////////////////////////////////////

// createPipeline builds an alpha blended triangle pipeline
// from a WGSL shader and the layout of its vertex buffer.
func createPipeline(
	device *sdl.GPUDevice,
	source string,
	format sdl.GPUTextureFormat,
	pitch uint32,
	attributes []sdl.GPUVertexAttribute,
) (*sdl.GPUGraphicsPipeline, error) {

	//----------------------------------------------------------------------------//

	vertex, fragment, err := compileShaders(device, source)
	if err != nil {
		return nil, err
	}
	defer device.ReleaseShader(vertex)
	defer device.ReleaseShader(fragment)

	//----------------------------------------------------------------------------//

	pipe, err := device.CreateGraphicsPipeline(&sdl.GPUGraphicsPipelineCreateInfo{
		VertexShader:   vertex,
		FragmentShader: fragment,
		VertexInputState: sdl.GPUVertexInputState{
			VertexBufferDescriptions: []sdl.GPUVertexBufferDescription{{
				Slot:      0,
				Pitch:     pitch,
				InputRate: sdl.GPU_VERTEXINPUTRATE_VERTEX,
			}},
			VertexAttributes: attributes,
		},
		PrimitiveType: sdl.GPU_PRIMITIVETYPE_TRIANGLELIST,
		TargetInfo: sdl.GPUGraphicsPipelineTargetInfo{
			ColorTargetDescriptions: []sdl.GPUColorTargetDescription{{
				Format: format,
				BlendState: sdl.GPUColorTargetBlendState{
					EnableBlend:         true,
					SrcColorBlendfactor: sdl.GPU_BLENDFACTOR_SRC_ALPHA,
					DstColorBlendfactor: sdl.GPU_BLENDFACTOR_ONE_MINUS_SRC_ALPHA,
					ColorBlendOp:        sdl.GPU_BLENDOP_ADD,
					SrcAlphaBlendfactor: sdl.GPU_BLENDFACTOR_ONE,
					DstAlphaBlendfactor: sdl.GPU_BLENDFACTOR_ONE_MINUS_SRC_ALPHA,
					AlphaBlendOp:        sdl.GPU_BLENDOP_ADD,
				},
			}},
		},
	})
	if err != nil {
		return nil, errors.New(
			"failed to create graphics pipeline",
			errors.Error("error", err),
		)
	}

	//----------------------------------------------------------------------------//

	return pipe, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

func putVec3(dst []byte, v math.Vector3) {

	binary.LittleEndian.PutUint32(dst[0:], sysMath.Float32bits(float32(v.X)))
	binary.LittleEndian.PutUint32(dst[4:], sysMath.Float32bits(float32(v.Y)))
	binary.LittleEndian.PutUint32(dst[8:], sysMath.Float32bits(float32(v.Z)))
}

////////////////////////////////////////////////////////////////////////////////

func putMatrix(dst []byte, m math.Matrix4) {

	for i, v := range m.ToSlice32() {
		binary.LittleEndian.PutUint32(dst[i*4:], sysMath.Float32bits(v))
	}
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
