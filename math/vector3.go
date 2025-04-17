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
	Vector3Zero  = Vector3{0, 0, 0}
	Vector3UnitX = Vector3{1, 0, 0}
	Vector3UnitY = Vector3{0, 1, 0}
	Vector3UnitZ = Vector3{0, 0, 1}
)

//----------------------------------------------------------------------------//
// Types                                                                      //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

type Vector3 struct {
	X float64
	Y float64
	Z float64
}

//----------------------------------------------------------------------------//
// Methods                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

func (v Vector3) String() string {

	return fmt.Sprintf("[%.2f, %.2f, %.2f]", v.X, v.Y, v.Z)
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector3) IsZero() bool {

	return v.X == 0 && v.Y == 0 && v.Z == 0
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector3) Normalize() Vector3 {

	magnitude := sysMath.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)

	// The default case
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

func (v Vector3) Distance(value Vector3) float64 {

	dx := v.X - value.X
	dy := v.Y - value.Y
	dz := v.Z - value.Z

	return sysMath.Sqrt(dx*dx + dy*dy + dz*dz)
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector3) DistanceSq(value Vector3) float64 {

	dx := v.X - value.X
	dy := v.Y - value.Y
	dz := v.Z - value.Z

	return dx*dx + dy*dy + dz*dz
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector3) Dot(value Vector3) float64 {

	return v.X*value.X + v.Y*value.Y + v.Z*value.Z
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector3) Reflect(normal Vector3) Vector3 {

	dot := 2 * (v.X*normal.X + v.Y*normal.Y + v.Z*normal.Z)

	return Vector3{
		v.X - dot*normal.X,
		v.Y - dot*normal.Y,
		v.Z - dot*normal.Z,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector3) Cross(value Vector3) Vector3 {

	return Vector3{
		v.Y*value.Z - v.Z*value.Y,
		v.Z*value.X - v.X*value.Z,
		v.X*value.Y - v.Y*value.X,
	}
}

////////////////////////////////////////////////////////////////////////////////

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

func (v Vector3) Length() float64 {

	return sysMath.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector3) LengthSq() float64 {

	return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector3) ToSlice32() []float32 {

	return []float32{
		float32(v.X),
		float32(v.Y),
		float32(v.Z),
	}
}

////////////////////////////////////////////////////////////////////////////////

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

func Vector3FromSlice32(values []float32) (Vector3, error) {

	if len(values) != 3 {
		return Vector3Zero, errors.New("not enough values")
	}

	v := Vector3{
		X: float64(values[0]),
		Y: float64(values[1]),
		Z: float64(values[2]),
	}

	return v, nil
}

////////////////////////////////////////////////////////////////////////////////

func Vector3FromSlice64(values []float64) (Vector3, error) {

	if len(values) != 3 {
		return Vector3Zero, errors.New("not enough values")
	}

	v := Vector3{
		X: values[0],
		Y: values[1],
		Z: values[2],
	}

	return v, nil
}

////////////////////////////////////////////////////////////////////////////////

func Vector3TransformVector2(matrix Matrix, value Vector2) Vector3 {

	return Vector3{
		matrix.M11*value.X + matrix.M21*value.Y + matrix.M41,
		matrix.M12*value.X + matrix.M22*value.Y + matrix.M42,
		matrix.M13*value.X + matrix.M23*value.Y + matrix.M43,
	}
}

////////////////////////////////////////////////////////////////////////////////

func Vector3TransformVector3(matrix Matrix, value Vector3) Vector3 {

	return Vector3{
		matrix.M11*value.X + matrix.M21*value.Y + matrix.M31*value.Z + matrix.M41,
		matrix.M12*value.X + matrix.M22*value.Y + matrix.M32*value.Z + matrix.M42,
		matrix.M13*value.X + matrix.M23*value.Y + matrix.M33*value.Z + matrix.M43,
	}
}

////////////////////////////////////////////////////////////////////////////////

func Vector3TransformVector4(matrix Matrix, value Vector4) Vector3 {

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

func (v Vector3) Add(value Vector3) Vector3 {

	return Vector3{
		v.X + value.X,
		v.Y + value.Y,
		v.Z + value.Z,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector3) Sub(value Vector3) Vector3 {

	return Vector3{
		v.X - value.X,
		v.Y - value.Y,
		v.Z - value.Z,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector3) Mul(value Vector3) Vector3 {

	return Vector3{
		v.X * value.X,
		v.Y * value.Y,
		v.Z * value.Z,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector3) Div(value Vector3) Vector3 {

	return Vector3{
		v.X / value.X,
		v.Y / value.Y,
		v.Z / value.Z,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector3) AddScalar(scalar float64) Vector3 {

	return Vector3{
		v.X + scalar,
		v.Y + scalar,
		v.Z + scalar,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector3) SubScalar(scalar float64) Vector3 {

	return Vector3{
		v.X - scalar,
		v.Y - scalar,
		v.Z - scalar,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector3) MulScalar(scalar float64) Vector3 {

	return Vector3{
		v.X * scalar,
		v.Y * scalar,
		v.Z * scalar,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector3) DivScalar(scalar float64) Vector3 {

	return Vector3{
		v.X / scalar,
		v.Y / scalar,
		v.Z / scalar,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector3) Neg() Vector3 {

	return Vector3{
		-v.X,
		-v.Y,
		-v.Z,
	}
}

////////////////////////////////////////////////////////////////////////////////

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

func (v Vector3) Lt(value Vector3) bool {
	return v.Compare(value) < 0
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector3) Gt(value Vector3) bool {
	return v.Compare(value) > 0
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector3) Le(value Vector3) bool {
	return v.Compare(value) <= 0
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector3) Ge(value Vector3) bool {
	return v.Compare(value) >= 0
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector3) Eq(value Vector3) bool {

	return v.X == value.X && v.Y == value.Y && v.Z == value.Z
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector3) Ne(value Vector3) bool {

	return v.X != value.X || v.Y != value.Y || v.Z != value.Z
}
