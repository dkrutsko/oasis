//go:build viewer3d

package overlay

import "github.com/dkrutsko/oasis/errors"

////////////////////////////////////////////////////////////////////////////////

// Viewer is a stub when building with the viewer3d tag.
// Ebiten requires CGO which is incompatible with GoGPU's
// pure Go FFI layer.
type Viewer struct{}

////////////////////////////////////////////////////////////////////////////////

func NewViewer() *Viewer { return &Viewer{} }

////////////////////////////////////////////////////////////////////////////////

// Run always fails so that `--viewer` on a viewer3d build
// reports an error instead of exiting.
func (v *Viewer) Run() error {

	return errors.New(
		"overlay viewer is not available in viewer3d builds",
	)
}

////////////////////////////////////////////////////////////////////////////////

func (v *Viewer) Close() {}
