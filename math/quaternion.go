package math

import (
	"errors"
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

// Normalize returns the unit quaternion in the same direction. Returns
// `QuaternionZero` if the magnitude is zero.
func (q Quaternion) Normalize() Quaternion {

	magnitude := sysMath.Sqrt(q.X*q.X + q.Y*q.Y + q.Z*q.Z + q.W*q.W)

	// The default case
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
		sin := sysMath.Acos(dot)
		sin := 1.0 / sysMath.Sin(sin)

		a = sysMath.Sin((1-amount)*sin) * sin
		b = sysMath.Sin(amount*sin) * sin
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

//----------------------------------------------------------------------------//
// Static                                                                     //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// QuaternionFromSlice32 creates a Quaternion from a float32 slice.
func QuaternionFromSlice32(values []float32) (Quaternion, error) {

	if len(values) != 4 {
		return QuaternionZero, errors.New("not enough values")
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
		return QuaternionZero, errors.New("not enough values")
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

// QuaternionFromYawPitchRoll creates a quaternion from yaw, pitch, and roll
// angles in radians.
func QuaternionFromYawPitchRoll(yaw, pitch, roll float64) Quaternion {

	halfYaw := yaw * 0.5
	halfPitch := pitch * 0.5
	halfRoll := roll * 0.5

	ySin := sysMath.Sin(halfYaw)
	yCos := sysMath.Cos(halfYaw)

	pSin := sysMath.Sin(halfPitch)
	pCos := sysMath.Cos(halfPitch)

	rSin := sysMath.Sin(halfRoll)
	rCos := sysMath.Cos(halfRoll)

	return Quaternion{
		yCos*pSin*rCos + ySin*pCos*rSin,
		ySin*pCos*rCos - yCos*pSin*rSin,
		yCos*pCos*rSin - ySin*pSin*rCos,
		yCos*pCos*rCos + ySin*pSin*rSin,
	}
}

////////////////////////////////////////////////////////////////////////////////

// QuaternionFromRotationMatrix creates a quaternion from the rotational part
// of the given matrix using a numerically stable trace-based extraction.
func QuaternionFromRotationMatrix(matrix Matrix) Quaternion {

	trace := matrix.M11 + matrix.M22 + matrix.M33

	if trace > 0 {
		s := sysMath.Sqrt(trace + 1)
		w := s * 0.5
		s = 0.5 / s
		return Quaternion{
			(matrix.M23 - matrix.M32) * s,
			(matrix.M31 - matrix.M13) * s,
			(matrix.M12 - matrix.M21) * s,
			w,
		}
	}

	if matrix.M11 >= matrix.M22 && matrix.M11 >= matrix.M33 {
		s := sysMath.Sqrt(1 + matrix.M11 - matrix.M22 - matrix.M33)
		half := 0.5 / s
		return Quaternion{
			0.5 * s,
			(matrix.M12 + matrix.M21) * half,
			(matrix.M13 + matrix.M31) * half,
			(matrix.M23 - matrix.M32) * half,
		}
	}

	if matrix.M22 > matrix.M33 {
		s := sysMath.Sqrt(1 + matrix.M22 - matrix.M11 - matrix.M33)
		half := 0.5 / s
		return Quaternion{
			(matrix.M21 + matrix.M12) * half,
			0.5 * s,
			(matrix.M32 + matrix.M23) * half,
			(matrix.M31 - matrix.M13) * half,
		}
	}

	s := sysMath.Sqrt(1 + matrix.M33 - matrix.M11 - matrix.M22)
	half := 0.5 / s
	return Quaternion{
		(matrix.M31 + matrix.M13) * half,
		(matrix.M32 + matrix.M23) * half,
		0.5 * s,
		(matrix.M12 - matrix.M21) * half,
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

// Mul returns the Hamilton product of two quaternions. This is the standard
// quaternion multiplication used to combine rotations.
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
