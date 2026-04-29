//go:build viewer3d

package viewer3d

import (
	"context"

	"github.com/gogpu/gogpu"
	ginput "github.com/gogpu/gogpu/input"
	"github.com/gogpu/gpucontext"

	"github.com/dkrutsko/oasis/game"
	"github.com/dkrutsko/oasis/input"
	"github.com/dkrutsko/oasis/logger"
)

////////////////////////////////////////////////////////////////////////////////

// Viewer3d renders the current map collision geometry and
// real-time entity positions in a GPU-accelerated window
// using GoGPU (WebGPU, pure Go, zero CGO).
type Viewer3d struct {
	game *game.Game
	keys *input.Keys

	camera    OrbitCamera
	renderer  Renderer
	loadedMap string
	ready     bool
}

////////////////////////////////////////////////////////////////////////////////

// NewViewer3d creates a viewer ready to run.
func NewViewer3d(g *game.Game, keys *input.Keys) *Viewer3d {

	return &Viewer3d{
		game:   g,
		keys:   keys,
		camera: NewOrbitCamera(),
	}
}

////////////////////////////////////////////////////////////////////////////////

// Run creates the window and enters the render loop. This
// blocks until the window is closed or the context cancels.
func (v *Viewer3d) Run(ctx context.Context) error {

	//----------------------------------------------------------------------------//

	app := gogpu.NewApp(gogpu.DefaultConfig().
		WithTitle("Oasis 3D Viewer").
		WithSize(1280, 720).
		WithContinuousRender(true))

	// Use the event API for scroll since the polling API
	// resets scroll deltas before OnDraw runs.
	events := app.EventSource()
	events.OnScroll(func(_, dy float64) {
		v.camera.HandleScroll(dy)
	})

	// Detect left-click for double-click auto mode toggle
	events.OnMousePress(func(button gpucontext.MouseButton, _, _ float64) {
		if button == gpucontext.MouseButtonLeft {
			v.camera.HandleLeftClick()
		}
	})

	//----------------------------------------------------------------------------//

	app.OnDraw(func(dc *gogpu.Context) {

		// Check for context cancellation
		if ctx.Err() != nil {
			return
		}

		// Lazy init: wait for DeviceProvider to become available
		provider := app.DeviceProvider()
		if provider == nil {
			return
		}

		sv := dc.SurfaceView()
		if sv == nil {
			return
		}

		if !v.ready {
			if err := v.renderer.Init(provider.Device(), provider.SurfaceFormat()); err != nil {
				logger.Err("failed to init 3d renderer",
					logger.Error("error", err),
				)
				return
			}
			v.ready = true

			w := dc.FramebufferWidth()
			h := dc.FramebufferHeight()
			v.camera.SetViewport(w, h)

			logger.Info("3d viewer started",
				logger.String("backend", dc.Backend()),
			)
		}

		//--------------------------------------------------------------------//

		// Camera input
		inp := app.Input()
		mouse := inp.Mouse()
		kb := inp.Keyboard()

		if v.camera.auto {
			// Auto mode: scroll to zoom, left-drag to rotate
			mx, my := mouse.Position()
			leftDown := mouse.Pressed(ginput.MouseButtonLeft)

			if v.keys != nil {
				if delta := v.keys.GetScrollDelta(); delta != 0 {
					v.camera.HandleScroll(float64(-delta) * 5.0)
				}
			}

			v.camera.UpdateAuto(mx, my, leftDown)

			// WASD or mouse drag exits auto mode
			if kb.Pressed(ginput.KeyW) || kb.Pressed(ginput.KeyA) ||
				kb.Pressed(ginput.KeyS) || kb.Pressed(ginput.KeyD) ||
				mouse.Pressed(ginput.MouseButtonRight) ||
				mouse.Pressed(ginput.MouseButtonMiddle) {
				v.camera.auto = false
			}
		} else {
			// Free mode: WASD to fly, left-drag to look,
			// double-click to return to auto mode
			mx, my := mouse.Position()
			anyMouse := mouse.Pressed(ginput.MouseButtonLeft) ||
				mouse.Pressed(ginput.MouseButtonRight) ||
				mouse.Pressed(ginput.MouseButtonMiddle)

			dt := 1.0 / 60.0
			v.camera.UpdateFree(dt,
				mx, my, anyMouse,
				kb.Pressed(ginput.KeyW),
				kb.Pressed(ginput.KeyS),
				kb.Pressed(ginput.KeyA),
				kb.Pressed(ginput.KeyD),
				kb.Pressed(ginput.KeySpace),
				kb.Pressed(ginput.KeyControlLeft) || kb.Pressed(ginput.KeyControlRight),
				kb.Pressed(ginput.KeyShiftLeft) || kb.Pressed(ginput.KeyShiftRight),
			)
		}

		//--------------------------------------------------------------------//

		// Check for map changes
		currentMap := v.game.GetCurrentMap()
		mapName := ""
		if currentMap != nil {
			mapName = currentMap.Name
		}

		if mapName != v.loadedMap {
			v.renderer.UnloadMap()

			if currentMap != nil {
				if err := v.renderer.LoadMap(currentMap); err != nil {
					logger.Warn("failed to load map in 3d viewer",
						logger.String("map", mapName),
						logger.Error("error", err),
					)
				}

				// Frame the camera on the new map
				if currentMap.Root != nil {
					box := currentMap.Root.Box
					v.camera.FrameMap(box.GetMin(), box.GetMax())
				}
			}

			v.loadedMap = mapName
		}

		//--------------------------------------------------------------------//

		// Update entity positions and aim ray
		action := v.game.GetActionState()
		if action != nil && action.Result == game.ActionResultSuccess {

			// In auto mode, follow the local player
			if action.Player != nil {
				headPos := action.Player.Head
				if action.Player.Bones.Valid && !action.Player.Bones.Pos[game.BoneHead].IsZero() {
					headPos = action.Player.Bones.Pos[game.BoneHead]
				}
				v.camera.FollowPlayer(headPos, action.Player.Angles.Y, action.Player.Angles.X)
			}

			v.renderer.UpdateEntities(action, v.camera.GetEyePosition(), currentMap)
		}

		// Draw
		v.renderer.UpdateHud(v.camera.auto, v.camera.viewport.W, v.camera.viewport.H)
		mvp := v.camera.GetViewProjection()
		if err := v.renderer.Draw(sv, mvp); err != nil {
			logger.Warn("3d viewer draw error",
				logger.Error("error", err),
			)
		}
	})

	//----------------------------------------------------------------------------//

	app.OnClose(func() {
		v.renderer.Destroy()
	})

	//----------------------------------------------------------------------------//

	return app.Run()

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// Close is a no-op. Resources are cleaned up in OnClose.
func (v *Viewer3d) Close() {}
