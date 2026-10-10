package geometry

import (
	"fmt"
	sysMath "math"

	"github.com/dkrutsko/oasis/math"
)

//----------------------------------------------------------------------------//
// Constants                                                                  //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// EllipticalCapsuleZero represents an elliptical capsule with zero dimensions
// at the origin.
var EllipticalCapsuleZero = EllipticalCapsule{
	math.Vector3{X: 0, Y: 0, Z: 0},
	math.Vector3{X: 0, Y: 0, Z: 0}, 0, 0, 0,
}

//----------------------------------------------------------------------------//
// Types                                                                      //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// EllipticalCapsule represents a capsule whose cross-section perpendicular to
// the axis is an ellipse rather than a circle. `RadiusA` and `RadiusB` are the
// semi-axes of the ellipse in the two perpendicular directions derived from the
// capsule axis. `Rotation` (radians) rotates the cross-section around the
// capsule axis so the ellipse can be aligned with an entity's facing direction.
type EllipticalCapsule struct {
	Start    math.Vector3
	End      math.Vector3
	RadiusA  float64
	RadiusB  float64
	Rotation float64
}

//----------------------------------------------------------------------------//
// Methods                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// String returns the string representation of the elliptical capsule.
func (c EllipticalCapsule) String() string {

	return fmt.Sprintf(
		"[%.2f, %.2f, %.2f, %.2f, %.2f, %.2f, %.2f, %.2f, %.2f]",
		c.Start.X, c.Start.Y, c.Start.Z,
		c.End.X, c.End.Y, c.End.Z,
		c.RadiusA, c.RadiusB, c.Rotation)
}

////////////////////////////////////////////////////////////////////////////////

// GetCenter returns the midpoint of the capsule axis.
func (c EllipticalCapsule) GetCenter() math.Vector3 {

	return c.Start.Add(c.End).MulScalar(0.5)
}

////////////////////////////////////////////////////////////////////////////////

// GetHeight returns the distance between the two endpoints.
func (c EllipticalCapsule) GetHeight() float64 {

	return c.Start.Distance(c.End)
}

////////////////////////////////////////////////////////////////////////////////

// GetBasis returns the perpendicular unit vectors for the elliptical
// cross-section, with `Rotation` applied. These are the directions along which
// `RadiusA` and `RadiusB` are measured.
func (c EllipticalCapsule) GetBasis() (math.Vector3, math.Vector3) {

	axis := c.End.Sub(c.Start)
	axisLen := axis.Length()
	if axisLen < 1e-10 {
		return math.Vector3{X: 1}, math.Vector3{Y: 1}
	}

	return ellipticalCapsuleBasis(axis.MulScalar(1.0/axisLen), c.Rotation)
}

//----------------------------------------------------------------------------//
// Static                                                                     //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// ellipticalCapsuleBasis builds the perpendicular basis used by both the
// intersection and wireframe code. When `rotation` is non-zero the basis is
// rotated around the capsule axis by that angle (radians). The two returned
// vectors are unit length.
func ellipticalCapsuleBasis(axisNorm math.Vector3, rotation float64) (math.Vector3, math.Vector3) {

	ref := math.Vector3{X: 0, Y: 0, Z: 1}
	if sysMath.Abs(axisNorm.Z) > 0.9 {
		ref = math.Vector3{X: 1, Y: 0, Z: 0}
	}

	u := axisNorm.Cross(ref).Normalize()
	v := axisNorm.Cross(u).Normalize()

	if rotation != 0 {
		cosR := sysMath.Cos(rotation)
		sinR := sysMath.Sin(rotation)
		ru := u.MulScalar(cosR).Add(v.MulScalar(sinR))
		rv := u.MulScalar(-sinR).Add(v.MulScalar(cosR))
		u = ru
		v = rv
	}

	return u, v
}
