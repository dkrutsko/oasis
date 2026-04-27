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
	// Vector4Zero represents a zero vector.
	Vector4Zero = Vector4{0, 0, 0, 0}

	// Vector4UnitX represents a unit vector along the x-axis.
	Vector4UnitX = Vector4{1, 0, 0, 0}

	// Vector4UnitY represents a unit vector along the y-axis.
	Vector4UnitY = Vector4{0, 1, 0, 0}

	// Vector4UnitZ represents a unit vector along the z-axis.
	Vector4UnitZ = Vector4{0, 0, 1, 0}

	// Vector4UnitW represents a unit vector along the w-axis.
	Vector4UnitW = Vector4{0, 0, 0, 1}
)

//----------------------------------------------------------------------------//
// Types                                                                      //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// Vector4 represents a vector with four components.
type Vector4 struct {
	// X component of the vector.
	X float64

	// Y component of the vector.
	Y float64

	// Z component of the vector.
	Z float64

	// W component of the vector.
	W float64
}

//----------------------------------------------------------------------------//
// Methods                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// String returns the string representation of the vector.
func (v Vector4) String() string {

	return fmt.Sprintf("[%.2f, %.2f, %.2f, %.2f]", v.X, v.Y, v.Z, v.W)
}

////////////////////////////////////////////////////////////////////////////////

// IsZero returns whether all components are zero.
func (v Vector4) IsZero() bool {

	return v.X == 0 && v.Y == 0 && v.Z == 0 && v.W == 0
}

////////////////////////////////////////////////////////////////////////////////

