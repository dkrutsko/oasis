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

// RayZero represents a ray with zero origin and direction.
var RayZero = Ray{math.Vector3{X: 0, Y: 0, Z: 0}, math.Vector3{X: 0, Y: 0, Z: 0}}

//----------------------------------------------------------------------------//
// Types                                                                      //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// Ray represents a ray defined by an origin point and a direction vector.
// The direction does not need to be normalized, but intersection results
// will be in units of the direction vector's length.
type Ray struct {
	Origin    math.Vector3
	Direction math.Vector3
}

//----------------------------------------------------------------------------//
// Methods                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// String returns the string representation of the ray.
func (r Ray) String() string {

	return fmt.Sprintf(
		"[[%.2f, %.2f, %.2f][%.2f, %.2f, %.2f]]",
		r.Origin.X, r.Origin.Y, r.Origin.Z,
		r.Direction.X, r.Direction.Y, r.Direction.Z)
}

////////////////////////////////////////////////////////////////////////////////

// GetPoint returns the point at the given distance along the ray.
func (r Ray) GetPoint(distance float64) math.Vector3 {

	return math.Vector3{
		X: r.Origin.X + r.Direction.X*distance,
		Y: r.Origin.Y + r.Direction.Y*distance,
		Z: r.Origin.Z + r.Direction.Z*distance,
	}
}

////////////////////////////////////////////////////////////////////////////////

// ClosestPoint returns the closest point on the ray to the given point.
func (r Ray) ClosestPoint(point math.Vector3) math.Vector3 {

	dirLenSq := r.Direction.LengthSq()

	if dirLenSq < 1e-20 {
		return r.Origin
	}

	t := point.Sub(r.Origin).Dot(r.Direction) / dirLenSq

	// Clamp to ray (t >= 0)
	if t < 0 {
		t = 0
	}

	return r.GetPoint(t)
}

////////////////////////////////////////////////////////////////////////////////

// Distance returns the distance from the given point to the ray.
func (r Ray) Distance(point math.Vector3) float64 {

	closest := r.ClosestPoint(point)
	return point.Sub(closest).Length()
}

////////////////////////////////////////////////////////////////////////////////

// IntersectPlane tests for intersection with a plane. Returns the distance
// along the ray and true if an intersection was found.
func (r Ray) IntersectPlane(plane Plane) (float64, bool) {

	denom := plane.DotNormal(r.Direction)

	if sysMath.Abs(denom) < 1e-10 {
		return 0, false
	}

	t := -plane.DotCoordinate(r.Origin) / denom

	if t < 0 {
		return 0, false
	}

	return t, true
}

////////////////////////////////////////////////////////////////////////////////

// IntersectSphere tests for intersection with a sphere. Returns the distance
// along the ray to the nearest intersection point and true if an intersection
// was found.
func (r Ray) IntersectSphere(sphere Sphere) (float64, bool) {

	oc := r.Origin.Sub(sphere.Center)

	a := r.Direction.Dot(r.Direction)
	b := 2 * oc.Dot(r.Direction)
	c := oc.Dot(oc) - sphere.Radius*sphere.Radius

	discriminant := b*b - 4*a*c

	if discriminant < 0 {
		return 0, false
	}

	sqrtDisc := sysMath.Sqrt(discriminant)

	// Try nearest intersection first
	t := (-b - sqrtDisc) / (2 * a)
	if t >= 0 {
		return t, true
	}

	// Try far intersection (origin may be inside the sphere)
	t = (-b + sqrtDisc) / (2 * a)
	if t >= 0 {
		return t, true
	}

	return 0, false
}

////////////////////////////////////////////////////////////////////////////////

