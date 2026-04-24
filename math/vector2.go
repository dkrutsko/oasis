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
	// Vector2Zero represents a zero vector.
	Vector2Zero = Vector2{0, 0}

	// Vector2UnitX represents a unit vector along the x-axis.
	Vector2UnitX = Vector2{1, 0}

	// Vector2UnitY represents a unit vector along the y-axis.
	Vector2UnitY = Vector2{0, 1}
)

//----------------------------------------------------------------------------//
// Types                                                                      //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// Vector2 represents a vector with two components.
type Vector2 struct {
	// X component of the vector.
	X float64

	// Y component of the vector.
	Y float64
}

//----------------------------------------------------------------------------//
// Methods                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// String returns the string representation of the vector.
func (v Vector2) String() string {

	return fmt.Sprintf("[%.2f, %.2f]", v.X, v.Y)
}

////////////////////////////////////////////////////////////////////////////////

// IsZero returns whether all components are zero.
func (v Vector2) IsZero() bool {

	return v.X == 0 && v.Y == 0
}

////////////////////////////////////////////////////////////////////////////////

// Normalize returns a unit vector in the same direction. Returns `Vector2Zero`
// if the magnitude is zero.
func (v Vector2) Normalize() Vector2 {

	magnitude := sysMath.Sqrt(v.X*v.X + v.Y*v.Y)

	// The default case
	if magnitude == 0 {
		return Vector2Zero
	}

	return Vector2{
		v.X / magnitude,
		v.Y / magnitude,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Distance returns the Euclidean distance to another vector.
func (v Vector2) Distance(value Vector2) float64 {

	dx := v.X - value.X
	dy := v.Y - value.Y

	return sysMath.Sqrt(dx*dx + dy*dy)
}

////////////////////////////////////////////////////////////////////////////////

// DistanceSq returns the squared Euclidean distance to another vector.
func (v Vector2) DistanceSq(value Vector2) float64 {

	dx := v.X - value.X
	dy := v.Y - value.Y

	return dx*dx + dy*dy
}

////////////////////////////////////////////////////////////////////////////////

// Dot returns the dot product with another vector.
func (v Vector2) Dot(value Vector2) float64 {

	return v.X*value.X + v.Y*value.Y
}

////////////////////////////////////////////////////////////////////////////////

// Reflect returns the reflection of the vector off a surface defined by the
// given normal.
func (v Vector2) Reflect(normal Vector2) Vector2 {

	dot := 2 * (v.X*normal.X + v.Y*normal.Y)

	return Vector2{
		v.X - dot*normal.X,
		v.Y - dot*normal.Y,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Min returns the component-wise minimum of two vectors.
func (v Vector2) Min(value Vector2) Vector2 {

	x := v.X
	if value.X < x {
		x = value.X
	}

	y := v.Y
	if value.Y < y {
		y = value.Y
	}

	return Vector2{x, y}
}

////////////////////////////////////////////////////////////////////////////////

// Max returns the component-wise maximum of two vectors.
func (v Vector2) Max(value Vector2) Vector2 {

	x := v.X
	if value.X > x {
		x = value.X
	}

	y := v.Y
	if value.Y > y {
		y = value.Y
	}

	return Vector2{x, y}
}

////////////////////////////////////////////////////////////////////////////////

// Clamp restricts each component to the specified range.
func (v Vector2) Clamp(min, max Vector2) Vector2 {

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

	return Vector2{x, y}
}

////////////////////////////////////////////////////////////////////////////////

// Lerp performs linear interpolation toward the target.
func (v Vector2) Lerp(target Vector2, amount float64) Vector2 {

	return Vector2{
		v.X + (target.X-v.X)*amount,
		v.Y + (target.Y-v.Y)*amount,
	}
}

////////////////////////////////////////////////////////////////////////////////

// SmoothStep performs Hermite interpolation toward the target with smoothing
// at the edges.
func (v Vector2) SmoothStep(target Vector2, amount float64) Vector2 {

	amount = Clamp(amount, 0, 1)
	amount = amount * amount * (3 - 2*amount)

	return Vector2{
		v.X + (target.X-v.X)*amount,
		v.Y + (target.Y-v.Y)*amount,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Length returns the magnitude of the vector.
func (v Vector2) Length() float64 {

	return sysMath.Sqrt(v.X*v.X + v.Y*v.Y)
}

////////////////////////////////////////////////////////////////////////////////

// LengthSq returns the squared magnitude of the vector.
func (v Vector2) LengthSq() float64 {

	return v.X*v.X + v.Y*v.Y
}

////////////////////////////////////////////////////////////////////////////////

// ToSlice32 returns the components as a float32 slice.
func (v Vector2) ToSlice32() []float32 {

	return []float32{
		float32(v.X),
		float32(v.Y),
	}
}

////////////////////////////////////////////////////////////////////////////////

// ToSlice64 returns the components as a float64 slice.
func (v Vector2) ToSlice64() []float64 {

	return []float64{
		v.X,
		v.Y,
	}
}

//----------------------------------------------------------------------------//
// Static                                                                     //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// Vector2FromSlice32 creates a Vector2 from a float32 slice.
func Vector2FromSlice32(values []float32) (Vector2, error) {

	if len(values) != 2 {
		return Vector2Zero, errors.New("not enough values")
	}

	v := Vector2{
		X: float64(values[0]),
		Y: float64(values[1]),
	}

	return v, nil
}

////////////////////////////////////////////////////////////////////////////////

// Vector2FromSlice64 creates a Vector2 from a float64 slice.
func Vector2FromSlice64(values []float64) (Vector2, error) {

	if len(values) != 2 {
		return Vector2Zero, errors.New("not enough values")
	}

	v := Vector2{
		X: values[0],
		Y: values[1],
	}

	return v, nil
}

////////////////////////////////////////////////////////////////////////////////

// Vector2TransformVector2 transforms a Vector2 by a matrix and returns the
// resulting Vector2.
func Vector2TransformVector2(matrix Matrix, value Vector2) Vector2 {

	return Vector2{
		matrix.M11*value.X + matrix.M21*value.Y + matrix.M41,
		matrix.M12*value.X + matrix.M22*value.Y + matrix.M42,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Vector2TransformVector3 transforms a Vector3 by a matrix and returns the
// resulting Vector2.
func Vector2TransformVector3(matrix Matrix, value Vector3) Vector2 {

	return Vector2{
		matrix.M11*value.X + matrix.M21*value.Y + matrix.M31*value.Z + matrix.M41,
		matrix.M12*value.X + matrix.M22*value.Y + matrix.M32*value.Z + matrix.M42,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Vector2TransformVector4 transforms a Vector4 by a matrix and returns the
// resulting Vector2.
func Vector2TransformVector4(matrix Matrix, value Vector4) Vector2 {

	return Vector2{
		matrix.M11*value.X + matrix.M21*value.Y + matrix.M31*value.Z + matrix.M41*value.W,
		matrix.M12*value.X + matrix.M22*value.Y + matrix.M32*value.Z + matrix.M42*value.W,
	}
}

//----------------------------------------------------------------------------//
// Operators                                                                  //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// Add returns the component-wise sum of two vectors.
func (v Vector2) Add(value Vector2) Vector2 {

	return Vector2{
		v.X + value.X,
		v.Y + value.Y,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Sub returns the component-wise difference of two vectors.
func (v Vector2) Sub(value Vector2) Vector2 {

	return Vector2{
		v.X - value.X,
		v.Y - value.Y,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Mul returns the component-wise product of two vectors.
func (v Vector2) Mul(value Vector2) Vector2 {

	return Vector2{
		v.X * value.X,
		v.Y * value.Y,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Div returns the component-wise quotient of two vectors.
func (v Vector2) Div(value Vector2) Vector2 {

	return Vector2{
		v.X / value.X,
		v.Y / value.Y,
	}
}

////////////////////////////////////////////////////////////////////////////////

// AddScalar adds a scalar to each component.
func (v Vector2) AddScalar(scalar float64) Vector2 {

	return Vector2{
		v.X + scalar,
		v.Y + scalar,
	}
}

////////////////////////////////////////////////////////////////////////////////

// SubScalar subtracts a scalar from each component.
func (v Vector2) SubScalar(scalar float64) Vector2 {

	return Vector2{
		v.X - scalar,
		v.Y - scalar,
	}
}

////////////////////////////////////////////////////////////////////////////////

// MulScalar multiplies each component by a scalar.
func (v Vector2) MulScalar(scalar float64) Vector2 {

	return Vector2{
		v.X * scalar,
		v.Y * scalar,
	}
}

////////////////////////////////////////////////////////////////////////////////

// DivScalar divides each component by a scalar.
func (v Vector2) DivScalar(scalar float64) Vector2 {

	return Vector2{
		v.X / scalar,
		v.Y / scalar,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Neg returns the negation of the vector.
func (v Vector2) Neg() Vector2 {

	return Vector2{
		-v.X,
		-v.Y,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Compare performs lexicographic comparison of two vectors. Returns -1, 0, or 1.
func (v Vector2) Compare(value Vector2) int {

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

	return 0
}

////////////////////////////////////////////////////////////////////////////////

// Lt returns whether the vector is lexicographically less than the given vector.
func (v Vector2) Lt(value Vector2) bool {
	return v.Compare(value) < 0
}

////////////////////////////////////////////////////////////////////////////////

// Gt returns whether the vector is lexicographically greater than the given vector.
func (v Vector2) Gt(value Vector2) bool {
	return v.Compare(value) > 0
}

////////////////////////////////////////////////////////////////////////////////

// Le returns whether the vector is lexicographically less than or equal to
// the given vector.
func (v Vector2) Le(value Vector2) bool {
	return v.Compare(value) <= 0
}

////////////////////////////////////////////////////////////////////////////////

// Ge returns whether the vector is lexicographically greater than or equal to
// the given vector.
func (v Vector2) Ge(value Vector2) bool {
	return v.Compare(value) >= 0
}

////////////////////////////////////////////////////////////////////////////////

// Eq returns whether all components are equal.
func (v Vector2) Eq(value Vector2) bool {

	return v.X == value.X && v.Y == value.Y
}

////////////////////////////////////////////////////////////////////////////////

// Ne returns whether any component is not equal.
func (v Vector2) Ne(value Vector2) bool {

	return v.X != value.X || v.Y != value.Y
}
