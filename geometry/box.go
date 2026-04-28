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

// BoxZero represents a box with zero extents at the origin.
var BoxZero = Box{math.Vector3{X: 0, Y: 0, Z: 0}, math.Vector3{X: 0, Y: 0, Z: 0}}

//----------------------------------------------------------------------------//
// Types                                                                      //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// Box represents an axis-aligned bounding box defined by a center point
// and half-size extents along each axis.
type Box struct {
	Center  math.Vector3
	Extents math.Vector3
}

//----------------------------------------------------------------------------//
// Methods                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// String returns the string representation of the box.
func (b Box) String() string {

	return fmt.Sprintf(
		"[%.2f, %.2f, %.2f, %.2f, %.2f, %.2f]",
		b.Center.X, b.Center.Y, b.Center.Z,
		b.Extents.X, b.Extents.Y, b.Extents.Z)
}

////////////////////////////////////////////////////////////////////////////////

// GetMin returns the minimum corner of the box.
func (b Box) GetMin() math.Vector3 {

	return b.Center.Sub(b.Extents)
}

////////////////////////////////////////////////////////////////////////////////

// GetMax returns the maximum corner of the box.
func (b Box) GetMax() math.Vector3 {

	return b.Center.Add(b.Extents)
}

////////////////////////////////////////////////////////////////////////////////

// GetSize returns the full size of the box along each axis.
func (b Box) GetSize() math.Vector3 {

	return b.Extents.MulScalar(2)
}

////////////////////////////////////////////////////////////////////////////////

// Contains returns whether the point is inside or on the box.
func (b Box) Contains(point math.Vector3) bool {

	min := b.GetMin()
	max := b.GetMax()

	return point.X >= min.X && point.X <= max.X &&
		point.Y >= min.Y && point.Y <= max.Y &&
		point.Z >= min.Z && point.Z <= max.Z
}

////////////////////////////////////////////////////////////////////////////////

// ClosestPoint returns the closest point on or in the box to the given point.
func (b Box) ClosestPoint(point math.Vector3) math.Vector3 {

	min := b.GetMin()
	max := b.GetMax()

	return math.Vector3{
		X: math.Clamp(point.X, min.X, max.X),
		Y: math.Clamp(point.Y, min.Y, max.Y),
		Z: math.Clamp(point.Z, min.Z, max.Z),
	}
}

////////////////////////////////////////////////////////////////////////////////

// Distance returns the signed distance from the point to the box surface.
// Negative values indicate the point is inside the box.
func (b Box) Distance(point math.Vector3) float64 {

	dx := sysMath.Abs(point.X-b.Center.X) - b.Extents.X
	dy := sysMath.Abs(point.Y-b.Center.Y) - b.Extents.Y
	dz := sysMath.Abs(point.Z-b.Center.Z) - b.Extents.Z

	outside := sysMath.Sqrt(
		sysMath.Max(dx, 0)*sysMath.Max(dx, 0) +
			sysMath.Max(dy, 0)*sysMath.Max(dy, 0) +
			sysMath.Max(dz, 0)*sysMath.Max(dz, 0))

	inside := sysMath.Min(sysMath.Max(dx, sysMath.Max(dy, dz)), 0)

	return outside + inside
}

////////////////////////////////////////////////////////////////////////////////

// IntersectsBox returns whether this box intersects another box.
func (b Box) IntersectsBox(other Box) bool {

	min := b.GetMin()
	max := b.GetMax()
	otherMin := other.GetMin()
	otherMax := other.GetMax()

	return min.X <= otherMax.X && max.X >= otherMin.X &&
		min.Y <= otherMax.Y && max.Y >= otherMin.Y &&
		min.Z <= otherMax.Z && max.Z >= otherMin.Z
}

////////////////////////////////////////////////////////////////////////////////

// IntersectsSphere returns whether this box intersects the given sphere.
func (b Box) IntersectsSphere(sphere Sphere) bool {

	closest := b.ClosestPoint(sphere.Center)
	return closest.Sub(sphere.Center).LengthSq() <= sphere.Radius*sphere.Radius
}

////////////////////////////////////////////////////////////////////////////////

// Encapsulate returns the box expanded to include the given point.
func (b Box) Encapsulate(point math.Vector3) Box {

	min := b.GetMin().Min(point)
	max := b.GetMax().Max(point)

	return BoxFromMinMax(min, max)
}

////////////////////////////////////////////////////////////////////////////////

// EncapsulateBox returns the box expanded to include the given box.
func (b Box) EncapsulateBox(other Box) Box {

	min := b.GetMin().Min(other.GetMin())
	max := b.GetMax().Max(other.GetMax())

	return BoxFromMinMax(min, max)
}

////////////////////////////////////////////////////////////////////////////////

// Expand returns the box with extents increased by half the given amount
// on each side.
func (b Box) Expand(amount float64) Box {

	half := amount * 0.5

	return Box{
		Center:  b.Center,
		Extents: b.Extents.AddScalar(half),
	}
}

////////////////////////////////////////////////////////////////////////////////

// ExpandVector returns the box with extents increased by half the given
// amount on each axis.
func (b Box) ExpandVector(amount math.Vector3) Box {

	return Box{
		Center:  b.Center,
		Extents: b.Extents.Add(amount.MulScalar(0.5)),
	}
}

//----------------------------------------------------------------------------//
// Static                                                                     //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// BoxFromMinMax creates a box from minimum and maximum corner points.
func BoxFromMinMax(min, max math.Vector3) Box {

	return Box{
		Center:  min.Add(max).MulScalar(0.5),
		Extents: max.Sub(min).MulScalar(0.5),
	}
}
