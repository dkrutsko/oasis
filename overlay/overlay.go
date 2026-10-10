package overlay

import (
	"context"
	"image/color"
	sysMath "math"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/dkrutsko/oasis/game"
	"github.com/dkrutsko/oasis/geometry"
	"github.com/dkrutsko/oasis/logger"
	"github.com/dkrutsko/oasis/maps"
	"github.com/dkrutsko/oasis/math"
)

////////////////////////////////////////////////////////////////////////////////

const (
	// Default dimensions used by the viewer when the shared memory
	// segment is not yet available.
	overlayWidth  = 1920
	overlayHeight = 1080

	// How often to check if the shared memory segment still
	// exists. Moonlight unlinks it when the stream ends.
	shmCheckInterval = 2 * time.Second
)

////////////////////////////////////////////////////////////////////////////////

// Overlay renders game entity positions into a shared memory
// region for consumption by Moonlight or the debug viewer.
type Overlay struct {
	game         *game.Game
	trigger      *game.Trigger
	triggerDebug bool

	// Last known local team for rendering when the player
	// pawn is not valid (e.g. dead). Only the render loop
	// writes this.
	lastTeam int32
}

////////////////////////////////////////////////////////////////////////////////

// NewOverlay creates an overlay ready to start its render loop.
// The shared memory connection is established lazily during the
// render loop once Moonlight creates the segment. When
// `triggerDebug` is true, triggerbot wireframes are rendered
// regardless of the trigger's enabled state.
func NewOverlay(g *game.Game, trigger *game.Trigger, triggerDebug bool) *Overlay {

	return &Overlay{game: g, trigger: trigger, triggerDebug: triggerDebug}
}

////////////////////////////////////////////////////////////////////////////////

// Start launches the render loop as a goroutine managed by the
// given errgroup.
func (o *Overlay) Start(group *errgroup.Group, ctx context.Context) {

	group.Go(func() error {
		logger.Dbg("starting overlay renderer")
		o.renderLoop(ctx)
		logger.Dbg("stopping overlay renderer")
		return nil
	})
}

////////////////////////////////////////////////////////////////////////////////

// Close is a no-op. The shared memory segment is owned by
// Moonlight and cleaned up inside the render loop.
func (o *Overlay) Close() error {

	return nil
}

////////////////////////////////////////////////////////////////////////////////

