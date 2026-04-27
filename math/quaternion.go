package math

import (
	"fmt"
	sysMath "math"
)

//----------------------------------------------------------------------------//
// Constants                                                                  //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

var (
	// QuaternionZero represents a zero quaternion.
	QuaternionZero = Quaternion{0, 0, 0, 0}

	// QuaternionIdentity represents an identity quaternion with no rotation.
	QuaternionIdentity = Quaternion{0, 0, 0, 1}
)

//----------------------------------------------------------------------------//
// Types                                                                      //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// Quaternion represents a rotation as a four-component quaternion where X, Y,
// and Z define the vector part and W defines the scalar part.
type Quaternion struct {
	// X component of the quaternion.
	X float64

	// Y component of the quaternion.
	Y float64

	// Z component of the quaternion.
	Z float64

	// W component of the quaternion.
	W float64
}

//----------------------------------------------------------------------------//
// Methods                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// String returns the string representation of the quaternion.
func (q Quaternion) String() string {

	return fmt.Sprintf("[%.2f, %.2f, %.2f, %.2f]", q.X, q.Y, q.Z, q.W)
}

////////////////////////////////////////////////////////////////////////////////

// IsZero returns whether all components are zero.
func (q Quaternion) IsZero() bool {

	return q.X == 0 && q.Y == 0 && q.Z == 0 && q.W == 0
}

////////////////////////////////////////////////////////////////////////////////

