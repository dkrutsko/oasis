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

// QuadZero represents a quad with zero dimensions at the origin.
var QuadZero = Quad{math.Vector3{X: 0, Y: 0, Z: 0}, math.Vector3{X: 0, Y: 0, Z: 0}, math.Vector3{X: 0, Y: 0, Z: 0}}

//----------------------------------------------------------------------------//
// Types                                                                      //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// Quad represents a finite rectangular region in 3D space defined by a corner
// point and two edge vectors. The four corners are Origin, Origin+EdgeA,
// Origin+EdgeB, and Origin+EdgeA+EdgeB.
type Quad struct {
	Origin math.Vector3
	EdgeA  math.Vector3
	EdgeB  math.Vector3
}

//----------------------------------------------------------------------------//
// Methods                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// String returns the string representation of the quad.
func (q Quad) String() string {

	return fmt.Sprintf(
		"[[%.2f, %.2f, %.2f][%.2f, %.2f, %.2f][%.2f, %.2f, %.2f]]",
		q.Origin.X, q.Origin.Y, q.Origin.Z,
		q.EdgeA.X, q.EdgeA.Y, q.EdgeA.Z,
		q.EdgeB.X, q.EdgeB.Y, q.EdgeB.Z)
}

////////////////////////////////////////////////////////////////////////////////

// GetNormal returns the unit normal of the quad, computed from the cross
// product of the two edge vectors.
func (q Quad) GetNormal() math.Vector3 {

	return q.EdgeA.Cross(q.EdgeB).Normalize()
}

////////////////////////////////////////////////////////////////////////////////

// GetCenter returns the center point of the quad.
func (q Quad) GetCenter() math.Vector3 {

	half := q.EdgeA.Add(q.EdgeB).MulScalar(0.5)
	return q.Origin.Add(half)
}

////////////////////////////////////////////////////////////////////////////////

// GetArea returns the surface area of the quad.
func (q Quad) GetArea() float64 {

	return q.EdgeA.Cross(q.EdgeB).Length()
}

////////////////////////////////////////////////////////////////////////////////

// GetVertex returns one of the four corners of the quad by index. The corners
// are ordered: 0=Origin, 1=Origin+EdgeA, 2=Origin+EdgeB, 3=Origin+EdgeA+EdgeB.
func (q Quad) GetVertex(index int) math.Vector3 {

	switch index {
	case 0:
		return q.Origin
	case 1:
		return q.Origin.Add(q.EdgeA)
	case 2:
		return q.Origin.Add(q.EdgeB)
	case 3:
		return q.Origin.Add(q.EdgeA).Add(q.EdgeB)
	default:
		panic("quad: vertex index out of range")
	}
}

////////////////////////////////////////////////////////////////////////////////

// Contains returns whether the point lies on the quad surface within a small
// tolerance. The point must be on the quad's plane and within its bounds.
func (q Quad) Contains(point math.Vector3) bool {

	local := point.Sub(q.Origin)

	// Check if point is on the plane
	normal := q.EdgeA.Cross(q.EdgeB)
	normalLen := normal.Length()

	if normalLen < 1e-10 {
		return false
	}

	planeDist := sysMath.Abs(local.Dot(normal)) / normalLen

	if planeDist > 1e-6 {
		return false
	}

	// Check if within bounds using edge projections
	edgeALenSq := q.EdgeA.LengthSq()
	edgeBLenSq := q.EdgeB.LengthSq()

	if edgeALenSq < 1e-20 || edgeBLenSq < 1e-20 {
		return false
	}

	u := local.Dot(q.EdgeA) / edgeALenSq
	v := local.Dot(q.EdgeB) / edgeBLenSq

	return u >= 0 && u <= 1 && v >= 0 && v <= 1
}

////////////////////////////////////////////////////////////////////////////////

// ClosestPoint returns the closest point on the quad surface to the given
// point.
func (q Quad) ClosestPoint(point math.Vector3) math.Vector3 {

	local := point.Sub(q.Origin)

	edgeALenSq := q.EdgeA.LengthSq()
	edgeBLenSq := q.EdgeB.LengthSq()

	// Handle degenerate quads
	if edgeALenSq < 1e-20 || edgeBLenSq < 1e-20 {
		return q.Origin
	}

	// Project onto edge vectors and clamp to [0, 1]
	u := math.Clamp(local.Dot(q.EdgeA)/edgeALenSq, 0, 1)
	v := math.Clamp(local.Dot(q.EdgeB)/edgeBLenSq, 0, 1)

	return q.Origin.Add(q.EdgeA.MulScalar(u)).Add(q.EdgeB.MulScalar(v))
}

//----------------------------------------------------------------------------//
// Static                                                                     //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// QuadFromCenterNormal creates a quad centered at the given point with the
// specified normal, up direction, width, and height.
func QuadFromCenterNormal(
	center, normal, up math.Vector3, width, height float64,
) Quad {

	n := normal.Normalize()
	right := up.Cross(n).Normalize()
	actualUp := n.Cross(right).Normalize()

	halfW := width * 0.5
	halfH := height * 0.5

	edgeA := right.MulScalar(width)
	edgeB := actualUp.MulScalar(height)

	origin := center.Sub(right.MulScalar(halfW)).Sub(actualUp.MulScalar(halfH))

	return Quad{
		Origin: origin,
		EdgeA:  edgeA,
		EdgeB:  edgeB,
	}
}

