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

// CylinderZero represents a cylinder with zero dimensions at the origin.
var CylinderZero = Cylinder{math.Vector3{X: 0, Y: 0, Z: 0}, math.Vector3{X: 0, Y: 0, Z: 0}, 0}

//----------------------------------------------------------------------------//
// Types                                                                      //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// Cylinder represents a capped cylinder defined by two endpoints forming the
// central axis and a radius. The cylinder includes flat caps at both ends.
type Cylinder struct {
	Start  math.Vector3
	End    math.Vector3
	Radius float64
}

//----------------------------------------------------------------------------//
// Methods                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// String returns the string representation of the cylinder.
func (c Cylinder) String() string {

	return fmt.Sprintf(
		"[%.2f, %.2f, %.2f, %.2f, %.2f, %.2f, %.2f]",
		c.Start.X, c.Start.Y, c.Start.Z,
		c.End.X, c.End.Y, c.End.Z, c.Radius)
}

////////////////////////////////////////////////////////////////////////////////

// GetCenter returns the midpoint of the cylinder axis.
func (c Cylinder) GetCenter() math.Vector3 {

	return c.Start.Add(c.End).MulScalar(0.5)
}

////////////////////////////////////////////////////////////////////////////////

// GetHeight returns the distance between the two endpoints.
func (c Cylinder) GetHeight() float64 {

	return c.Start.Distance(c.End)
}

////////////////////////////////////////////////////////////////////////////////

// Contains returns whether the point is inside or on the cylinder.
func (c Cylinder) Contains(point math.Vector3) bool {

	ab := c.End.Sub(c.Start)
	abLenSq := ab.LengthSq()

	if abLenSq < 1e-20 {
		return false
	}

	// Check axial range
	t := point.Sub(c.Start).Dot(ab) / abLenSq

	if t < 0 || t > 1 {
		return false
	}

	// Check radial distance
	axisPoint := c.Start.Add(ab.MulScalar(t))

	return point.Sub(axisPoint).LengthSq() <= c.Radius*c.Radius
}

////////////////////////////////////////////////////////////////////////////////

// ClosestPoint returns the closest point on the cylinder surface to the given
// point. The surface includes the cylindrical body and both flat caps.
func (c Cylinder) ClosestPoint(point math.Vector3) math.Vector3 {

	ab := c.End.Sub(c.Start)
	abLenSq := ab.LengthSq()

	if abLenSq < 1e-20 {
		return c.Start
	}

	baLen := sysMath.Sqrt(abLenSq)
	baNorm := ab.MulScalar(1 / baLen)

	pa := point.Sub(c.Start)
	axial := pa.Dot(baNorm)

	// Perpendicular vector from axis to point
	perpVec := pa.Sub(baNorm.MulScalar(axial))
	perpDist := perpVec.Length()

	//------------------------------------------------------------------------//

	// Beyond start cap
	if axial < 0 {
		if perpDist <= c.Radius {
			return c.Start.Add(perpVec)
		}
		return c.Start.Add(perpVec.MulScalar(c.Radius / perpDist))
	}

	// Beyond end cap
	if axial > baLen {
		if perpDist <= c.Radius {
			return c.End.Add(perpVec)
		}
		return c.End.Add(perpVec.MulScalar(c.Radius / perpDist))
	}

	//------------------------------------------------------------------------//

	// Within axial range
	axisPoint := c.Start.Add(baNorm.MulScalar(axial))

	// Handle point on the axis
	if perpDist < 1e-10 {
		perp := math.Vector3{X: 1, Y: 0, Z: 0}
		if sysMath.Abs(baNorm.X) > 0.9 {
			perp = math.Vector3{X: 0, Y: 1, Z: 0}
		}
		perpDir := baNorm.Cross(perp).Normalize()
		return axisPoint.Add(perpDir.MulScalar(c.Radius))
	}

	perpDir := perpVec.MulScalar(1 / perpDist)

	// Outside the body - project onto cylindrical surface
	if perpDist >= c.Radius {
		return axisPoint.Add(perpDir.MulScalar(c.Radius))
	}

	//------------------------------------------------------------------------//

	// Inside the cylinder - find nearest surface among body, start cap, end cap
	bodyDist := c.Radius - perpDist
	startDist := axial
	endDist := baLen - axial

	if bodyDist <= startDist && bodyDist <= endDist {
		return axisPoint.Add(perpDir.MulScalar(c.Radius))
	}

	if startDist <= endDist {
		return c.Start.Add(perpVec)
	}

	return c.End.Add(perpVec)
}

////////////////////////////////////////////////////////////////////////////////

// Distance returns the signed distance from the point to the cylinder surface.
// Negative values indicate the point is inside the cylinder.
func (c Cylinder) Distance(point math.Vector3) float64 {

	ab := c.End.Sub(c.Start)
	abLenSq := ab.LengthSq()

	if abLenSq < 1e-20 {
		return point.Sub(c.Start).Length()
	}

	baLen := sysMath.Sqrt(abLenSq)
	baNorm := ab.MulScalar(1 / baLen)
	halfHeight := baLen * 0.5

	pa := point.Sub(c.Start)
	axial := pa.Dot(baNorm)

	// Perpendicular distance from axis
	radial := pa.Sub(baNorm.MulScalar(axial)).Length()

	// 2D signed distance of a rectangle in (radial, axial) space
	dx := radial - c.Radius
	dy := sysMath.Abs(axial-halfHeight) - halfHeight

	outside := sysMath.Sqrt(
		sysMath.Max(dx, 0)*sysMath.Max(dx, 0) +
			sysMath.Max(dy, 0)*sysMath.Max(dy, 0))

	inside := sysMath.Min(sysMath.Max(dx, dy), 0)

	return outside + inside
}
