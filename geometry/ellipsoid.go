package geometry

import (
	"fmt"

	"github.com/dkrutsko/oasis/math"
)

//----------------------------------------------------------------------------//
// Constants                                                                  //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// EllipsoidZero represents an ellipsoid with zero radii at the origin.
var EllipsoidZero = Ellipsoid{math.Vector3{X: 0, Y: 0, Z: 0}, 0, 0, 0}

//----------------------------------------------------------------------------//
// Types                                                                      //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// Ellipsoid represents an axis-aligned ellipsoid defined by a center point and
// three semi-axis radii along the X, Y, and Z axes.
type Ellipsoid struct {
	Center  math.Vector3
	RadiusX float64
	RadiusY float64
	RadiusZ float64
}

//----------------------------------------------------------------------------//
// Methods                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// String returns the string representation of the ellipsoid.
func (e Ellipsoid) String() string {

	return fmt.Sprintf("[%.2f, %.2f, %.2f, %.2f, %.2f, %.2f]",
		e.Center.X, e.Center.Y, e.Center.Z,
		e.RadiusX, e.RadiusY, e.RadiusZ)
}

////////////////////////////////////////////////////////////////////////////////

// Contains returns whether the point is inside or on the ellipsoid.
func (e Ellipsoid) Contains(point math.Vector3) bool {

	dx := (point.X - e.Center.X) / e.RadiusX
	dy := (point.Y - e.Center.Y) / e.RadiusY
	dz := (point.Z - e.Center.Z) / e.RadiusZ
	return dx*dx+dy*dy+dz*dz <= 1.0
}