// IntersectBox tests for intersection with an axis-aligned box using the
// slab method. Returns the distance along the ray and true if an intersection
// was found.
func (r Ray) IntersectBox(box Box) (float64, bool) {

	min := box.GetMin()
	max := box.GetMax()

	//------------------------------------------------------------------------//

	// Compute inverse direction for each axis
	var t1x, t2x, t1y, t2y, t1z, t2z float64

	if sysMath.Abs(r.Direction.X) < 1e-20 {
		if r.Origin.X < min.X || r.Origin.X > max.X {
			return 0, false
		}
		t1x = sysMath.Inf(-1)
		t2x = sysMath.Inf(1)
	} else {
		invX := 1 / r.Direction.X
		t1x = (min.X - r.Origin.X) * invX
		t2x = (max.X - r.Origin.X) * invX
	}

	if sysMath.Abs(r.Direction.Y) < 1e-20 {
		if r.Origin.Y < min.Y || r.Origin.Y > max.Y {
			return 0, false
		}
		t1y = sysMath.Inf(-1)
		t2y = sysMath.Inf(1)
	} else {
		invY := 1 / r.Direction.Y
		t1y = (min.Y - r.Origin.Y) * invY
		t2y = (max.Y - r.Origin.Y) * invY
	}

	if sysMath.Abs(r.Direction.Z) < 1e-20 {
		if r.Origin.Z < min.Z || r.Origin.Z > max.Z {
			return 0, false
		}
		t1z = sysMath.Inf(-1)
		t2z = sysMath.Inf(1)
	} else {
		invZ := 1 / r.Direction.Z
		t1z = (min.Z - r.Origin.Z) * invZ
		t2z = (max.Z - r.Origin.Z) * invZ
	}

	//------------------------------------------------------------------------//

	tmin := sysMath.Max(
		sysMath.Max(
			sysMath.Min(t1x, t2x),
			sysMath.Min(t1y, t2y)),
		sysMath.Min(t1z, t2z))

	tmax := sysMath.Min(
		sysMath.Min(
			sysMath.Max(t1x, t2x),
			sysMath.Max(t1y, t2y)),
		sysMath.Max(t1z, t2z))

	//------------------------------------------------------------------------//

	if tmax < 0 || tmin > tmax {
		return 0, false
	}

	if tmin < 0 {
		return tmax, true
	}

	return tmin, true
}

////////////////////////////////////////////////////////////////////////////////

// IntersectCapsule tests for intersection with a capsule. The capsule is
// treated as a cylinder with hemispherical caps. Returns the distance along
// the ray and true if an intersection was found.
func (r Ray) IntersectCapsule(capsule Capsule) (float64, bool) {

	ba := capsule.End.Sub(capsule.Start)
	oa := r.Origin.Sub(capsule.Start)

	baba := ba.Dot(ba)
	bard := ba.Dot(r.Direction)
	baoa := ba.Dot(oa)
	rdoa := r.Direction.Dot(oa)
	oaoa := oa.Dot(oa)

	a := baba - bard*bard
	b := baba*rdoa - baoa*bard
	c := baba*oaoa - baoa*baoa - capsule.Radius*capsule.Radius*baba

	h := b*b - a*c

	if h >= 0 {
		sqrtH := sysMath.Sqrt(h)
		t := (-b - sqrtH) / a

		// Check if hit is on the cylinder body
		y := baoa + t*bard
		if y > 0 && y < baba && t >= 0 {
			return t, true
		}

		// Test the nearer spherical cap
		var oc math.Vector3
		if y <= 0 {
			oc = oa
		} else {
			oc = r.Origin.Sub(capsule.End)
		}

		capB := r.Direction.Dot(oc)
		capC := oc.Dot(oc) - capsule.Radius*capsule.Radius
		capH := capB*capB - capC

		if capH >= 0 {
			t = -capB - sysMath.Sqrt(capH)
			if t >= 0 {
				return t, true
			}
		}
	}

	return 0, false
}

////////////////////////////////////////////////////////////////////////////////