// Normalize returns the unit quaternion in the same direction. Returns the
// zero quaternion if the magnitude is zero.
func (q Quaternion) Normalize() Quaternion {

	magnitude := sysMath.Sqrt(q.X*q.X + q.Y*q.Y + q.Z*q.Z + q.W*q.W)

	if magnitude == 0 {
		return QuaternionZero
	}

	d := 1 / magnitude

	return Quaternion{
		q.X * d,
		q.Y * d,
		q.Z * d,
		q.W * d,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Conjugate returns the conjugate of the quaternion by negating the vector
// part while preserving the scalar part.
func (q Quaternion) Conjugate() Quaternion {

	return Quaternion{-q.X, -q.Y, -q.Z, q.W}
}

////////////////////////////////////////////////////////////////////////////////

// Inverse returns the multiplicative inverse of the quaternion. For unit
// quaternions this is equivalent to the conjugate.
func (q Quaternion) Inverse() Quaternion {

	d := 1 / (q.X*q.X + q.Y*q.Y + q.Z*q.Z + q.W*q.W)

	return Quaternion{
		-q.X * d,
		-q.Y * d,
		-q.Z * d,
		q.W * d,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Dot returns the dot product with another quaternion.
func (q Quaternion) Dot(value Quaternion) float64 {

	return q.X*value.X + q.Y*value.Y + q.Z*value.Z + q.W*value.W
}

////////////////////////////////////////////////////////////////////////////////

// Lerp performs normalized linear interpolation toward the target. The dot
// product is used to determine the shortest interpolation path.
func (q Quaternion) Lerp(target Quaternion, amount float64) Quaternion {

	inverse := 1 - amount

	var result Quaternion

	if q.Dot(target) >= 0 {
		result = Quaternion{
			inverse*q.X + amount*target.X,
			inverse*q.Y + amount*target.Y,
			inverse*q.Z + amount*target.Z,
			inverse*q.W + amount*target.W,
		}

	} else {
		result = Quaternion{
			inverse*q.X - amount*target.X,
			inverse*q.Y - amount*target.Y,
			inverse*q.Z - amount*target.Z,
			inverse*q.W - amount*target.W,
		}
	}

	return result.Normalize()
}

////////////////////////////////////////////////////////////////////////////////

// Slerp performs spherical linear interpolation toward the target. Falls back
// to linear interpolation when the quaternions are nearly identical.
func (q Quaternion) Slerp(target Quaternion, amount float64) Quaternion {

	var a, b float64
	dot := q.Dot(target)
	negative := false

	if dot < 0 {
		negative = true
		dot = -dot
	}

	if dot > 0.999999 {
		a = 1 - amount
		b = amount
		if negative {
			b = -b
		}

	} else {
		cos := sysMath.Acos(dot)
		sin := 1.0 / sysMath.Sin(cos)

		a = sysMath.Sin((1-amount)*cos) * sin
		b = sysMath.Sin(amount*cos) * sin
		if negative {
			b = -b
		}
	}

	return Quaternion{
		a*q.X + b*target.X,
		a*q.Y + b*target.Y,
		a*q.Z + b*target.Z,
		a*q.W + b*target.W,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Length returns the magnitude of the quaternion.
func (q Quaternion) Length() float64 {

	return sysMath.Sqrt(q.X*q.X + q.Y*q.Y + q.Z*q.Z + q.W*q.W)
}

////////////////////////////////////////////////////////////////////////////////

// LengthSq returns the squared magnitude of the quaternion.
func (q Quaternion) LengthSq() float64 {

	return q.X*q.X + q.Y*q.Y + q.Z*q.Z + q.W*q.W
}

////////////////////////////////////////////////////////////////////////////////

// ToVector4 returns a `Vector4` with the same components.
func (q Quaternion) ToVector4() Vector4 {

	return Vector4{q.X, q.Y, q.Z, q.W}
}

////////////////////////////////////////////////////////////////////////////////

// ToMatrix3 returns a 3x3 rotation matrix from the quaternion.
func (q Quaternion) ToMatrix3() Matrix3 {

	x1 := q.X
	y1 := q.Y
	z1 := q.Z
	w1 := q.W

	x2 := x1 + x1
	y2 := y1 + y1
	z2 := z1 + z1

	xx := x1 * x2
	yx := y1 * x2
	yy := y1 * y2

	zx := z1 * x2
	zy := z1 * y2
	zz := z1 * z2

	wx := w1 * x2
	wy := w1 * y2
	wz := w1 * z2

	return Matrix3{
		1 - yy - zz, yx + wz, zx - wy,
		yx - wz, 1 - xx - zz, zy + wx,
		zx + wy, zy - wx, 1 - xx - yy,
	}
}

////////////////////////////////////////////////////////////////////////////////

// ToMatrix4 returns a 4x4 rotation matrix from the quaternion.
func (q Quaternion) ToMatrix4() Matrix4 {

	xx := q.X * q.X
	yy := q.Y * q.Y
	zz := q.Z * q.Z
	xy := q.X * q.Y
	zw := q.Z * q.W
	zx := q.Z * q.X
	yw := q.Y * q.W
	yz := q.Y * q.Z
	xw := q.X * q.W

	return Matrix4{
		1 - 2*(yy+zz), 2 * (xy + zw), 2 * (zx - yw), 0,
		2 * (xy - zw), 1 - 2*(zz+xx), 2 * (yz + xw), 0,
		2 * (zx + yw), 2 * (yz - xw), 1 - 2*(xx+yy), 0,
		0, 0, 0, 1,
	}
}

////////////////////////////////////////////////////////////////////////////////

// ToSlice32 returns the components as a float32 slice.
func (q Quaternion) ToSlice32() []float32 {

	return []float32{
		float32(q.X),
		float32(q.Y),
		float32(q.Z),
		float32(q.W),
	}
}

////////////////////////////////////////////////////////////////////////////////

// ToSlice64 returns the components as a float64 slice.
func (q Quaternion) ToSlice64() []float64 {

	return []float64{q.X, q.Y, q.Z, q.W}
}

////////////////////////////////////////////////////////////////////////////////

// ToPacked compresses the quaternion into a 64-bit integer. The
// packed format stores X in the upper 22 bits and Y and Z in 21
// bits each. The W component is discarded and reconstructed from
// the unit quaternion constraint when unpacking.
func (q Quaternion) ToPacked() int64 {

	x := int64(q.X * 2097152.0)
	y := int64(q.Y * 1048576.0)
	z := int64(q.Z * 1048576.0)

	return (x << 42) | ((y & 0x1FFFFF) << 21) | (z & 0x1FFFFF)
}

//----------------------------------------------------------------------------//
// Static                                                                     //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// QuaternionFromSlice32 creates a Quaternion from a float32 slice.
func QuaternionFromSlice32(values []float32) (Quaternion, error) {

	if len(values) != 4 {
		return QuaternionZero, ErrInvalidLength
	}

	q := Quaternion{
		X: float64(values[0]),
		Y: float64(values[1]),
		Z: float64(values[2]),
		W: float64(values[3]),
	}

	return q, nil
}

////////////////////////////////////////////////////////////////////////////////

// QuaternionFromSlice64 creates a Quaternion from a float64 slice.
func QuaternionFromSlice64(values []float64) (Quaternion, error) {

	if len(values) != 4 {
		return QuaternionZero, ErrInvalidLength
	}

	q := Quaternion{
		X: values[0],
		Y: values[1],
		Z: values[2],
		W: values[3],
	}

	return q, nil
}

////////////////////////////////////////////////////////////////////////////////

// QuaternionFromAxisAngle creates a quaternion representing a rotation around
// the given axis by the specified angle in radians.
func QuaternionFromAxisAngle(axis Vector3, angle float64) Quaternion {

	half := angle * 0.5
	sin := sysMath.Sin(half)
	cos := sysMath.Cos(half)

	return Quaternion{
		axis.X * sin,
		axis.Y * sin,
		axis.Z * sin,
		cos,
	}
}

////////////////////////////////////////////////////////////////////////////////

// QuaternionFromPacked decompresses a packed 64-bit quaternion into its four
// components. The packed format stores x in the upper 22 bits and y and z in
// 21 bits each. The w component is reconstructed from the unit quaternion
// constraint.
func QuaternionFromPacked(packed int64) Quaternion {

	const a = 1 / 2097152.0
	const b = 1 / 1048576.0

	x := float64((packed<<0)>>42) * a
	y := float64((packed<<22)>>43) * b
	z := float64((packed<<43)>>43) * b

	wSquared := x*x + y*y + z*z

	w := 0.0
	if wSquared < 1 && sysMath.Abs(wSquared-1) >= b {
		w = sysMath.Sqrt(1 - wSquared)
	}

	return Quaternion{x, y, z, w}
}

//----------------------------------------------------------------------------//
// Operators                                                                  //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// Add returns the component-wise sum of two quaternions.
func (q Quaternion) Add(value Quaternion) Quaternion {

	return Quaternion{
		q.X + value.X,
		q.Y + value.Y,
		q.Z + value.Z,
		q.W + value.W,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Sub returns the component-wise difference of two quaternions.
func (q Quaternion) Sub(value Quaternion) Quaternion {

	return Quaternion{
		q.X - value.X,
		q.Y - value.Y,
		q.Z - value.Z,
		q.W - value.W,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Mul returns the Hamilton product of two quaternions. This
// combines the rotations represented by each quaternion.
func (q Quaternion) Mul(value Quaternion) Quaternion {

	return Quaternion{
		q.W*value.X + q.X*value.W + q.Z*value.Y - q.Y*value.Z,
		q.W*value.Y + q.Y*value.W + q.X*value.Z - q.Z*value.X,
		q.W*value.Z + q.Z*value.W + q.Y*value.X - q.X*value.Y,
		q.W*value.W - q.X*value.X - q.Y*value.Y - q.Z*value.Z,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Div returns the result of multiplying the quaternion by the inverse of the
// given quaternion.
func (q Quaternion) Div(value Quaternion) Quaternion {

	d := 1 / (value.X*value.X + value.Y*value.Y + value.Z*value.Z + value.W*value.W)
	ix := -value.X * d
	iy := -value.Y * d
	iz := -value.Z * d
	iw := value.W * d

	return Quaternion{
		q.X*iw + ix*q.W + q.Y*iz - q.Z*iy,
		q.Y*iw + iy*q.W + q.Z*ix - q.X*iz,
		q.Z*iw + iz*q.W + q.X*iy - q.Y*ix,
		q.W*iw - q.X*ix - q.Y*iy - q.Z*iz,
	}
}

////////////////////////////////////////////////////////////////////////////////

// AddScalar adds a scalar to each component.
func (q Quaternion) AddScalar(scalar float64) Quaternion {

	return Quaternion{
		q.X + scalar,
		q.Y + scalar,
		q.Z + scalar,
		q.W + scalar,
	}
}

////////////////////////////////////////////////////////////////////////////////

// SubScalar subtracts a scalar from each component.
func (q Quaternion) SubScalar(scalar float64) Quaternion {

	return Quaternion{
		q.X - scalar,
		q.Y - scalar,
		q.Z - scalar,
		q.W - scalar,
	}
}

////////////////////////////////////////////////////////////////////////////////

// MulScalar multiplies each component by a scalar.
func (q Quaternion) MulScalar(scalar float64) Quaternion {

	return Quaternion{
		q.X * scalar,
		q.Y * scalar,
		q.Z * scalar,
		q.W * scalar,
	}
}

////////////////////////////////////////////////////////////////////////////////

// DivScalar divides each component by a scalar.
func (q Quaternion) DivScalar(scalar float64) Quaternion {

	return Quaternion{
		q.X / scalar,
		q.Y / scalar,
		q.Z / scalar,
		q.W / scalar,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Neg returns the negation of the quaternion.
func (q Quaternion) Neg() Quaternion {

	return Quaternion{
		-q.X,
		-q.Y,
		-q.Z,
		-q.W,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Compare performs lexicographic comparison of two quaternions. Returns
// -1, 0, or 1.
func (q Quaternion) Compare(value Quaternion) int {

	if q.X < value.X {
		return -1
	}
	if q.X > value.X {
		return 1
	}

	if q.Y < value.Y {
		return -1
	}
	if q.Y > value.Y {
		return 1
	}

	if q.Z < value.Z {
		return -1
	}
	if q.Z > value.Z {
		return 1
	}

	if q.W < value.W {
		return -1
	}
	if q.W > value.W {
		return 1
	}

	return 0
}

////////////////////////////////////////////////////////////////////////////////

// Lt returns whether the quaternion is lexicographically less than the given
// quaternion.
func (q Quaternion) Lt(value Quaternion) bool {
	return q.Compare(value) < 0
}

////////////////////////////////////////////////////////////////////////////////

// Gt returns whether the quaternion is lexicographically greater than the
// given quaternion.
func (q Quaternion) Gt(value Quaternion) bool {
	return q.Compare(value) > 0
}

////////////////////////////////////////////////////////////////////////////////

// Le returns whether the quaternion is lexicographically less than or equal to
// the given quaternion.
func (q Quaternion) Le(value Quaternion) bool {
	return q.Compare(value) <= 0
}

////////////////////////////////////////////////////////////////////////////////

// Ge returns whether the quaternion is lexicographically greater than or equal
// to the given quaternion.
func (q Quaternion) Ge(value Quaternion) bool {
	return q.Compare(value) >= 0
}

////////////////////////////////////////////////////////////////////////////////

// Eq returns whether all components are equal.
func (q Quaternion) Eq(value Quaternion) bool {

	return q.X == value.X && q.Y == value.Y && q.Z == value.Z && q.W == value.W
}

////////////////////////////////////////////////////////////////////////////////

// Ne returns whether any component is not equal.
func (q Quaternion) Ne(value Quaternion) bool {

	return q.X != value.X || q.Y != value.Y || q.Z != value.Z || q.W != value.W
}
