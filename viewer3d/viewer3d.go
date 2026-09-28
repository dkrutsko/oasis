package viewer3d

import (
	"context"
	"time"

	"github.com/Zyko0/go-sdl3/bin/binsdl"
	"github.com/Zyko0/go-sdl3/sdl"

	"github.com/dkrutsko/oasis/errors"
	"github.com/dkrutsko/oasis/game"
	"github.com/dkrutsko/oasis/input"
	"github.com/dkrutsko/oasis/logger"
)

////////////////////////////////////////////////////////////////////////////////

// Scale from SDL wheel deltas to the camera's scroll units.
// SDL divides macOS trackpad deltas by 10 and reports the
// wheel moving away from the user as positive, the opposite
// of what the camera zooms in on.
const scrollScale = -10.0

////////////////////////////////////////////////////////////////////////////////

// Viewer3d renders the current map collision geometry and
// real-time entity positions in a GPU-accelerated window
// using SDL3 (Metal or Vulkan, loaded with purego).
type Viewer3d struct {
	game *game.Game
	keys *input.Keys

	camera    OrbitCamera
	renderer  Renderer
	loadedMap string
	lastFrame time.Time
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
// It must be called from the main goroutine, which macOS
// requires for window operations.
func (v *Viewer3d) Run(ctx context.Context) error {

	//----------------------------------------------------------------------------//

	// SDL is embedded in the binary and loaded from a temporary
	// directory, so it does not have to be installed. Loading
	// exits the process if the library cannot be written.
	defer binsdl.Load().Unload()

	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		return errors.New(
			"failed to initialize sdl",
			errors.Error("error", err),
		)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("Oasis 3D Viewer", 1280, 720,
		sdl.WINDOW_RESIZABLE|sdl.WINDOW_HIGH_PIXEL_DENSITY,
	)
	if err != nil {
		return errors.New(
			"failed to create window",
			errors.Error("error", err),
		)
	}
	defer window.Destroy()

	//----------------------------------------------------------------------------//

	// The shaders are compiled to MSL for Metal and SPIR-V for
	// Vulkan, so SDL picks one of those backends
	device, err := sdl.CreateGPUDevice(sdl.GPU_SHADERFORMAT_MSL|sdl.GPU_SHADERFORMAT_SPIRV, false, "")
	if err != nil {
		return errors.New(
			"failed to create gpu device",
			errors.Error("error", err),
		)
	}
	defer device.Destroy()

	if err := device.ClaimWindow(window); err != nil {
		return errors.New(
			"failed to claim window for gpu device",
			errors.Error("error", err),
		)
	}
	defer device.ReleaseWindow(window)

	err = v.renderer.Init(device, window)
	defer v.renderer.Destroy()

	if err != nil {
		return errors.New(
			"failed to init 3d renderer",
			errors.Error("error", err),
		)
	}

	logger.Info("3d viewer started",
		logger.String("backend", device.Driver()),
	)

	//----------------------------------------------------------------------------//

	for ctx.Err() == nil {

		// Closing the window ends the viewer
		var event sdl.Event
		for sdl.PollEvent(&event) {
			switch event.Type {
			case sdl.EVENT_QUIT, sdl.EVENT_WINDOW_CLOSE_REQUESTED:
				return nil

			case sdl.EVENT_MOUSE_WHEEL:
				v.camera.HandleScroll(float64(event.MouseWheelEvent().Y) * scrollScale)

			case sdl.EVENT_MOUSE_BUTTON_DOWN:
				// Detect left-click for double-click auto mode toggle
				if event.MouseButtonEvent().Button == uint8(sdl.BUTTON_LEFT) {
					v.camera.HandleLeftClick()
				}
			}
		}

		// A minimized window has nothing to draw into, so the
		// loop only keeps handling events
		if window.Flags()&sdl.WINDOW_MINIMIZED != 0 {
			sdl.Delay(100)
			continue
		}

		v.drawFrame(window)
	}

	//----------------------------------------------------------------------------//

	return nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// Close is a no-op. Resources are released when Run returns.
func (v *Viewer3d) Close() {}

////////////////////////////////////////////////////////////////////////////////

func (v *Viewer3d) drawFrame(window *sdl.Window) {

	//----------------------------------------------------------------------------//

	// Update viewport on every frame so the projection
	// matrix stays correct after window resizes.
	width, height, err := window.SizeInPixels()
	if err == nil {
		v.camera.SetViewport(int(width), int(height))
	}

	// Measure the frame time so free mode moves at the
	// same speed regardless of the frame rate
	now := time.Now()
	dt := now.Sub(v.lastFrame).Seconds()
	if dt > freeMaxFrameTime {
		dt = freeMaxFrameTime
	}
	v.lastFrame = now

	//----------------------------------------------------------------------------//

	// Camera input
	buttons, mx, my := sdl.GetMouseState()
	keyboard := sdl.GetKeyboardState()

	leftDown := buttons&sdl.ButtonMask(sdl.BUTTON_LEFT) != 0
	rightDown := buttons&sdl.ButtonMask(sdl.BUTTON_RIGHT) != 0
	middleDown := buttons&sdl.ButtonMask(sdl.BUTTON_MIDDLE) != 0

	if v.camera.auto {
		// Auto mode: scroll to zoom, left-drag to rotate
		if v.keys != nil {
			if delta := v.keys.GetScrollDelta(); delta != 0 {
				v.camera.HandleScroll(float64(-delta) * 5.0)
			}
		}

		v.camera.UpdateAuto(mx, my, leftDown)

		// WASD or mouse drag exits auto mode
		if keyboard[sdl.SCANCODE_W] || keyboard[sdl.SCANCODE_A] ||
			keyboard[sdl.SCANCODE_S] || keyboard[sdl.SCANCODE_D] ||
			rightDown || middleDown {
			v.camera.auto = false
		}
	} else {
		// Free mode: WASD to fly, left-drag to look,
		// double-click to return to auto mode
		v.camera.UpdateFree(dt,
			mx, my, leftDown || rightDown || middleDown,
			keyboard[sdl.SCANCODE_W],
			keyboard[sdl.SCANCODE_S],
			keyboard[sdl.SCANCODE_A],
			keyboard[sdl.SCANCODE_D],
			keyboard[sdl.SCANCODE_SPACE],
			keyboard[sdl.SCANCODE_LCTRL] || keyboard[sdl.SCANCODE_RCTRL],
			keyboard[sdl.SCANCODE_LSHIFT] || keyboard[sdl.SCANCODE_RSHIFT],
		)
	}

	//----------------------------------------------------------------------------//

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

	//----------------------------------------------------------------------------//

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
	if err := v.renderer.Draw(mvp); err != nil {
		logger.Warn("3d viewer draw error",
			logger.Error("error", err),
		)
	}

	//----------------------------------------------------------------------------//
}
