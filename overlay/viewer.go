//go:build darwin || linux

package overlay

import (
	"image/color"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/hajimehoshi/ebiten/v2"
)

////////////////////////////////////////////////////////////////////////////////

const viewerScale = 0.5

////////////////////////////////////////////////////////////////////////////////

// Viewer reads frames from shared memory and displays them in a
// window using ebiten. It automatically connects and reconnects
// to the shared memory segment as it becomes available.
type Viewer struct {
	shm       *SharedMemory
	img       *ebiten.Image
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

// Run starts the ebiten event loop and blocks until the window is
// closed.
func (v *Viewer) Run() error {

	w := int(float64(overlayWidth) * viewerScale)
	h := int(float64(overlayHeight) * viewerScale)

	ebiten.SetWindowSize(w, h)
	ebiten.SetWindowTitle("Oasis Viewer")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	return ebiten.RunGame(v)
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

// Update implements `ebiten.Game`.
func (v *Viewer) Update() error {

	// Try to connect if not attached
	if v.shm == nil {
		shm, err := ShmOpen()
		if err != nil {
			return nil
		}

		v.shm = shm
		v.img = ebiten.NewImage(shm.Width(), shm.Height())
		v.lastDirty = 0
		v.staleTime = time.Time{}
		return nil
	}

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
			v.Close()
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

// Draw implements `ebiten.Game`.
func (v *Viewer) Draw(screen *ebiten.Image) {

	screen.Fill(color.RGBA{30, 30, 30, 255})

	if v.shm == nil {
		return
	}

	v.img.WritePixels(v.shm.ReadBuffer())
	screen.DrawImage(v.img, nil)
}

////////////////////////////////////////////////////////////////////////////////

// Layout implements `ebiten.Game`.
func (v *Viewer) Layout(outsideWidth, outsideHeight int) (int, int) {

	return overlayWidth, overlayHeight
}
