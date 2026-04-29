//go:build darwin || linux

package overlay

import (
	"context"
	"image/color"
	sysMath "math"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/dkrutsko/oasis/game"
	"github.com/dkrutsko/oasis/logger"
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
	game    *game.Game
	trigger *game.Trigger
}

////////////////////////////////////////////////////////////////////////////////

// NewOverlay creates an overlay ready to start its render loop.
// The shared memory connection is established lazily during the
// render loop once Moonlight creates the segment.
func NewOverlay(g *game.Game, trigger *game.Trigger) *Overlay {

	return &Overlay{game: g, trigger: trigger}
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
			action.Result == game.ActionResultSuccess &&
			action.Player != nil {

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
	localTeam := action.Player.Team

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

		// Determine if enemy or teammate
		isEnemy := entity.Team != localTeam

		if isEnemy {
			o.renderSkeleton(dc, entity, mvp, viewport, vpX, vpY)
		} else {
			// Teammate indicator
			headScreen, headVisible := math.ProjectToScreenMvp(
				entity.Head, viewport, mvp,
			)

			if headVisible {
				headScreen.X += vpX
				headScreen.Y += vpY
				dc.DrawCircle(headScreen.X, headScreen.Y, 3, color.NRGBA{50, 200, 50, 150})
			}
		}
	}

}

////////////////////////////////////////////////////////////////////////////////

func (o *Overlay) renderSkeleton(
	dc *Canvas,
	entity *game.ActionEntity,
	mvp math.Matrix4,
	viewport math.Size,
	vpX, vpY float64,
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

	// Pick color based on health
	var c color.NRGBA
	if entity.Health > 70 {
		c = color.NRGBA{50, 220, 50, 255}
	} else if entity.Health > 30 {
		c = color.NRGBA{255, 160, 30, 255}
	} else {
		c = color.NRGBA{255, 50, 50, 255}
	}

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
