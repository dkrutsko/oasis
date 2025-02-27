package math

import (
	"fmt"
	sysMath "math"
)

////////////////////////////////////////////////////////////////////////////////

var (
	Vector3Zero  = Vector3{0, 0, 0}
	Vector3UnitX = Vector3{1, 0, 0}
	Vector3UnitY = Vector3{0, 1, 0}
	Vector3UnitZ = Vector3{0, 0, 1}
)

////////////////////////////////////////////////////////////////////////////////

type Vector3 struct {
	X float64
	Y float64
	Z float64
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector3) String() string {

	return fmt.Sprintf("[%.2f, %.2f, %.2f]", v.X, v.Y, v.Z)
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

func (v Vector3) Eq(value Vector3) bool {

	return v.X == value.X && v.Y == value.Y && v.Z == value.Z
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector3) Ne(value Vector3) bool {

	return v.X != value.X || v.Y != value.Y || v.Z != value.Z
}
