//go:build darwin || linux

package overlay

import (
	"context"
	"image"
	"image/color"
	sysMath "math"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/fogleman/gg"

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

	overlayFps = 60

	// How often to check if the shared memory segment still
	// exists. Moonlight unlinks it when the stream ends.
	shmCheckInterval = 2 * time.Second
)

////////////////////////////////////////////////////////////////////////////////

// Overlay renders game entity positions into a shared memory
// region for consumption by Moonlight or the debug viewer.
type Overlay struct {
	game *game.Game
}

////////////////////////////////////////////////////////////////////////////////

// NewOverlay creates an overlay ready to start its render loop.
// The shared memory connection is established lazily during the
// render loop once Moonlight creates the segment.
func NewOverlay(g *game.Game) *Overlay {

	return &Overlay{game: g}
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

	var (
		shm       *SharedMemory
		dc        *gg.Context
		width     int
		height    int
		lastCheck time.Time
	)

	defer func() {
		if shm != nil {
			shm.Close()
		}
	}()

	for {
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

			width = shm.Width()
			height = shm.Height()
			dc = gg.NewContext(width, height)
			lastCheck = time.Now()

			logger.Dbg("overlay attached to shared memory",
				logger.Int("width", width),
				logger.Int("height", height),
			)
		}

		//--------------------------------------------------------------------//

		// Get a start time
		start := time.Now()

		// Clear to fully transparent
		dc.SetColor(color.NRGBA{0, 0, 0, 0})
		dc.Clear()

		// Get game state snapshots
		camera := o.game.GetCameraState()
		action := o.game.GetActionState()

		// Only render when camera and entity data are valid
		if camera.Result == game.CameraResultSuccess &&
			action.Result == game.ActionResultSuccess &&
			action.Player != nil {

			o.renderEntities(dc, camera, action, width, height)
		}

		// Copy pixels to shared memory and flip
		pix := dc.Image().(*image.RGBA).Pix
		copy(shm.WriteBuffer(), pix)
		shm.Flip()

		//--------------------------------------------------------------------//

		// Periodically check if the segment has been unlinked.
		// Moonlight calls shm_unlink when the stream ends. The
		// existing mmap stays valid but the data goes stale.
		if time.Since(lastCheck) > shmCheckInterval {
			lastCheck = time.Now()

			if shm.IsUnlinked() {
				logger.Dbg("overlay segment unlinked, waiting for reconnect")
				shm.Close()
				shm = nil
				continue
			}
		}

		//--------------------------------------------------------------------//

		// Rate limit to target FPS
		elapsed := time.Since(start)
		rem := (time.Second / overlayFps) - elapsed
		if rem > 0 {
			select {
			case <-ctx.Done():
				return
			case <-time.After(rem):
			}
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

func (o *Overlay) renderEntities(
	dc *gg.Context,
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

		// Project head and feet to screen for sizing
		headScreen, headVisible := math.ProjectToScreenMvp(
			entity.Head, viewport, mvp,
		)
		feetScreen, feetVisible := math.ProjectToScreenMvp(
			entity.Origin, viewport, mvp,
		)

		if !headVisible {
			continue
		}

		// Offset from game viewport to overlay canvas
		headScreen.X += vpX
		headScreen.Y += vpY
		feetScreen.X += vpX
		feetScreen.Y += vpY

		sx := headScreen.X
		sy := headScreen.Y

		// Determine if enemy or teammate
		isEnemy := entity.Team != localTeam

		if isEnemy {
			drewBox := false

			// Use head-to-feet distance for scaling when both visible
			if feetVisible {
				boxH := feetScreen.Y - headScreen.Y
				if boxH > 4 {
					boxW := boxH * 0.6

					// Bounding box
					dc.SetColor(color.NRGBA{255, 50, 50, 220})
					dc.SetLineWidth(1.5)
					dc.DrawRectangle(
						sx-boxW/2, sy,
						boxW, boxH,
					)
					dc.Stroke()

					// Health bar on left side
					barX := sx - boxW/2 - 4
					pct := float64(entity.Health) / 100.0

					// Background
					dc.SetColor(color.NRGBA{0, 0, 0, 160})
					dc.DrawRectangle(barX, sy, 2, boxH)
					dc.Fill()

					// Fill from bottom up
					r := uint8(255 * (1 - pct))
					g := uint8(255 * pct)
					dc.SetColor(color.NRGBA{r, g, 0, 220})
					dc.DrawRectangle(barX, sy+boxH*(1-pct), 2, boxH*pct)
					dc.Fill()

					drewBox = true
				}
			}

			if !drewBox {
				// Fallback for distant enemies: simple dot
				dc.SetColor(color.NRGBA{255, 50, 50, 220})
				dc.DrawCircle(sx, sy, 4)
				dc.Fill()
			}

			// View direction indicator
			o.renderViewDir(dc, entity, sx, sy, mvp, viewport, vpX, vpY)

		} else {
			// Teammate indicator
			dc.SetColor(color.NRGBA{50, 200, 50, 150})
			dc.DrawCircle(sx, sy, 3)
			dc.Fill()
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

func (o *Overlay) renderViewDir(
	dc *gg.Context,
	entity *game.ActionEntity,
	sx, sy float64,
	mvp math.Matrix4,
	viewport math.Size,
	vpX, vpY float64,
) {

	// Convert eye angles to a forward direction vector
	yawRad := entity.Angles.Y * sysMath.Pi / 180.0
	pitchRad := entity.Angles.X * sysMath.Pi / 180.0

	dirX := sysMath.Cos(pitchRad) * sysMath.Cos(yawRad)
	dirY := sysMath.Cos(pitchRad) * sysMath.Sin(yawRad)
	dirZ := -sysMath.Sin(pitchRad)

	// Project a point ahead in the look direction
	lookPoint := math.Vector3{
		X: entity.Head.X + dirX*150,
		Y: entity.Head.Y + dirY*150,
		Z: entity.Head.Z + dirZ*150,
	}

	lookScreen, lookVisible := math.ProjectToScreenMvp(
		lookPoint, viewport, mvp,
	)

	if !lookVisible {
		return
	}

	// Offset from game viewport to overlay canvas
	lookScreen.X += vpX
	lookScreen.Y += vpY

	dc.SetColor(color.NRGBA{255, 100, 100, 140})
	dc.SetLineWidth(1.0)
	dc.DrawLine(sx, sy, lookScreen.X, lookScreen.Y)
	dc.Stroke()
}