// Normalize returns a unit vector in the same direction. Returns the zero
// vector if the magnitude is zero.
func (v Vector4) Normalize() Vector4 {

	magnitude := sysMath.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z + v.W*v.W)

	if magnitude == 0 {
		return Vector4Zero
	}

	return Vector4{
		v.X / magnitude,
		v.Y / magnitude,
		v.Z / magnitude,
		v.W / magnitude,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Distance returns the Euclidean distance to another vector.
func (v Vector4) Distance(value Vector4) float64 {

	dx := v.X - value.X
	dy := v.Y - value.Y
	dz := v.Z - value.Z
	dw := v.W - value.W

	return sysMath.Sqrt(dx*dx + dy*dy + dz*dz + dw*dw)
}

////////////////////////////////////////////////////////////////////////////////

// DistanceSq returns the squared Euclidean distance to another vector.
func (v Vector4) DistanceSq(value Vector4) float64 {

	dx := v.X - value.X
	dy := v.Y - value.Y
	dz := v.Z - value.Z
	dw := v.W - value.W

	return dx*dx + dy*dy + dz*dz + dw*dw
}

////////////////////////////////////////////////////////////////////////////////

// Dot returns the dot product with another vector.
func (v Vector4) Dot(value Vector4) float64 {

	return v.X*value.X + v.Y*value.Y + v.Z*value.Z + v.W*value.W
}

////////////////////////////////////////////////////////////////////////////////

// Reflect returns the reflection of the vector off a surface defined by the
// given normal.
func (v Vector4) Reflect(normal Vector4) Vector4 {

	dot := 2 * (v.X*normal.X + v.Y*normal.Y + v.Z*normal.Z + v.W*normal.W)

	return Vector4{
		v.X - dot*normal.X,
		v.Y - dot*normal.Y,
		v.Z - dot*normal.Z,
		v.W - dot*normal.W,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Min returns the component-wise minimum of two vectors.
func (v Vector4) Min(value Vector4) Vector4 {

	x := v.X
	if value.X < x {
		x = value.X
	}

	y := v.Y
	if value.Y < y {
		y = value.Y
	}

	z := v.Z
	if value.Z < z {
		z = value.Z
	}

	w := v.W
	if value.W < w {
		w = value.W
	}

	return Vector4{x, y, z, w}
}

////////////////////////////////////////////////////////////////////////////////

// Max returns the component-wise maximum of two vectors.
func (v Vector4) Max(value Vector4) Vector4 {

	x := v.X
	if value.X > x {
		x = value.X
	}

	y := v.Y
	if value.Y > y {
		y = value.Y
	}

	z := v.Z
	if value.Z > z {
		z = value.Z
	}

	w := v.W
	if value.W > w {
		w = value.W
	}

	return Vector4{x, y, z, w}
}

////////////////////////////////////////////////////////////////////////////////

// Clamp restricts each component to the specified range.
func (v Vector4) Clamp(min, max Vector4) Vector4 {

	x := v.X
	if x > max.X {
		x = max.X
	} else if x < min.X {
		x = min.X
	}

	y := v.Y
	if y > max.Y {
		y = max.Y
	} else if y < min.Y {
		y = min.Y
	}

	z := v.Z
	if z > max.Z {
		z = max.Z
	} else if z < min.Z {
		z = min.Z
	}

	w := v.W
	if w > max.W {
		w = max.W
	} else if w < min.W {
		w = min.W
	}

	return Vector4{x, y, z, w}
}

////////////////////////////////////////////////////////////////////////////////

// Lerp performs linear interpolation toward the target.
func (v Vector4) Lerp(target Vector4, amount float64) Vector4 {

	return Vector4{
		v.X + (target.X-v.X)*amount,
		v.Y + (target.Y-v.Y)*amount,
		v.Z + (target.Z-v.Z)*amount,
		v.W + (target.W-v.W)*amount,
	}
}

////////////////////////////////////////////////////////////////////////////////

// SmoothStep performs Hermite interpolation toward the target with smoothing
// at the edges.
func (v Vector4) SmoothStep(target Vector4, amount float64) Vector4 {

	amount = Clamp(amount, 0, 1)
	amount = amount * amount * (3 - 2*amount)

	return Vector4{
		v.X + (target.X-v.X)*amount,
		v.Y + (target.Y-v.Y)*amount,
		v.Z + (target.Z-v.Z)*amount,
		v.W + (target.W-v.W)*amount,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Length returns the magnitude of the vector.
func (v Vector4) Length() float64 {

	return sysMath.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z + v.W*v.W)
}

////////////////////////////////////////////////////////////////////////////////

// LengthSq returns the squared magnitude of the vector.
func (v Vector4) LengthSq() float64 {

	return v.X*v.X + v.Y*v.Y + v.Z*v.Z + v.W*v.W
}

////////////////////////////////////////////////////////////////////////////////

// ToVector2 returns a Vector2 by dropping the Z and W components.
func (v Vector4) ToVector2() Vector2 {

	return Vector2{v.X, v.Y}
}

////////////////////////////////////////////////////////////////////////////////

// ToVector3 returns a Vector3 by dropping the W component.
func (v Vector4) ToVector3() Vector3 {

	return Vector3{v.X, v.Y, v.Z}
}

////////////////////////////////////////////////////////////////////////////////

// ToQuaternion returns a `Quaternion` with the same components.
func (v Vector4) ToQuaternion() Quaternion {

	return Quaternion{v.X, v.Y, v.Z, v.W}
}

////////////////////////////////////////////////////////////////////////////////

// ToColor returns a `Color` with X, Y, Z, W mapped to R, G, B, A.
func (v Vector4) ToColor() Color {

	return Color{v.X, v.Y, v.Z, v.W}
}

////////////////////////////////////////////////////////////////////////////////

// ToSlice32 returns the components as a float32 slice.
func (v Vector4) ToSlice32() []float32 {

	return []float32{
		float32(v.X),
		float32(v.Y),
		float32(v.Z),
		float32(v.W),
	}
}

////////////////////////////////////////////////////////////////////////////////

// ToSlice64 returns the components as a float64 slice.
func (v Vector4) ToSlice64() []float64 {

	return []float64{
		v.X,
		v.Y,
		v.Z,
		v.W,
	}
}

//----------------------------------------------------------------------------//
// Static                                                                     //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// Vector4FromSlice32 creates a Vector4 from a float32 slice.
func Vector4FromSlice32(values []float32) (Vector4, error) {

	if len(values) != 4 {
		return Vector4Zero, ErrInvalidLength
	}

	v := Vector4{
		X: float64(values[0]),
		Y: float64(values[1]),
		Z: float64(values[2]),
		W: float64(values[3]),
	}

	return v, nil
}

////////////////////////////////////////////////////////////////////////////////

// Vector4FromSlice64 creates a Vector4 from a float64 slice.
func Vector4FromSlice64(values []float64) (Vector4, error) {

	if len(values) != 4 {
		return Vector4Zero, ErrInvalidLength
	}

	v := Vector4{
		X: values[0],
		Y: values[1],
		Z: values[2],
		W: values[3],
	}

	return v, nil
}

////////////////////////////////////////////////////////////////////////////////

// Vector4TransformVector2 transforms a Vector2 by a matrix and returns the
// resulting Vector4.
func Vector4TransformVector2(matrix Matrix4, value Vector2) Vector4 {

	return Vector4{
		matrix.M11*value.X + matrix.M21*value.Y + matrix.M41,
		matrix.M12*value.X + matrix.M22*value.Y + matrix.M42,
		matrix.M13*value.X + matrix.M23*value.Y + matrix.M43,
		matrix.M14*value.X + matrix.M24*value.Y + matrix.M44,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Vector4TransformVector3 transforms a Vector3 by a matrix and returns the
// resulting Vector4.
func Vector4TransformVector3(matrix Matrix4, value Vector3) Vector4 {

	return Vector4{
		matrix.M11*value.X + matrix.M21*value.Y + matrix.M31*value.Z + matrix.M41,
		matrix.M12*value.X + matrix.M22*value.Y + matrix.M32*value.Z + matrix.M42,
		matrix.M13*value.X + matrix.M23*value.Y + matrix.M33*value.Z + matrix.M43,
		matrix.M14*value.X + matrix.M24*value.Y + matrix.M34*value.Z + matrix.M44,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Vector4TransformVector4 transforms a Vector4 by a matrix and returns the
// resulting Vector4.
func Vector4TransformVector4(matrix Matrix4, value Vector4) Vector4 {

	return Vector4{
		matrix.M11*value.X + matrix.M21*value.Y + matrix.M31*value.Z + matrix.M41*value.W,
		matrix.M12*value.X + matrix.M22*value.Y + matrix.M32*value.Z + matrix.M42*value.W,
		matrix.M13*value.X + matrix.M23*value.Y + matrix.M33*value.Z + matrix.M43*value.W,
		matrix.M14*value.X + matrix.M24*value.Y + matrix.M34*value.Z + matrix.M44*value.W,
	}
}

//----------------------------------------------------------------------------//
// Operators                                                                  //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// Add returns the component-wise sum of two vectors.
func (v Vector4) Add(value Vector4) Vector4 {

	return Vector4{
		v.X + value.X,
		v.Y + value.Y,
		v.Z + value.Z,
		v.W + value.W,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Sub returns the component-wise difference of two vectors.
func (v Vector4) Sub(value Vector4) Vector4 {

	return Vector4{
		v.X - value.X,
		v.Y - value.Y,
		v.Z - value.Z,
		v.W - value.W,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Mul returns the component-wise product of two vectors.
func (v Vector4) Mul(value Vector4) Vector4 {

	return Vector4{
		v.X * value.X,
		v.Y * value.Y,
		v.Z * value.Z,
		v.W * value.W,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Div returns the component-wise quotient of two vectors.
func (v Vector4) Div(value Vector4) Vector4 {

	return Vector4{
		v.X / value.X,
		v.Y / value.Y,
		v.Z / value.Z,
		v.W / value.W,
	}
}

////////////////////////////////////////////////////////////////////////////////

// AddScalar adds a scalar to each component.
func (v Vector4) AddScalar(scalar float64) Vector4 {

	return Vector4{
		v.X + scalar,
		v.Y + scalar,
		v.Z + scalar,
		v.W + scalar,
	}
}

////////////////////////////////////////////////////////////////////////////////

// SubScalar subtracts a scalar from each component.
func (v Vector4) SubScalar(scalar float64) Vector4 {

	return Vector4{
		v.X - scalar,
		v.Y - scalar,
		v.Z - scalar,
		v.W - scalar,
	}
}

////////////////////////////////////////////////////////////////////////////////

// MulScalar multiplies each component by a scalar.
func (v Vector4) MulScalar(scalar float64) Vector4 {

	return Vector4{
		v.X * scalar,
		v.Y * scalar,
		v.Z * scalar,
		v.W * scalar,
	}
}

////////////////////////////////////////////////////////////////////////////////

// DivScalar divides each component by a scalar.
func (v Vector4) DivScalar(scalar float64) Vector4 {

	return Vector4{
		v.X / scalar,
		v.Y / scalar,
		v.Z / scalar,
		v.W / scalar,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Neg returns the negation of the vector.
func (v Vector4) Neg() Vector4 {

	return Vector4{
		-v.X,
		-v.Y,
		-v.Z,
		-v.W,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Compare performs lexicographic comparison of two vectors. Returns -1, 0, or 1.
func (v Vector4) Compare(value Vector4) int {

	if v.X < value.X {
		return -1
	}
	if v.X > value.X {
		return 1
	}

	if v.Y < value.Y {
		return -1
	}
	if v.Y > value.Y {
		return 1
	}

	if v.Z < value.Z {
		return -1
	}
	if v.Z > value.Z {
		return 1
	}

	if v.W < value.W {
		return -1
	}
	if v.W > value.W {
		return 1
	}

	return 0
}

////////////////////////////////////////////////////////////////////////////////

// Lt returns whether the vector is lexicographically less than the given vector.
func (v Vector4) Lt(value Vector4) bool {
	return v.Compare(value) < 0
}

////////////////////////////////////////////////////////////////////////////////

// Gt returns whether the vector is lexicographically greater than the given vector.
func (v Vector4) Gt(value Vector4) bool {
	return v.Compare(value) > 0
}

////////////////////////////////////////////////////////////////////////////////

// Le returns whether the vector is lexicographically less than or equal to
// the given vector.
func (v Vector4) Le(value Vector4) bool {
	return v.Compare(value) <= 0
}

////////////////////////////////////////////////////////////////////////////////

// Ge returns whether the vector is lexicographically greater than or equal to
// the given vector.
func (v Vector4) Ge(value Vector4) bool {
	return v.Compare(value) >= 0
}

////////////////////////////////////////////////////////////////////////////////

// Eq returns whether all components are equal.
func (v Vector4) Eq(value Vector4) bool {

	return v.X == value.X && v.Y == value.Y && v.Z == value.Z && v.W == value.W
}

////////////////////////////////////////////////////////////////////////////////

// Ne returns whether any component is not equal.
func (v Vector4) Ne(value Vector4) bool {

	return v.X != value.X || v.Y != value.Y || v.Z != value.Z || v.W != value.W
}
