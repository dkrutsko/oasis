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
	// Vector3Zero represents a zero vector.
	Vector3Zero = Vector3{0, 0, 0}

	// Vector3UnitX represents a unit vector along the x-axis.
	Vector3UnitX = Vector3{1, 0, 0}

	// Vector3UnitY represents a unit vector along the y-axis.
	Vector3UnitY = Vector3{0, 1, 0}

	// Vector3UnitZ represents a unit vector along the z-axis.
	Vector3UnitZ = Vector3{0, 0, 1}
)

//----------------------------------------------------------------------------//
// Types                                                                      //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// Vector3 represents a vector with three components.
type Vector3 struct {
	// X component of the vector.
	X float64

	// Y component of the vector.
	Y float64

	// Z component of the vector.
	Z float64
}

//----------------------------------------------------------------------------//
// Methods                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// String returns the string representation of the vector.
func (v Vector3) String() string {

	return fmt.Sprintf("[%.2f, %.2f, %.2f]", v.X, v.Y, v.Z)
}

////////////////////////////////////////////////////////////////////////////////

// IsZero returns whether all components are zero.
func (v Vector3) IsZero() bool {

	return v.X == 0 && v.Y == 0 && v.Z == 0
}

////////////////////////////////////////////////////////////////////////////////

// Normalize returns a unit vector in the same direction. Returns the zero
// vector if the magnitude is zero.
func (v Vector3) Normalize() Vector3 {

	magnitude := sysMath.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)

	if magnitude == 0 {
		return Vector3Zero
	}

	return Vector3{
		v.X / magnitude,
		v.Y / magnitude,
		v.Z / magnitude,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Distance returns the Euclidean distance to another vector.
func (v Vector3) Distance(value Vector3) float64 {

	dx := v.X - value.X
	dy := v.Y - value.Y
	dz := v.Z - value.Z

	return sysMath.Sqrt(dx*dx + dy*dy + dz*dz)
}

////////////////////////////////////////////////////////////////////////////////

// DistanceSq returns the squared Euclidean distance to another vector.
func (v Vector3) DistanceSq(value Vector3) float64 {

	dx := v.X - value.X
	dy := v.Y - value.Y
	dz := v.Z - value.Z

	return dx*dx + dy*dy + dz*dz
}

////////////////////////////////////////////////////////////////////////////////

// Dot returns the dot product with another vector.
func (v Vector3) Dot(value Vector3) float64 {

	return v.X*value.X + v.Y*value.Y + v.Z*value.Z
}

////////////////////////////////////////////////////////////////////////////////

// Reflect returns the reflection of the vector off a surface defined by the
// given normal.
func (v Vector3) Reflect(normal Vector3) Vector3 {

	dot := 2 * (v.X*normal.X + v.Y*normal.Y + v.Z*normal.Z)

	return Vector3{
		v.X - dot*normal.X,
		v.Y - dot*normal.Y,
		v.Z - dot*normal.Z,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Cross returns the cross product with another vector.
func (v Vector3) Cross(value Vector3) Vector3 {

	return Vector3{
		v.Y*value.Z - v.Z*value.Y,
		v.Z*value.X - v.X*value.Z,
		v.X*value.Y - v.Y*value.X,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Rotate returns the vector rotated around the given axis by the specified
// angle in radians using the Rodrigues rotation formula.
func (v Vector3) Rotate(axis Vector3, angle float64) Vector3 {

	sinAngle := sysMath.Sin(-angle)
	cosAngle := sysMath.Cos(-angle)

	crossTerm := v.Cross(axis.MulScalar(sinAngle))
	scaledVal := v.MulScalar(cosAngle)
	scaledAxis := axis.MulScalar(1 - cosAngle)
	dotTerm := axis.MulScalar(v.Dot(scaledAxis))

	return crossTerm.Add(scaledVal).Add(dotTerm)
}

////////////////////////////////////////////////////////////////////////////////

// Min returns the component-wise minimum of two vectors.
func (v Vector3) Min(value Vector3) Vector3 {

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

	return Vector3{x, y, z}
}

////////////////////////////////////////////////////////////////////////////////

// Max returns the component-wise maximum of two vectors.
func (v Vector3) Max(value Vector3) Vector3 {

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

	return Vector3{x, y, z}
}

////////////////////////////////////////////////////////////////////////////////

// Clamp restricts each component to the specified range.
func (v Vector3) Clamp(min, max Vector3) Vector3 {

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

	return Vector3{x, y, z}
}

////////////////////////////////////////////////////////////////////////////////

// Lerp performs linear interpolation toward the target.
func (v Vector3) Lerp(target Vector3, amount float64) Vector3 {

	return Vector3{
		v.X + (target.X-v.X)*amount,
		v.Y + (target.Y-v.Y)*amount,
		v.Z + (target.Z-v.Z)*amount,
	}
}

////////////////////////////////////////////////////////////////////////////////

// SmoothStep performs Hermite interpolation toward the target with smoothing
// at the edges.
func (v Vector3) SmoothStep(target Vector3, amount float64) Vector3 {

	amount = Clamp(amount, 0, 1)
	amount = amount * amount * (3 - 2*amount)

	return Vector3{
		v.X + (target.X-v.X)*amount,
		v.Y + (target.Y-v.Y)*amount,
		v.Z + (target.Z-v.Z)*amount,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Length returns the magnitude of the vector.
func (v Vector3) Length() float64 {

	return sysMath.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

////////////////////////////////////////////////////////////////////////////////

// LengthSq returns the squared magnitude of the vector.
func (v Vector3) LengthSq() float64 {

	return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

////////////////////////////////////////////////////////////////////////////////

// Angle returns the angle in radians between this vector and the given vector.
func (v Vector3) Angle(value Vector3) float64 {

	a := v.Normalize()
	b := value.Normalize()

	cosine := a.Dot(b)

	if cosine > 1 {
		return 0
	}
	if cosine < -1 {
		return sysMath.Pi
	}

	return sysMath.Acos(cosine)
}

////////////////////////////////////////////////////////////////////////////////

// ApplyQuaternion rotates the vector by the given quaternion.
func (v Vector3) ApplyQuaternion(q Quaternion) Vector3 {

	uvx := q.Y*v.Z - q.Z*v.Y
	uvy := q.Z*v.X - q.X*v.Z
	uvz := q.X*v.Y - q.Y*v.X

	uuvx := q.Y*uvz - q.Z*uvy
	uuvy := q.Z*uvx - q.X*uvz
	uuvz := q.X*uvy - q.Y*uvx

	w2 := q.W * 2

	return Vector3{
		v.X + uvx*w2 + uuvx*2,
		v.Y + uvy*w2 + uuvy*2,
		v.Z + uvz*w2 + uuvz*2,
	}
}

////////////////////////////////////////////////////////////////////////////////

// ToVector2 returns a Vector2 by dropping the Z component.
func (v Vector3) ToVector2() Vector2 {

	return Vector2{v.X, v.Y}
}

////////////////////////////////////////////////////////////////////////////////

// ToVector4 returns a Vector4 with the given W component.
func (v Vector3) ToVector4(w float64) Vector4 {

	return Vector4{v.X, v.Y, v.Z, w}
}

////////////////////////////////////////////////////////////////////////////////

// ToSlice32 returns the components as a float32 slice.
func (v Vector3) ToSlice32() []float32 {

	return []float32{
		float32(v.X),
		float32(v.Y),
		float32(v.Z),
	}
}

////////////////////////////////////////////////////////////////////////////////

// ToSlice64 returns the components as a float64 slice.
func (v Vector3) ToSlice64() []float64 {

	return []float64{
		v.X,
		v.Y,
		v.Z,
	}
}

//----------------------------------------------------------------------------//
// Static                                                                     //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// Vector3FromSlice32 creates a Vector3 from a float32 slice.
func Vector3FromSlice32(values []float32) (Vector3, error) {

	if len(values) != 3 {
		return Vector3Zero, ErrInvalidLength
	}

	v := Vector3{
		X: float64(values[0]),
		Y: float64(values[1]),
		Z: float64(values[2]),
	}

	return v, nil
}

////////////////////////////////////////////////////////////////////////////////

// Vector3FromSlice64 creates a Vector3 from a float64 slice.
func Vector3FromSlice64(values []float64) (Vector3, error) {

	if len(values) != 3 {
		return Vector3Zero, ErrInvalidLength
	}

	v := Vector3{
		X: values[0],
		Y: values[1],
		Z: values[2],
	}

	return v, nil
}

////////////////////////////////////////////////////////////////////////////////

// Vector3TransformVector2 transforms a Vector2 by a matrix and returns the
// resulting Vector3.
func Vector3TransformVector2(matrix Matrix4, value Vector2) Vector3 {

	return Vector3{
		matrix.M11*value.X + matrix.M21*value.Y + matrix.M41,
		matrix.M12*value.X + matrix.M22*value.Y + matrix.M42,
		matrix.M13*value.X + matrix.M23*value.Y + matrix.M43,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Vector3TransformVector3 transforms a Vector3 by a matrix and returns the
// resulting Vector3.
func Vector3TransformVector3(matrix Matrix4, value Vector3) Vector3 {

	return Vector3{
		matrix.M11*value.X + matrix.M21*value.Y + matrix.M31*value.Z + matrix.M41,
		matrix.M12*value.X + matrix.M22*value.Y + matrix.M32*value.Z + matrix.M42,
		matrix.M13*value.X + matrix.M23*value.Y + matrix.M33*value.Z + matrix.M43,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Vector3TransformVector4 transforms a Vector4 by a matrix and returns the
// resulting Vector3.
func Vector3TransformVector4(matrix Matrix4, value Vector4) Vector3 {

	return Vector3{
		matrix.M11*value.X + matrix.M21*value.Y + matrix.M31*value.Z + matrix.M41*value.W,
		matrix.M12*value.X + matrix.M22*value.Y + matrix.M32*value.Z + matrix.M42*value.W,
		matrix.M13*value.X + matrix.M23*value.Y + matrix.M33*value.Z + matrix.M43*value.W,
	}
}

//----------------------------------------------------------------------------//
// Operators                                                                  //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// Add returns the component-wise sum of two vectors.
func (v Vector3) Add(value Vector3) Vector3 {

	return Vector3{
		v.X + value.X,
		v.Y + value.Y,
		v.Z + value.Z,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Sub returns the component-wise difference of two vectors.
func (v Vector3) Sub(value Vector3) Vector3 {

	return Vector3{
		v.X - value.X,
		v.Y - value.Y,
		v.Z - value.Z,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Mul returns the component-wise product of two vectors.
func (v Vector3) Mul(value Vector3) Vector3 {

	return Vector3{
		v.X * value.X,
		v.Y * value.Y,
		v.Z * value.Z,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Div returns the component-wise quotient of two vectors.
func (v Vector3) Div(value Vector3) Vector3 {

	return Vector3{
		v.X / value.X,
		v.Y / value.Y,
		v.Z / value.Z,
	}
}

////////////////////////////////////////////////////////////////////////////////

// AddScalar adds a scalar to each component.
func (v Vector3) AddScalar(scalar float64) Vector3 {

	return Vector3{
		v.X + scalar,
		v.Y + scalar,
		v.Z + scalar,
	}
}

////////////////////////////////////////////////////////////////////////////////

// SubScalar subtracts a scalar from each component.
func (v Vector3) SubScalar(scalar float64) Vector3 {

	return Vector3{
		v.X - scalar,
		v.Y - scalar,
		v.Z - scalar,
	}
}

////////////////////////////////////////////////////////////////////////////////

// MulScalar multiplies each component by a scalar.
func (v Vector3) MulScalar(scalar float64) Vector3 {

	return Vector3{
		v.X * scalar,
		v.Y * scalar,
		v.Z * scalar,
	}
}

////////////////////////////////////////////////////////////////////////////////

// DivScalar divides each component by a scalar.
func (v Vector3) DivScalar(scalar float64) Vector3 {

	return Vector3{
		v.X / scalar,
		v.Y / scalar,
		v.Z / scalar,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Neg returns the negation of the vector.
func (v Vector3) Neg() Vector3 {

	return Vector3{
		-v.X,
		-v.Y,
		-v.Z,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Compare performs lexicographic comparison of two vectors. Returns -1, 0, or 1.
func (v Vector3) Compare(value Vector3) int {

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

	return 0
}

////////////////////////////////////////////////////////////////////////////////

// Lt returns whether the vector is lexicographically less than the given vector.
func (v Vector3) Lt(value Vector3) bool {
	return v.Compare(value) < 0
}

////////////////////////////////////////////////////////////////////////////////

// Gt returns whether the vector is lexicographically greater than the given vector.
func (v Vector3) Gt(value Vector3) bool {
	return v.Compare(value) > 0
}

////////////////////////////////////////////////////////////////////////////////

// Le returns whether the vector is lexicographically less than or equal to
// the given vector.
func (v Vector3) Le(value Vector3) bool {
	return v.Compare(value) <= 0
}

////////////////////////////////////////////////////////////////////////////////

// Ge returns whether the vector is lexicographically greater than or equal to
// the given vector.
func (v Vector3) Ge(value Vector3) bool {
	return v.Compare(value) >= 0
}

////////////////////////////////////////////////////////////////////////////////

// Eq returns whether all components are equal.
func (v Vector3) Eq(value Vector3) bool {

	return v.X == value.X && v.Y == value.Y && v.Z == value.Z
}

////////////////////////////////////////////////////////////////////////////////

// Ne returns whether any component is not equal.
func (v Vector3) Ne(value Vector3) bool {

	return v.X != value.X || v.Y != value.Y || v.Z != value.Z
}
