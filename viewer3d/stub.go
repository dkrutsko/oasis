//go:build !viewer3d

package viewer3d

import (
	"context"

	"github.com/dkrutsko/oasis/errors"
	"github.com/dkrutsko/oasis/game"
	"github.com/dkrutsko/oasis/input"
)

////////////////////////////////////////////////////////////////////////////////

// Viewer3d is a stub when building without the viewer3d
// tag. Build with -tags viewer3d and CGO_ENABLED=0 to
// enable the GoGPU-based 3D viewer.
type Viewer3d struct{}

////////////////////////////////////////////////////////////////////////////////

func NewViewer3d(_ *game.Game, _ *input.Keys) *Viewer3d { return &Viewer3d{} }

////////////////////////////////////////////////////////////////////////////////

// Run always fails so that `--viewer3d` on a build without
// the viewer3d tag reports an error instead of exiting.
func (v *Viewer3d) Run(_ context.Context) error {

	return errors.New(
		"3d viewer is not available in this build",
		errors.String("required_tags", "viewer3d nofakecgo"),
	)
}

////////////////////////////////////////////////////////////////////////////////

func (v *Viewer3d) Close() {}
