//go:build darwin || linux

package overlay

import (
	"context"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/Zyko0/go-sdl3/bin/binsdl"
	"github.com/Zyko0/go-sdl3/sdl"

	"github.com/dkrutsko/oasis/errors"
)

////////////////////////////////////////////////////////////////////////////////

const viewerScale = 0.5

////////////////////////////////////////////////////////////////////////////////

// Viewer reads frames from shared memory and displays them in a
// window using SDL. It automatically connects and reconnects to
// the shared memory segment as it becomes available.
type Viewer struct {
	shm       *SharedMemory
	texture   *sdl.Texture
	lastDirty uint32
	staleTime time.Time
}

////////////////////////////////////////////////////////////////////////////////

// NewViewer prepares a viewer for display. The shared memory
// connection is established lazily during the render loop.
func NewViewer() *Viewer {

	return &Viewer{}
}

////////////////////////////////////////////////////////////////////////////////

// Run opens the window and displays the overlay until the window
// is closed or the context cancels. It must be called from the
// main goroutine, which macOS requires for window operations.
func (v *Viewer) Run(ctx context.Context) error {

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

	w := int(float64(overlayWidth) * viewerScale)
	h := int(float64(overlayHeight) * viewerScale)

	window, err := sdl.CreateWindow("Oasis Viewer", w, h, sdl.WINDOW_RESIZABLE)
	if err != nil {
		return errors.New(
			"failed to create window",
			errors.Error("error", err),
		)
	}
	defer window.Destroy()

	renderer, err := window.CreateRenderer("")
	if err != nil {
		return errors.New(
			"failed to create renderer",
			errors.Error("error", err),
		)
	}
	defer renderer.Destroy()

	// Present at the display refresh rate instead of spinning.
	// This is best effort, not every driver supports it.
	renderer.SetVSync(1)

	// The texture belongs to the renderer, so it is released
	// before the renderer is destroyed
	defer v.disconnect()

	//----------------------------------------------------------------------------//

	for ctx.Err() == nil {
		var event sdl.Event
		for sdl.PollEvent(&event) {
			if event.Type == sdl.EVENT_QUIT || event.Type == sdl.EVENT_WINDOW_CLOSE_REQUESTED {
				return nil
			}
		}

		v.update(renderer)
		v.draw(renderer)
	}

	//----------------------------------------------------------------------------//

	return nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// Close releases the shared memory handle.
func (v *Viewer) Close() {

	if v.shm != nil {
		v.shm.Close()
		v.shm = nil
	}
}

////////////////////////////////////////////////////////////////////////////////

func (v *Viewer) update(renderer *sdl.Renderer) {

	//----------------------------------------------------------------------------//

	// Try to connect if not attached
	if v.shm == nil {
		shm, err := ShmOpen(true)
		if err != nil {
			return
		}

		texture, err := renderer.CreateTexture(
			sdl.PIXELFORMAT_RGBA32,
			sdl.TEXTUREACCESS_STREAMING,
			shm.Width(), shm.Height(),
		)
		if err != nil {
			shm.Close()
			return
		}

		// The overlay is drawn with straight alpha over the
		// background, scaled to fit the window
		texture.SetBlendMode(sdl.BLENDMODE_BLEND)
		texture.SetScaleMode(sdl.SCALEMODE_LINEAR)
		renderer.SetLogicalPresentation(
			int32(shm.Width()), int32(shm.Height()),
			sdl.LOGICAL_PRESENTATION_LETTERBOX,
		)

		v.shm = shm
		v.texture = texture
		v.lastDirty = 0
		v.staleTime = time.Time{}
		return
	}

	//----------------------------------------------------------------------------//

	// Detect if the producer stopped writing by watching the
	// dirty flag. If it hasn't changed for a full second, the
	// producer is likely gone. Disconnect so we can reconnect
	// to a fresh segment.
	data := v.shm.seg.GetData()
	dirty := atomic.LoadUint32((*uint32)(unsafe.Pointer(&data[20])))

	if dirty != v.lastDirty {
		v.lastDirty = dirty
		v.staleTime = time.Time{}
	} else {
		if v.staleTime.IsZero() {
			v.staleTime = time.Now()
		} else if time.Since(v.staleTime) > time.Second {
			v.disconnect()
		}
	}

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

func (v *Viewer) draw(renderer *sdl.Renderer) {

	renderer.SetDrawColor(30, 30, 30, 255)
	renderer.Clear()

	if v.shm != nil {
		pixels := v.shm.ReadBuffer()
		if err := v.texture.Update(nil, pixels, int32(v.shm.Width()*4)); err == nil {
			renderer.RenderTexture(v.texture, nil, nil)
		}
	}

	renderer.Present()
}

////////////////////////////////////////////////////////////////////////////////

func (v *Viewer) disconnect() {

	if v.texture != nil {
		v.texture.Destroy()
		v.texture = nil
	}

	v.Close()
}