// IntersectCylinder tests for intersection with a capped cylinder. Returns
// the distance along the ray and true if an intersection was found.
func (r Ray) IntersectCylinder(cylinder Cylinder) (float64, bool) {

	ba := cylinder.End.Sub(cylinder.Start)
	oc := r.Origin.Sub(cylinder.Start)

	baba := ba.Dot(ba)
	bard := ba.Dot(r.Direction)
	baoc := ba.Dot(oc)

	k2 := baba - bard*bard
	k1 := baba*r.Direction.Dot(oc) - baoc*bard
	k0 := baba*oc.Dot(oc) - baoc*baoc - cylinder.Radius*cylinder.Radius*baba

	h := k1*k1 - k2*k0

	if h < 0 {
		return 0, false
	}

	sqrtH := sysMath.Sqrt(h)

	//------------------------------------------------------------------------//

	// Test cylinder body
	t := (-k1 - sqrtH) / k2
	y := baoc + t*bard

	if y > 0 && y < baba && t >= 0 {
		return t, true
	}

	//------------------------------------------------------------------------//

	// Test caps - skip if ray is parallel to cylinder axis
	if sysMath.Abs(bard) > 1e-10 {
		var capT float64
		if y <= 0 {
			capT = -baoc / bard
		} else {
			capT = (baba - baoc) / bard
		}

		if capT >= 0 && sysMath.Abs(k1+k2*capT) < sqrtH {
			return capT, true
		}
	}

	//------------------------------------------------------------------------//

	return 0, false
}

////////////////////////////////////////////////////////////////////////////////

// IntersectQuad tests for intersection with a quad. Returns the distance
// along the ray and true if an intersection was found.
func (r Ray) IntersectQuad(quad Quad) (float64, bool) {

	normal := quad.EdgeA.Cross(quad.EdgeB)
	denom := normal.Dot(r.Direction)

	if sysMath.Abs(denom) < 1e-10 {
		return 0, false
	}

	// Intersect with the quad's plane
	t := normal.Dot(quad.Origin.Sub(r.Origin)) / denom

	if t < 0 {
		return 0, false
	}

	// Check if hit point is within the quad bounds
	hitPoint := r.GetPoint(t)
	local := hitPoint.Sub(quad.Origin)

	edgeALenSq := quad.EdgeA.LengthSq()
	edgeBLenSq := quad.EdgeB.LengthSq()

	if edgeALenSq < 1e-20 || edgeBLenSq < 1e-20 {
		return 0, false
	}

	u := local.Dot(quad.EdgeA) / edgeALenSq
	v := local.Dot(quad.EdgeB) / edgeBLenSq

	if u >= 0 && u <= 1 && v >= 0 && v <= 1 {
		return t, true
	}

	return 0, false
}

////////////////////////////////////////////////////////////////////////////////

// IntersectTriangle tests for intersection with a triangle defined by three
// vertices using the Moller-Trumbore algorithm. Returns the distance along
// the ray and true if an intersection was found.
func (r Ray) IntersectTriangle(v0, v1, v2 math.Vector3) (float64, bool) {

	edge1 := v1.Sub(v0)
	edge2 := v2.Sub(v0)

	h := r.Direction.Cross(edge2)
	a := edge1.Dot(h)

	if a > -1e-10 && a < 1e-10 {
		return 0, false
	}

	f := 1 / a
	s := r.Origin.Sub(v0)
	u := f * s.Dot(h)

	if u < 0 || u > 1 {
		return 0, false
	}

	q := s.Cross(edge1)
	v := f * r.Direction.Dot(q)

	if v < 0 || u+v > 1 {
		return 0, false
	}

	t := f * edge2.Dot(q)

	if t > 0 {
		return t, true
	}

	return 0, false
}

////////////////////////////////////////////////////////////////////////////////

// Transform returns the ray transformed by the given matrix. The origin is
// transformed as a point and the direction as a vector.
func (r Ray) Transform(matrix math.Matrix4) Ray {

	return Ray{
		Origin:    math.Vector3TransformVector3(matrix, r.Origin),
		Direction: math.Vector3TransformVector3(matrix.SetTranslation(math.Vector3Zero), r.Direction),
	}
}

//----------------------------------------------------------------------------//
// Static                                                                     //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// RayFromPoints creates a ray from an origin point toward a target point.
// The direction is normalized.
func RayFromPoints(from, to math.Vector3) Ray {

	return Ray{
		Origin:    from,
		Direction: to.Sub(from).Normalize(),
	}
}
