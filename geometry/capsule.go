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

// CapsuleZero represents a capsule with zero dimensions at the origin.
var CapsuleZero = Capsule{math.Vector3{X: 0, Y: 0, Z: 0}, math.Vector3{X: 0, Y: 0, Z: 0}, 0}

//----------------------------------------------------------------------------//
// Types                                                                      //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// Capsule represents a capsule defined by two endpoints forming the central
// axis and a radius. The capsule is the set of all points within the given
// radius of the line segment between the endpoints.
type Capsule struct {
	Start  math.Vector3
	End    math.Vector3
	Radius float64
}

//----------------------------------------------------------------------------//
// Methods                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// String returns the string representation of the capsule.
func (c Capsule) String() string {

	return fmt.Sprintf(
		"[%.2f, %.2f, %.2f, %.2f, %.2f, %.2f, %.2f]",
		c.Start.X, c.Start.Y, c.Start.Z,
		c.End.X, c.End.Y, c.End.Z, c.Radius)
}

////////////////////////////////////////////////////////////////////////////////

// GetCenter returns the midpoint of the capsule axis.
func (c Capsule) GetCenter() math.Vector3 {

	return c.Start.Add(c.End).MulScalar(0.5)
}

////////////////////////////////////////////////////////////////////////////////

// GetHeight returns the distance between the two endpoints.
func (c Capsule) GetHeight() float64 {

	return c.Start.Distance(c.End)
}

////////////////////////////////////////////////////////////////////////////////

// Contains returns whether the point is inside or on the capsule.
func (c Capsule) Contains(point math.Vector3) bool {

	ab := c.End.Sub(c.Start)
	abLenSq := ab.LengthSq()

	// Degenerate capsule reduces to a sphere
	if abLenSq < 1e-20 {
		return point.Sub(c.Start).LengthSq() <= c.Radius*c.Radius
	}

	t := math.Clamp(point.Sub(c.Start).Dot(ab)/abLenSq, 0, 1)
	closest := c.Start.Add(ab.MulScalar(t))

	return point.Sub(closest).LengthSq() <= c.Radius*c.Radius
}

////////////////////////////////////////////////////////////////////////////////

// ClosestPoint returns the closest point on the capsule surface to the given
// point. If the point is at the center of the axis, an arbitrary surface
// point is returned.
func (c Capsule) ClosestPoint(point math.Vector3) math.Vector3 {

	ab := c.End.Sub(c.Start)
	abLenSq := ab.LengthSq()

	// Degenerate capsule reduces to a sphere
	if abLenSq < 1e-20 {
		dir := point.Sub(c.Start)
		length := dir.Length()
		if length < 1e-10 {
			return math.Vector3{X: c.Start.X + c.Radius, Y: c.Start.Y, Z: c.Start.Z}
		}
		return c.Start.Add(dir.MulScalar(c.Radius / length))
	}

	// Find closest point on the axis segment
	t := math.Clamp(point.Sub(c.Start).Dot(ab)/abLenSq, 0, 1)
	axisPoint := c.Start.Add(ab.MulScalar(t))

	// Project outward to the capsule surface
	dir := point.Sub(axisPoint)
	length := dir.Length()

	if length < 1e-10 {
		// Point is on the axis - pick a perpendicular direction
		perp := math.Vector3{X: 1, Y: 0, Z: 0}
		abNorm := ab.Normalize()
		if sysMath.Abs(abNorm.X) > 0.9 {
			perp = math.Vector3{X: 0, Y: 1, Z: 0}
		}
		dir = abNorm.Cross(perp).Normalize()
		return axisPoint.Add(dir.MulScalar(c.Radius))
	}

	return axisPoint.Add(dir.MulScalar(c.Radius / length))
}

////////////////////////////////////////////////////////////////////////////////

// Distance returns the signed distance from the point to the capsule surface.
// Negative values indicate the point is inside the capsule.
func (c Capsule) Distance(point math.Vector3) float64 {

	ab := c.End.Sub(c.Start)
	abLenSq := ab.LengthSq()

	// Degenerate capsule reduces to a sphere
	if abLenSq < 1e-20 {
		return point.Sub(c.Start).Length() - c.Radius
	}

	t := math.Clamp(point.Sub(c.Start).Dot(ab)/abLenSq, 0, 1)
	closest := c.Start.Add(ab.MulScalar(t))

	return point.Sub(closest).Length() - c.Radius
}

