package geometry

import (
	"fmt"

	"github.com/dkrutsko/oasis/math"
)

//----------------------------------------------------------------------------//
// Constants                                                                  //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// PlaneZero represents a plane with a zero normal and zero distance.
var PlaneZero = Plane{math.Vector3{X: 0, Y: 0, Z: 0}, 0}

//----------------------------------------------------------------------------//
// Types                                                                      //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// Plane represents an infinite plane defined by a normal vector and a signed
// distance from the origin. The plane equation is:
// Normal.X*x + Normal.Y*y + Normal.Z*z + Distance = 0
type Plane struct {
	Normal   math.Vector3
	Distance float64
}

//----------------------------------------------------------------------------//
// Methods                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// String returns the string representation of the plane.
func (p Plane) String() string {

	return fmt.Sprintf("[%.2f, %.2f, %.2f, %.2f]",
		p.Normal.X, p.Normal.Y, p.Normal.Z, p.Distance)
}

////////////////////////////////////////////////////////////////////////////////

// Dot computes the dot product of the plane coefficients with a Vector4.
func (p Plane) Dot(value math.Vector4) float64 {

	return p.Normal.X*value.X +
		p.Normal.Y*value.Y +
		p.Normal.Z*value.Z +
		p.Distance*value.W
}

////////////////////////////////////////////////////////////////////////////////

// DotCoordinate returns the signed distance from a point to the plane.
// Positive values indicate the point is on the side the normal points toward.
func (p Plane) DotCoordinate(value math.Vector3) float64 {

	return p.Normal.X*value.X +
		p.Normal.Y*value.Y +
		p.Normal.Z*value.Z +
		p.Distance
}

////////////////////////////////////////////////////////////////////////////////

// DotNormal computes the dot product of the plane normal with a vector.
func (p Plane) DotNormal(value math.Vector3) float64 {

	return p.Normal.X*value.X +
		p.Normal.Y*value.Y +
		p.Normal.Z*value.Z
}

////////////////////////////////////////////////////////////////////////////////

// Normalize returns the plane with a unit-length normal. Both the normal
// and distance are scaled uniformly.
func (p Plane) Normalize() Plane {

	mag := p.Normal.Length()

	if mag < 1.192093e-07 {
		return p
	}

	inv := 1 / mag

	return Plane{
		Normal: math.Vector3{
			X: p.Normal.X * inv,
			Y: p.Normal.Y * inv,
			Z: p.Normal.Z * inv,
		},
		Distance: p.Distance * inv,
	}
}

////////////////////////////////////////////////////////////////////////////////

// ClosestPoint returns the closest point on the plane to the given point.
func (p Plane) ClosestPoint(point math.Vector3) math.Vector3 {

	dist := p.DotCoordinate(point)

	return math.Vector3{
		X: point.X - p.Normal.X*dist,
		Y: point.Y - p.Normal.Y*dist,
		Z: point.Z - p.Normal.Z*dist,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Transform returns the plane transformed by the given matrix. The plane is
// multiplied as a row vector by the transposed inverse of the matrix.
func (p Plane) Transform(matrix math.Matrix4) Plane {

	inv := matrix.Invert()

	nx := p.Normal.X
	ny := p.Normal.Y
	nz := p.Normal.Z
	d := p.Distance

	return Plane{
		Normal: math.Vector3{
			X: nx*inv.M11 + ny*inv.M21 + nz*inv.M31 + d*inv.M41,
			Y: nx*inv.M12 + ny*inv.M22 + nz*inv.M32 + d*inv.M42,
			Z: nx*inv.M13 + ny*inv.M23 + nz*inv.M33 + d*inv.M43,
		},
		Distance: nx*inv.M14 + ny*inv.M24 + nz*inv.M34 + d*inv.M44,
	}
}

////////////////////////////////////////////////////////////////////////////////

// TransformQuaternion returns the plane with its normal rotated by the given
// quaternion. The distance is preserved.
func (p Plane) TransformQuaternion(rotation math.Quaternion) Plane {

	return Plane{
		Normal:   p.Normal.ApplyQuaternion(rotation),
		Distance: p.Distance,
	}
}

//----------------------------------------------------------------------------//
// Static                                                                     //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// PlaneFromPoints creates a plane from three points by computing the normal
// from the cross product of the two edge vectors.
func PlaneFromPoints(p1, p2, p3 math.Vector3) Plane {

	edge1 := p2.Sub(p1)
	edge2 := p3.Sub(p1)

	normal := edge1.Cross(edge2).Normalize()
	distance := -normal.Dot(p1)

	return Plane{
		Normal:   normal,
		Distance: distance,
	}
}

////////////////////////////////////////////////////////////////////////////////

// PlaneFromPointNormal creates a plane from a point on the plane and a normal
// direction. The normal does not need to be normalized.
func PlaneFromPointNormal(point, normal math.Vector3) Plane {

	n := normal.Normalize()

	return Plane{
		Normal:   n,
		Distance: -n.Dot(point),
	}
}

