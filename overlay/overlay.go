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
	overlayWidth  = 1920
	overlayHeight = 1080
	overlayFps    = 60
)

////////////////////////////////////////////////////////////////////////////////

// Overlay renders game entity positions into a shared memory
// region for consumption by Moonlight or the debug viewer.
type Overlay struct {
	shm  *SharedMemory
	game *game.Game
}

////////////////////////////////////////////////////////////////////////////////

// NewOverlay creates a shared memory segment and returns an
// overlay ready to start its render loop.
func NewOverlay(g *game.Game) (*Overlay, error) {

	shm, err := ShmCreate(overlayWidth, overlayHeight)
	if err != nil {
		return nil, err
	}

	return &Overlay{shm: shm, game: g}, nil
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

// Close releases the shared memory segment.
func (o *Overlay) Close() error {

	if o.shm != nil {
		o.shm.Close()
		o.shm = nil
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

func (o *Overlay) renderLoop(ctx context.Context) {

	dc := gg.NewContext(overlayWidth, overlayHeight)

	for {
		// Check for shutdown
		if ctx.Err() != nil {
			return
		}

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

			o.renderEntities(dc, camera, action)
		}

		// Copy pixels to shared memory and flip
		pix := dc.Image().(*image.RGBA).Pix
		copy(o.shm.WriteBuffer(), pix)
		o.shm.Flip()

		// Rate limit to target FPS
		elapsed := time.Since(start)
		rem := (time.Second / overlayFps) - elapsed
		if rem > 0 {
			time.Sleep(rem)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

func (o *Overlay) renderEntities(
	dc *gg.Context,
	camera *game.CameraState,
	action *game.ActionState,
) {

	// CS2 view matrix is row-major (M * v convention) but
	// ProjectToScreenMvp uses v * M convention. Transpose
	// to align the two.
	mvp := camera.View.Transpose()
	localTeam := action.Player.Team

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
		viewport := math.Size{W: overlayWidth, H: overlayHeight}
		headScreen, headVisible := math.ProjectToScreenMvp(
			entity.Head, viewport, mvp,
		)
		feetScreen, feetVisible := math.ProjectToScreenMvp(
			entity.Origin, viewport, mvp,
		)

		if !headVisible {
			continue
		}

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
			o.renderViewDir(dc, entity, sx, sy, mvp)

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
		lookPoint, math.Size{W: overlayWidth, H: overlayHeight}, mvp,
	)

	if !lookVisible {
		return
	}

	dc.SetColor(color.NRGBA{255, 100, 100, 140})
	dc.SetLineWidth(1.0)
	dc.DrawLine(sx, sy, lookScreen.X, lookScreen.Y)
	dc.Stroke()
}
