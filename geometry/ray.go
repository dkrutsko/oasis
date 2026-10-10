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

// IntersectEllipsoid tests for intersection with an axis-aligned ellipsoid.
// The test transforms the ray into a scaled space where the ellipsoid becomes
// a unit sphere. The returned distance is in the original ray's units.
func (r Ray) IntersectEllipsoid(ellipsoid Ellipsoid) (float64, bool) {

	// Scale origin and direction so the ellipsoid maps to a unit sphere
	oc := math.Vector3{
		X: (r.Origin.X - ellipsoid.Center.X) / ellipsoid.RadiusX,
		Y: (r.Origin.Y - ellipsoid.Center.Y) / ellipsoid.RadiusY,
		Z: (r.Origin.Z - ellipsoid.Center.Z) / ellipsoid.RadiusZ,
	}

	dir := math.Vector3{
		X: r.Direction.X / ellipsoid.RadiusX,
		Y: r.Direction.Y / ellipsoid.RadiusY,
		Z: r.Direction.Z / ellipsoid.RadiusZ,
	}

	a := dir.Dot(dir)
	b := 2 * oc.Dot(dir)
	c := oc.Dot(oc) - 1

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

	// Try far intersection (origin may be inside the ellipsoid)
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

// IntersectEllipticalCapsule tests for intersection with a capsule whose
// cross-section is an ellipse. The cylinder body and each hemispherical cap
// are tested separately against the ray. Each cap is an oriented half-
// ellipsoid whose axis extent equals the larger of the two cross-section
// radii. Returns the distance along the ray and true if an intersection
// was found.
func (r Ray) IntersectEllipticalCapsule(ec EllipticalCapsule) (float64, bool) {

	//------------------------------------------------------------------------//

	axis := ec.End.Sub(ec.Start)
	axisLenSq := axis.LengthSq()
	if axisLenSq < 1e-20 {
		return 0, false
	}

	axisLen := sysMath.Sqrt(axisLenSq)
	axisNorm := axis.MulScalar(1.0 / axisLen)
	uDir, vDir := ellipticalCapsuleBasis(axisNorm, ec.Rotation)

	// Decompose the ray into the capsule's local frame
	oc := r.Origin.Sub(ec.Start)
	ocA := oc.Dot(axisNorm)
	ocU := oc.Dot(uDir)
	ocV := oc.Dot(vDir)

	dA := r.Direction.Dot(axisNorm)
	dU := r.Direction.Dot(uDir)
	dV := r.Direction.Dot(vDir)

	invA2 := 1.0 / (ec.RadiusA * ec.RadiusA)
	invB2 := 1.0 / (ec.RadiusB * ec.RadiusB)

	best := sysMath.MaxFloat64

	//------------------------------------------------------------------------//

	// Cylinder body: solve the elliptical cylinder quadratic
	// and clip the hit to the segment [0, axisLen]
	qa := dU*dU*invA2 + dV*dV*invB2
	qb := ocU*dU*invA2 + ocV*dV*invB2
	qc := ocU*ocU*invA2 + ocV*ocV*invB2 - 1

	disc := qb*qb - qa*qc
	if disc >= 0 {
		sqrtDisc := sysMath.Sqrt(disc)

		t := (-qb - sqrtDisc) / qa
		s := ocA + t*dA
		if t >= 0 && s >= 0 && s <= axisLen {
			best = t
		}

		if best == sysMath.MaxFloat64 {
			t = (-qb + sqrtDisc) / qa
			s = ocA + t*dA
			if t >= 0 && s >= 0 && s <= axisLen {
				best = t
			}
		}
	}

	//------------------------------------------------------------------------//

	// Each cap is a half-ellipsoid with semi-axes (RadiusA,
	// RadiusB, capR) along (u, v, axisNorm). The axis extent
	// uses the larger cross-section radius so the cap matches
	// a standard hemisphere when both radii are equal.
	capR := sysMath.Max(ec.RadiusA, ec.RadiusB)
	invCapR2 := 1.0 / (capR * capR)

	// Start cap: center = Start, valid half-space s <= 0
	ea := dU*dU*invA2 + dV*dV*invB2 + dA*dA*invCapR2
	eb := ocU*dU*invA2 + ocV*dV*invB2 + ocA*dA*invCapR2
	capC := ocU*ocU*invA2 + ocV*ocV*invB2 + ocA*ocA*invCapR2 - 1

	disc = eb*eb - ea*capC
	if disc >= 0 {
		sqrtDisc := sysMath.Sqrt(disc)

		t := (-eb - sqrtDisc) / ea
		if t >= 0 && t < best && (ocA+t*dA) <= 0 {
			best = t
		}

		t = (-eb + sqrtDisc) / ea
		if t >= 0 && t < best && (ocA+t*dA) <= 0 {
			best = t
		}
	}

	// End cap: center = End, valid half-space s >= 0
	ocEnd := r.Origin.Sub(ec.End)
	ocEndA := ocEnd.Dot(axisNorm)
	ocEndU := ocEnd.Dot(uDir)
	ocEndV := ocEnd.Dot(vDir)

	eb = ocEndU*dU*invA2 + ocEndV*dV*invB2 + ocEndA*dA*invCapR2
	capC = ocEndU*ocEndU*invA2 + ocEndV*ocEndV*invB2 + ocEndA*ocEndA*invCapR2 - 1

	disc = eb*eb - ea*capC
	if disc >= 0 {
		sqrtDisc := sysMath.Sqrt(disc)

		t := (-eb - sqrtDisc) / ea
		if t >= 0 && t < best && (ocEndA+t*dA) >= 0 {
			best = t
		}

		t = (-eb + sqrtDisc) / ea
		if t >= 0 && t < best && (ocEndA+t*dA) >= 0 {
			best = t
		}
	}

	//------------------------------------------------------------------------//

	if best < sysMath.MaxFloat64 {
		return best, true
	}

	return 0, false

	//------------------------------------------------------------------------//
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