func (o *Overlay) renderLoop(ctx context.Context) {

	hb := o.game.GetHealthMonitor().Register("overlay")

	var (
		shm       *SharedMemory
		canvases  [shmBufferCount]Canvas
		width     int
		height    int
		lastCheck time.Time
	)

	defer func() {
		if shm != nil {
			// Clear all three buffers and publish so Moonlight
			// picks up a blank frame on shutdown.
			for i := uint32(0); i < shmBufferCount; i++ {
				clear(shm.GetBuffer(i))
			}
			shm.Publish(0)
			shm.Close()
		}
	}()

	for {
		hb.Beat()

		// Check for shutdown
		if ctx.Err() != nil {
			return
		}

		//--------------------------------------------------------------------//

		// Try to attach to shared memory if not connected.
		// Moonlight creates the segment when a stream starts.
		if shm == nil {
			var err error
			shm, err = ShmOpen(false)
			if err != nil {
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Second):
				}
				continue
			}

			newW := shm.Width()
			newH := shm.Height()

			// Create a gg context for each buffer so we can
			// draw directly into shared memory. This avoids
			// the per-frame copy from a separate canvas.
			if newW != width || newH != height {
				width = newW
				height = newH
				stride := width * 4

				for i := uint32(0); i < shmBufferCount; i++ {
					canvases[i] = Canvas{
						Pix:    shm.GetBuffer(i),
						Width:  width,
						Height: height,
						Stride: stride,
					}
				}
			}

			lastCheck = time.Now()

			logger.Dbg("overlay attached to shared memory",
				logger.Int("width", width),
				logger.Int("height", height),
			)
		}

		//--------------------------------------------------------------------//

		hb.Beat()

		// Pick a buffer that is not being read by Moonlight
		// and is not the latest completed frame.
		writeIdx := shm.GetWriteIndex()
		dc := &canvases[writeIdx]

		dc.Clear()
		hb.Beat()

		// Get game state snapshots
		camera := o.game.GetCameraState()
		action := o.game.GetActionState()

		// Only render when camera and entity data are valid
		if camera.Result == game.CameraResultSuccess &&
			action.Result == game.ActionResultSuccess {

			o.renderEntities(dc, camera, action, width, height)
		}
		hb.Beat()

		// Crosshair at screen center. Red when the triggerbot
		// is enabled, green otherwise. Draw outline first,
		// then fill on top.
		cx := float64(width) / 2
		cy := float64(height) / 2

		// Black outline
		dc.DrawLine(cx-20, cy, cx+20, cy, 4, color.NRGBA{0, 0, 0, 255})
		dc.DrawLine(cx, cy-20, cx, cy+20, 4, color.NRGBA{0, 0, 0, 255})

		// Inner fill
		inner := color.NRGBA{0, 255, 0, 255}
		if o.trigger != nil && o.trigger.Enabled.Load() {
			inner = color.NRGBA{255, 0, 0, 255}
		}

		dc.DrawLine(cx-20, cy, cx+20, cy, 2, inner)
		dc.DrawLine(cx, cy-20, cx, cy+20, 2, inner)

		// Publish the completed frame so Moonlight can
		// pick it up on its next read cycle.
		shm.Publish(writeIdx)

		//--------------------------------------------------------------------//

		// Periodically check if the segment has been unlinked.
		// Moonlight calls shm_unlink when the stream ends. The
		// existing mmap stays valid but the data goes stale.
		if time.Since(lastCheck) > shmCheckInterval {
			lastCheck = time.Now()
			hb.Beat()

			unlinkStart := time.Now()
			unlinked := shm.IsUnlinked()
			if dur := time.Since(unlinkStart); dur > 50*time.Millisecond {
				logger.Warn("overlay unlink check slow",
					logger.Duration("dur", dur),
				)
			}

			if unlinked {
				logger.Dbg("overlay segment unlinked, waiting for reconnect")
				shm.Close()
				shm = nil
				continue
			}
		}

		//--------------------------------------------------------------------//

		// Wait for action or camera to produce new data
		// before rendering the next frame.
		select {
		case <-ctx.Done():
			return
		case <-o.game.GetUpdated():
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

func (o *Overlay) renderEntities(
	dc *Canvas,
	camera *game.CameraState,
	action *game.ActionState,
	width, height int,
) {

	// CS2 view matrix is row-major (M * v convention) but
	// ProjectToScreenMvp uses v * M convention. Transpose
	// to align the two.
	mvp := camera.View.Transpose()

	var localTeam int32
	if action.Player != nil {
		localTeam = action.Player.Team
		o.lastTeam = localTeam
	} else {
		localTeam = o.lastTeam
	}

	// Extract the game's viewport aspect ratio from the view
	// matrix. For a combined view-projection matrix, the 3D
	// length of row 0 gives |P[0][0]| (horizontal projection
	// scale) and row 1 gives |P[1][1]| (vertical). Their
	// ratio is the viewport width/height aspect ratio. This
	// lets us detect black bars when the game runs at a
	// different aspect ratio than the stream.
	view := camera.View
	xScale := sysMath.Sqrt(view.M11*view.M11 + view.M12*view.M12 + view.M13*view.M13)
	yScale := sysMath.Sqrt(view.M21*view.M21 + view.M22*view.M22 + view.M23*view.M23)

	streamW := float64(width)
	streamH := float64(height)

	var vpX, vpY float64
	var vpW, vpH float64

	if xScale > 0 && yScale > 0 {
		gameAspect := yScale / xScale
		streamAspect := streamW / streamH

		if gameAspect < streamAspect {
			// Pillarboxing (black bars on sides)
			vpH = streamH
			vpW = vpH * gameAspect
			vpX = (streamW - vpW) / 2
		} else {
			// Letterboxing (black bars top/bottom) or matching
			vpW = streamW
			vpH = streamW / gameAspect
			vpY = (streamH - vpH) / 2
		}
	} else {
		vpW = streamW
		vpH = streamH
	}

	viewport := math.Size{W: int(vpW), H: int(vpH)}

	// In trigger-debug mode, skip normal entity rendering
	// so only the wireframes are visible
	if !o.triggerDebug {
		for i := range action.Entities {
			entity := &action.Entities[i]

			// Skip local player and invalid entities
			if !entity.Valid || entity == action.Player {
				continue
			}

			// Skip spectators
			if entity.Team <= 1 {
				continue
			}

			isEnemy := localTeam != 0 && entity.Team != localTeam
			isCarrier := game.IsObjectiveCarrier(action, entity.Index)

			if isEnemy {
				var c color.NRGBA
				if isCarrier {
					c = color.NRGBA{180, 80, 255, 255}
				} else if entity.Health > 70 {
					c = color.NRGBA{50, 220, 50, 255}
				} else if entity.Health > 30 {
					c = color.NRGBA{255, 160, 30, 255}
				} else {
					c = color.NRGBA{255, 50, 50, 255}
				}
				o.renderSkeleton(dc, entity, mvp, viewport, vpX, vpY, c)
			} else {
				headScreen, headVisible := math.ProjectToScreenMvp(
					entity.Head, viewport, mvp,
				)

				if headVisible {
					headScreen.X += vpX
					headScreen.Y += vpY

					c := color.NRGBA{50, 200, 50, 150}
					if isCarrier {
						c = color.NRGBA{180, 80, 255, 255}
					}

					dc.DrawCircle(headScreen.X, headScreen.Y, 3, c)
				}
			}
		}

		// Bomb and hostage world position markers
		purple := color.NRGBA{180, 80, 255, 255}

		if action.Bomb != nil && !action.Bomb.Origin.IsZero() &&
			(action.Bomb.Dropped || action.Bomb.Planted) {

			o.renderMarker(dc, action.Bomb.Origin, mvp, viewport, vpX, vpY, purple)
		}

		for i := range action.Hostages {
			h := &action.Hostages[i]
			if h.Rescued || h.Origin.IsZero() {
				continue
			}

			o.renderMarker(dc, h.Origin, mvp, viewport, vpX, vpY, purple)
		}
	}

	// Triggerbot debug wireframes
	m := o.game.GetCurrentMap()
	o.renderTriggerDebug(dc, action, m, localTeam, mvp, viewport, vpX, vpY)

}

////////////////////////////////////////////////////////////////////////////////

func (o *Overlay) renderSkeleton(
	dc *Canvas,
	entity *game.ActionEntity,
	mvp math.Matrix4,
	viewport math.Size,
	vpX, vpY float64,
	c color.NRGBA,
) {

	//----------------------------------------------------------------------------//

	// Genesis-style rendering: a vertical spine line from feet
	// to head height, plus a horizontal shoulder bar perpendicular
	// to the entity's facing direction.

	headPos := entity.Head
	if !headPos.IsZero() && entity.Bones.Valid {
		headPos = entity.Bones.Pos[game.BoneHead]
	}

	yawRad := entity.Angles.Y * sysMath.Pi / 180.0
	pitchRad := entity.Angles.X * sysMath.Pi / 180.0
	sin90 := sysMath.Sin(yawRad + sysMath.Pi/2)
	cos90 := sysMath.Cos(yawRad + sysMath.Pi/2)

	// Vertical spine: from (head.x, head.y, origin.z) to
	// (head.x, head.y, head.z)
	feetPoint := math.Vector3{X: headPos.X, Y: headPos.Y, Z: entity.Origin.Z}
	topPoint := headPos

	// Shoulder bar: 40 units wide, perpendicular to facing,
	// at feet height
	shoulderL := math.Vector3{
		X: headPos.X - 20*cos90,
		Y: headPos.Y - 20*sin90,
		Z: entity.Origin.Z,
	}
	shoulderR := math.Vector3{
		X: headPos.X + 20*cos90,
		Y: headPos.Y + 20*sin90,
		Z: entity.Origin.Z,
	}

	//----------------------------------------------------------------------------//

	// Project all four points
	pFeet, feetVis := math.ProjectToScreenMvp(feetPoint, viewport, mvp)
	pHead, headVis := math.ProjectToScreenMvp(topPoint, viewport, mvp)
	pShL, shLVis := math.ProjectToScreenMvp(shoulderL, viewport, mvp)
	pShR, shRVis := math.ProjectToScreenMvp(shoulderR, viewport, mvp)

	if !headVis && !feetVis {
		return
	}

	// Offset to overlay canvas
	pFeet.X += vpX
	pFeet.Y += vpY
	pHead.X += vpX
	pHead.Y += vpY
	pShL.X += vpX
	pShL.Y += vpY
	pShR.X += vpX
	pShR.Y += vpY

	//----------------------------------------------------------------------------//

	// Vertical spine line
	if feetVis && headVis {
		dc.DrawLine(pFeet.X, pFeet.Y, pHead.X, pHead.Y, 3, c)
	}

	// Shoulder bar
	if shLVis && shRVis {
		dc.DrawLine(pShL.X, pShL.Y, pShR.X, pShR.Y, 3, c)
	}

	// View direction line from head
	if headVis {
		dirX := sysMath.Cos(pitchRad) * sysMath.Cos(yawRad)
		dirY := sysMath.Cos(pitchRad) * sysMath.Sin(yawRad)
		dirZ := -sysMath.Sin(pitchRad)

		lookPoint := math.Vector3{
			X: headPos.X + dirX*10,
			Y: headPos.Y + dirY*10,
			Z: headPos.Z + dirZ*10,
		}

		pLook, lookVis := math.ProjectToScreenMvp(lookPoint, viewport, mvp)
		if lookVis {
			pLook.X += vpX
			pLook.Y += vpY
			dc.DrawLine(pHead.X, pHead.Y, pLook.X, pLook.Y, 3, c)
		}
	}

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// renderMarker draws a vertical line with an X-shaped base
// at the given world position, used for bomb and hostage
// markers.
func (o *Overlay) renderMarker(
	dc *Canvas,
	pos math.Vector3,
	mvp math.Matrix4,
	viewport math.Size,
	vpX, vpY float64,
	c color.NRGBA,
) {

	//----------------------------------------------------------------------------//

	topPoint := math.Vector3{X: pos.X, Y: pos.Y, Z: pos.Z + 30}

	// X pattern: two diagonal bars crossing at the base
	diagAL := math.Vector3{X: pos.X - 10, Y: pos.Y - 10, Z: pos.Z}
	diagAR := math.Vector3{X: pos.X + 10, Y: pos.Y + 10, Z: pos.Z}
	diagBL := math.Vector3{X: pos.X - 10, Y: pos.Y + 10, Z: pos.Z}
	diagBR := math.Vector3{X: pos.X + 10, Y: pos.Y - 10, Z: pos.Z}

	//----------------------------------------------------------------------------//

	pBase, baseVis := math.ProjectToScreenMvp(pos, viewport, mvp)
	pTop, topVis := math.ProjectToScreenMvp(topPoint, viewport, mvp)
	pAL, alVis := math.ProjectToScreenMvp(diagAL, viewport, mvp)
	pAR, arVis := math.ProjectToScreenMvp(diagAR, viewport, mvp)
	pBL, blVis := math.ProjectToScreenMvp(diagBL, viewport, mvp)
	pBR, brVis := math.ProjectToScreenMvp(diagBR, viewport, mvp)

	if !baseVis && !topVis {
		return
	}

	pBase.X += vpX
	pBase.Y += vpY
	pTop.X += vpX
	pTop.Y += vpY
	pAL.X += vpX
	pAL.Y += vpY
	pAR.X += vpX
	pAR.Y += vpY
	pBL.X += vpX
	pBL.Y += vpY
	pBR.X += vpX
	pBR.Y += vpY

	//----------------------------------------------------------------------------//

	// Vertical line
	if baseVis && topVis {
		dc.DrawLine(pBase.X, pBase.Y, pTop.X, pTop.Y, 3, c)
	}

	// Diagonal bars
	if alVis && arVis {
		dc.DrawLine(pAL.X, pAL.Y, pAR.X, pAR.Y, 3, c)
	}

	if blVis && brVis {
		dc.DrawLine(pBL.X, pBL.Y, pBR.X, pBR.Y, 3, c)
	}

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// renderTriggerDebug draws wireframe outlines of the triggerbot
// hit shapes (head ellipsoid and body capsule) for each enemy
// entity. Active when `--trigger-debug` is set or when the
// triggerbot is enabled via the activation key.
func (o *Overlay) renderTriggerDebug(
	dc *Canvas,
	action *game.ActionState,
	m *maps.Map,
	localTeam int32,
	mvp math.Matrix4,
	viewport math.Size,
	vpX, vpY float64,
) {

	//----------------------------------------------------------------------------//

	if !o.triggerDebug {
		return
	}

	//----------------------------------------------------------------------------//

	cyan := color.NRGBA{0, 255, 255, 255}
	red := color.NRGBA{255, 50, 50, 255}
	const segments = 16
	const lineWidth = 1

	headRx := float64(game.TriggerHeadRadiusX)
	headRy := float64(game.TriggerHeadRadiusY)
	headRz := float64(game.TriggerHeadRadiusZ)
	bodyRX := float64(game.TriggerBodyRadiusX)
	bodyRY := float64(game.TriggerBodyRadiusY)

	// Build the aim ray from the local player so we can
	// highlight whichever shape the crosshair intersects
	var aimRay geometry.Ray
	hasAim := false

	player := action.Player
	if player != nil {
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

		yawRad := player.Angles.Y * sysMath.Pi / 180.0
		pitchRad := player.Angles.X * sysMath.Pi / 180.0

		aimRay = geometry.Ray{
			Origin: eyePos,
			Direction: math.Vector3{
				X: sysMath.Cos(pitchRad) * sysMath.Cos(yawRad),
				Y: sysMath.Cos(pitchRad) * sysMath.Sin(yawRad),
				Z: -sysMath.Sin(pitchRad),
			},
		}
		hasAim = true
	}

	//----------------------------------------------------------------------------//

	for i := range action.Entities {
		entity := &action.Entities[i]

		if !entity.Valid || entity == action.Player || !entity.Bones.Valid {
			continue
		}

		if entity.Team == localTeam || entity.Team <= 1 {
			continue
		}

		headPos := entity.Bones.Pos[game.BoneHead]
		neckPos := entity.Bones.Pos[game.BoneNeck]
		pelvisPos := entity.Bones.Pos[game.BonePelvis]

		// Test aim ray against each shape to pick the color.
		// A geometry hit is suppressed when the map blocks the
		// line of sight, matching the real triggerbot logic.
		headHit := false
		bodyHit := false

		if hasAim {
			if !headPos.IsZero() {
				dist, hit := aimRay.IntersectEllipsoid(geometry.Ellipsoid{
					Center:  headPos,
					RadiusX: game.TriggerHeadRadiusX,
					RadiusY: game.TriggerHeadRadiusY,
					RadiusZ: game.TriggerHeadRadiusZ,
				})
				if hit {
					headHit = m == nil || m.IsVisible(aimRay.Origin, aimRay.GetPoint(dist))
				}
			}
			if !neckPos.IsZero() && !pelvisPos.IsZero() {
				entYaw := entity.Angles.Y * sysMath.Pi / 180.0
				dist, hit := aimRay.IntersectEllipticalCapsule(geometry.EllipticalCapsule{
					Start:    neckPos,
					End:      pelvisPos,
					RadiusA:  game.TriggerBodyRadiusY,
					RadiusB:  game.TriggerBodyRadiusX,
					Rotation: entYaw,
				})
				if hit {
					bodyHit = m == nil || m.IsVisible(aimRay.Origin, aimRay.GetPoint(dist))
				}
			}
		}

		// Head ellipsoid: three great ellipses along the
		// major planes (XY, XZ, YZ)
		if !headPos.IsZero() {
			axisX := math.Vector3{X: headRx}
			axisY := math.Vector3{Y: headRy}
			axisZ := math.Vector3{Z: headRz}

			headColor := cyan
			if headHit {
				headColor = red
			}

			o.renderWireEllipse(dc, headPos, axisX, axisY, segments, lineWidth, mvp, viewport, vpX, vpY, headColor)
			o.renderWireEllipse(dc, headPos, axisX, axisZ, segments, lineWidth, mvp, viewport, vpX, vpY, headColor)
			o.renderWireEllipse(dc, headPos, axisY, axisZ, segments, lineWidth, mvp, viewport, vpX, vpY, headColor)
		}

		// Limb capsules connecting adjacent bones
		limbSegs := [...]struct {
			startBone int
			endBone   int
			radius    float64
		}{
			{game.BoneLeftShoulder, game.BoneLeftElbow, game.TriggerUpperArmRadius},
			{game.BoneLeftElbow, game.BoneLeftHand, game.TriggerForearmRadius},
			{game.BoneRightShoulder, game.BoneRightElbow, game.TriggerUpperArmRadius},
			{game.BoneRightElbow, game.BoneRightHand, game.TriggerForearmRadius},
			{game.BoneLeftHip, game.BoneLeftKnee, game.TriggerThighRadius},
			{game.BoneLeftKnee, game.BoneLeftFoot, game.TriggerShinRadius},
			{game.BoneRightHip, game.BoneRightKnee, game.TriggerThighRadius},
			{game.BoneRightKnee, game.BoneRightFoot, game.TriggerShinRadius},
		}

		for _, seg := range limbSegs {
			startPos := entity.Bones.Pos[seg.startBone]
			endPos := entity.Bones.Pos[seg.endBone]
			if startPos.IsZero() || endPos.IsZero() {
				continue
			}

			limbHit := false
			if hasAim {
				dist, hit := aimRay.IntersectCapsule(geometry.Capsule{
					Start:  startPos,
					End:    endPos,
					Radius: seg.radius,
				})
				if hit {
					limbHit = m == nil || m.IsVisible(aimRay.Origin, aimRay.GetPoint(dist))
				}
			}

			limbColor := cyan
			if limbHit {
				limbColor = red
			}

			axis := endPos.Sub(startPos)
			axisLen := axis.Length()
			if axisLen < 0.01 {
				continue
			}

			axisNorm := axis.MulScalar(1.0 / axisLen)

			// Build perpendicular basis for the capsule
			ref := math.Vector3{Z: 1}
			if sysMath.Abs(axisNorm.Z) > 0.9 {
				ref = math.Vector3{X: 1}
			}

			uDir := axisNorm.Cross(ref).Normalize()
			vDir := axisNorm.Cross(uDir).Normalize()
			u := uDir.MulScalar(seg.radius)
			v := vDir.MulScalar(seg.radius)

			// Circles at both ends
			o.renderWireEllipse(dc, startPos, u, v, segments, lineWidth, mvp, viewport, vpX, vpY, limbColor)
			o.renderWireEllipse(dc, endPos, u, v, segments, lineWidth, mvp, viewport, vpX, vpY, limbColor)

			// Connecting lines at four cardinal directions
			for s := 0; s < 4; s++ {
				t := 2 * sysMath.Pi * float64(s) / 4.0
				offset := u.MulScalar(sysMath.Cos(t)).Add(v.MulScalar(sysMath.Sin(t)))
				p0 := startPos.Add(offset)
				p1 := endPos.Add(offset)

				sp0, v0 := math.ProjectToScreenMvp(p0, viewport, mvp)
				sp1, v1 := math.ProjectToScreenMvp(p1, viewport, mvp)
				if v0 && v1 {
					dc.DrawLine(sp0.X+vpX, sp0.Y+vpY, sp1.X+vpX, sp1.Y+vpY, lineWidth, limbColor)
				}
			}

			// Hemispherical caps
			capDirStart := axisNorm.Neg().MulScalar(seg.radius)
			capDirEnd := axisNorm.MulScalar(seg.radius)

			o.renderWireArc(dc, startPos, u, capDirStart, segments/2, lineWidth, mvp, viewport, vpX, vpY, limbColor)
			o.renderWireArc(dc, startPos, v, capDirStart, segments/2, lineWidth, mvp, viewport, vpX, vpY, limbColor)
			o.renderWireArc(dc, endPos, u, capDirEnd, segments/2, lineWidth, mvp, viewport, vpX, vpY, limbColor)
			o.renderWireArc(dc, endPos, v, capDirEnd, segments/2, lineWidth, mvp, viewport, vpX, vpY, limbColor)
		}

		// Body capsule: elliptical cross-section circles at
		// the cylinder ends, connecting lines, and
		// hemispherical end caps
		if !neckPos.IsZero() && !pelvisPos.IsZero() {
			entYaw := entity.Angles.Y * sysMath.Pi / 180.0
			ec := geometry.EllipticalCapsule{
				Start:    neckPos,
				End:      pelvisPos,
				RadiusA:  bodyRY,
				RadiusB:  bodyRX,
				Rotation: entYaw,
			}

			uDir, vDir := ec.GetBasis()
			u := uDir.MulScalar(bodyRY)
			v := vDir.MulScalar(bodyRX)

			axis := pelvisPos.Sub(neckPos)
			axisLen := axis.Length()
			if axisLen > 0.01 {
				axisNorm := axis.MulScalar(1.0 / axisLen)

				bodyColor := cyan
				if bodyHit {
					bodyColor = red
				}

				// Elliptical cross-section at both ends
				o.renderWireEllipse(dc, neckPos, u, v, segments, lineWidth, mvp, viewport, vpX, vpY, bodyColor)
				o.renderWireEllipse(dc, pelvisPos, u, v, segments, lineWidth, mvp, viewport, vpX, vpY, bodyColor)

				// Connecting lines at four cardinal directions
				for s := 0; s < 4; s++ {
					t := 2 * sysMath.Pi * float64(s) / 4.0
					offset := u.MulScalar(sysMath.Cos(t)).Add(v.MulScalar(sysMath.Sin(t)))
					p0 := neckPos.Add(offset)
					p1 := pelvisPos.Add(offset)

					sp0, v0 := math.ProjectToScreenMvp(p0, viewport, mvp)
					sp1, v1 := math.ProjectToScreenMvp(p1, viewport, mvp)
					if v0 && v1 {
						dc.DrawLine(sp0.X+vpX, sp0.Y+vpY, sp1.X+vpX, sp1.Y+vpY, lineWidth, bodyColor)
					}
				}

				// Hemispherical caps: semicircular arcs in the
				// u-axis and v-axis planes. The cap height
				// uses the larger radius to match the
				// intersection geometry.
				capR := sysMath.Max(bodyRX, bodyRY)
				capDirNeck := axisNorm.Neg().MulScalar(capR)
				capDirPelvis := axisNorm.MulScalar(capR)

				o.renderWireArc(dc, neckPos, u, capDirNeck, segments/2, lineWidth, mvp, viewport, vpX, vpY, bodyColor)
				o.renderWireArc(dc, neckPos, v, capDirNeck, segments/2, lineWidth, mvp, viewport, vpX, vpY, bodyColor)
				o.renderWireArc(dc, pelvisPos, u, capDirPelvis, segments/2, lineWidth, mvp, viewport, vpX, vpY, bodyColor)
				o.renderWireArc(dc, pelvisPos, v, capDirPelvis, segments/2, lineWidth, mvp, viewport, vpX, vpY, bodyColor)
			}
		}
	}

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// renderWireEllipse draws a wireframe ellipse in 3D, defined
// by a center and two axis vectors that span the ellipse plane
// and encode the radii in their lengths.
func (o *Overlay) renderWireEllipse(
	dc *Canvas,
	center math.Vector3,
	axisU, axisV math.Vector3,
	segments, lineWidth int,
	mvp math.Matrix4,
	viewport math.Size,
	vpX, vpY float64,
	c color.NRGBA,
) {

	for s := 0; s < segments; s++ {
		t0 := 2 * sysMath.Pi * float64(s) / float64(segments)
		t1 := 2 * sysMath.Pi * float64(s+1) / float64(segments)

		p0 := center.Add(axisU.MulScalar(sysMath.Cos(t0))).Add(axisV.MulScalar(sysMath.Sin(t0)))
		p1 := center.Add(axisU.MulScalar(sysMath.Cos(t1))).Add(axisV.MulScalar(sysMath.Sin(t1)))

		sp0, v0 := math.ProjectToScreenMvp(p0, viewport, mvp)
		sp1, v1 := math.ProjectToScreenMvp(p1, viewport, mvp)

		if v0 && v1 {
			dc.DrawLine(sp0.X+vpX, sp0.Y+vpY, sp1.X+vpX, sp1.Y+vpY, lineWidth, c)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

// renderWireArc draws a semicircular arc from `+equatorAxis`
// through `poleAxis` to `-equatorAxis`. Used for the
// hemispherical caps of capsule wireframes.
func (o *Overlay) renderWireArc(
	dc *Canvas,
	center math.Vector3,
	equatorAxis, poleAxis math.Vector3,
	segments, lineWidth int,
	mvp math.Matrix4,
	viewport math.Size,
	vpX, vpY float64,
	c color.NRGBA,
) {

	for s := 0; s < segments; s++ {
		t0 := sysMath.Pi * float64(s) / float64(segments)
		t1 := sysMath.Pi * float64(s+1) / float64(segments)

		p0 := center.Add(equatorAxis.MulScalar(sysMath.Cos(t0))).Add(poleAxis.MulScalar(sysMath.Sin(t0)))
		p1 := center.Add(equatorAxis.MulScalar(sysMath.Cos(t1))).Add(poleAxis.MulScalar(sysMath.Sin(t1)))

		sp0, v0 := math.ProjectToScreenMvp(p0, viewport, mvp)
		sp1, v1 := math.ProjectToScreenMvp(p1, viewport, mvp)

		if v0 && v1 {
			dc.DrawLine(sp0.X+vpX, sp0.Y+vpY, sp1.X+vpX, sp1.Y+vpY, lineWidth, c)
		}
	}
}
