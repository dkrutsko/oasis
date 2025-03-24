package math

import (
	"fmt"
	sysMath "math"
)

////////////////////////////////////////////////////////////////////////////////

var (
	Vector2Zero  = Vector2{0, 0}
	Vector2UnitX = Vector2{1, 0}
	Vector2UnitY = Vector2{0, 1}
)

////////////////////////////////////////////////////////////////////////////////

type Vector2 struct {
	X float64
	Y float64
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector2) String() string {

	return fmt.Sprintf("[%.2f, %.2f]", v.X, v.Y)
}

////////////////////////////////////////////////////////////////////////////////

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

func (v Vector2) Distance(value Vector2) float64 {

	dx := v.X - value.X
	dy := v.Y - value.Y

	return sysMath.Sqrt(dx*dx + dy*dy)
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector2) DistanceSq(value Vector2) float64 {

	dx := v.X - value.X
	dy := v.Y - value.Y

	return dx*dx + dy*dy
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector2) Dot(value Vector2) float64 {

	return v.X*value.X + v.Y*value.Y
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector2) Reflect(normal Vector2) Vector2 {

	dot := 2 * (v.X*normal.X + v.Y*normal.Y)

	return Vector2{
		v.X - dot*normal.X,
		v.Y - dot*normal.Y,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector2) Length() float64 {

	return sysMath.Sqrt(v.X*v.X + v.Y*v.Y)
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector2) LengthSq() float64 {

	return v.X*v.X + v.Y*v.Y
}

////////////////////////////////////////////////////////////////////////////////

func Vector2TransformVector2(matrix Matrix, value Vector2) Vector2 {

	return Vector2{
		matrix.M11*value.X + matrix.M21*value.Y + matrix.M41,
		matrix.M12*value.X + matrix.M22*value.Y + matrix.M42,
	}
}

////////////////////////////////////////////////////////////////////////////////

func Vector2TransformVector3(matrix Matrix, value Vector3) Vector2 {

	return Vector2{
		matrix.M11*value.X + matrix.M21*value.Y + matrix.M31*value.Z + matrix.M41,
		matrix.M12*value.X + matrix.M22*value.Y + matrix.M32*value.Z + matrix.M42,
	}
}

////////////////////////////////////////////////////////////////////////////////

func Vector2TransformVector4(matrix Matrix, value Vector4) Vector2 {

	return Vector2{
		matrix.M11*value.X + matrix.M21*value.Y + matrix.M31*value.Z + matrix.M41*value.W,
		matrix.M12*value.X + matrix.M22*value.Y + matrix.M32*value.Z + matrix.M42*value.W,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector2) Add(value Vector2) Vector2 {

	return Vector2{
		v.X + value.X,
		v.Y + value.Y,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector2) Sub(value Vector2) Vector2 {

	return Vector2{
		v.X - value.X,
		v.Y - value.Y,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector2) Mul(value Vector2) Vector2 {

	return Vector2{
		v.X * value.X,
		v.Y * value.Y,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector2) Div(value Vector2) Vector2 {

	return Vector2{
		v.X / value.X,
		v.Y / value.Y,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector2) AddScalar(scalar float64) Vector2 {

	return Vector2{
		v.X + scalar,
		v.Y + scalar,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector2) SubScalar(scalar float64) Vector2 {

	return Vector2{
		v.X - scalar,
		v.Y - scalar,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector2) MulScalar(scalar float64) Vector2 {

	return Vector2{
		v.X * scalar,
		v.Y * scalar,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector2) DivScalar(scalar float64) Vector2 {

	return Vector2{
		v.X / scalar,
		v.Y / scalar,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector2) Neg() Vector2 {

	return Vector2{
		-v.X,
		-v.Y,
	}
}

////////////////////////////////////////////////////////////////////////////////

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

func (v Vector2) Lt(value Vector2) bool {

	if v.X < value.X {
		return true
	}
	if v.X > value.X {
		return false
	}

	return v.Y < value.Y
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector2) Gt(value Vector2) bool {

	if v.X > value.X {
		return true
	}
	if v.X < value.X {
		return false
	}

	return v.Y > value.Y
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector2) Le(value Vector2) bool {

	if v.X < value.X {
		return true
	}
	if v.X > value.X {
		return false
	}

	return v.Y <= value.Y
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector2) Ge(value Vector2) bool {

	if v.X > value.X {
		return true
	}
	if v.X < value.X {
		return false
	}

	return v.Y >= value.Y
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector2) Eq(value Vector2) bool {

	return v.X == value.X && v.Y == value.Y
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector2) Ne(value Vector2) bool {

	return v.X != value.X || v.Y != value.Y
}
