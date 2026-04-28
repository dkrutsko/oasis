package geometry

import (
	"fmt"

	"github.com/dkrutsko/oasis/math"
)

//----------------------------------------------------------------------------//
// Constants                                                                  //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// SphereZero represents a sphere with zero radius at the origin.
var SphereZero = Sphere{math.Vector3{X: 0, Y: 0, Z: 0}, 0}

//----------------------------------------------------------------------------//
// Types                                                                      //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// Sphere represents a sphere defined by a center point and radius.
type Sphere struct {
	Center math.Vector3
	Radius float64
}

//----------------------------------------------------------------------------//
// Methods                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// String returns the string representation of the sphere.
func (s Sphere) String() string {

	return fmt.Sprintf("[%.2f, %.2f, %.2f, %.2f]",
		s.Center.X, s.Center.Y, s.Center.Z, s.Radius)
}

////////////////////////////////////////////////////////////////////////////////

// Contains returns whether the point is inside or on the sphere.
func (s Sphere) Contains(point math.Vector3) bool {

	return point.Sub(s.Center).LengthSq() <= s.Radius*s.Radius
}

////////////////////////////////////////////////////////////////////////////////

// ClosestPoint returns the closest point on the sphere surface to the given
// point. If the point is at the center, an arbitrary surface point is returned.
func (s Sphere) ClosestPoint(point math.Vector3) math.Vector3 {

	dir := point.Sub(s.Center)
	length := dir.Length()

	if length < 1e-10 {
		return math.Vector3{X: s.Center.X + s.Radius, Y: s.Center.Y, Z: s.Center.Z}
	}

	return s.Center.Add(dir.MulScalar(s.Radius / length))
}

////////////////////////////////////////////////////////////////////////////////

// Distance returns the signed distance from the point to the sphere surface.
// Negative values indicate the point is inside the sphere.
func (s Sphere) Distance(point math.Vector3) float64 {

	return point.Sub(s.Center).Length() - s.Radius
}

////////////////////////////////////////////////////////////////////////////////

// IntersectsSphere returns whether this sphere intersects another sphere.
func (s Sphere) IntersectsSphere(other Sphere) bool {

	combined := s.Radius + other.Radius
	return s.Center.Sub(other.Center).LengthSq() <= combined*combined
}

