package math

import (
	"fmt"
	sysMath "math"
)

////////////////////////////////////////////////////////////////////////////////

var (
	Vector4Zero     = Vector4{0, 0, 0, 0}
	Vector4UnitX    = Vector4{1, 0, 0, 0}
	Vector4UnitY    = Vector4{0, 1, 0, 0}
	Vector4UnitZ    = Vector4{0, 0, 1, 0}
	Vector4UnitW    = Vector4{0, 0, 0, 1}
	Vector4Identity = Vector4{0, 0, 0, 1}
)

////////////////////////////////////////////////////////////////////////////////

type Vector4 struct {
	X float64
	Y float64
	Z float64
	W float64
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector4) String() string {

	return fmt.Sprintf("[%.2f, %.2f, %.2f, %.2f]", v.X, v.Y, v.Z, v.W)
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector4) Normalize() Vector4 {

	magnitude := sysMath.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z + v.W*v.W)

	// The default case
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

func (v Vector4) Distance(value Vector4) float64 {

	dx := v.X - value.X
	dy := v.Y - value.Y
	dz := v.Z - value.Z
	dw := v.W - value.W

	return sysMath.Sqrt(dx*dx + dy*dy + dz*dz + dw*dw)
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector4) DistanceSq(value Vector4) float64 {

	dx := v.X - value.X
	dy := v.Y - value.Y
	dz := v.Z - value.Z
	dw := v.W - value.W

	return dx*dx + dy*dy + dz*dz + dw*dw
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector4) Dot(value Vector4) float64 {

	return v.X*value.X + v.Y*value.Y + v.Z*value.Z + v.W*value.W
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector4) Length() float64 {

	return sysMath.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z + v.W*v.W)
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector4) LengthSq() float64 {

	return v.X*v.X + v.Y*v.Y + v.Z*v.Z + v.W*v.W
}

////////////////////////////////////////////////////////////////////////////////

func Vector4TransformVector2(matrix Matrix, value Vector2) Vector4 {

	return Vector4{
		matrix.M11*value.X + matrix.M21*value.Y + matrix.M41,
		matrix.M12*value.X + matrix.M22*value.Y + matrix.M42,
		matrix.M13*value.X + matrix.M23*value.Y + matrix.M43,
		matrix.M14*value.X + matrix.M24*value.Y + matrix.M44,
	}
}

////////////////////////////////////////////////////////////////////////////////

func Vector4TransformVector3(matrix Matrix, value Vector3) Vector4 {

	return Vector4{
		matrix.M11*value.X + matrix.M21*value.Y + matrix.M31*value.Z + matrix.M41,
		matrix.M12*value.X + matrix.M22*value.Y + matrix.M32*value.Z + matrix.M42,
		matrix.M13*value.X + matrix.M23*value.Y + matrix.M33*value.Z + matrix.M43,
		matrix.M14*value.X + matrix.M24*value.Y + matrix.M34*value.Z + matrix.M44,
	}
}

////////////////////////////////////////////////////////////////////////////////

func Vector4TransformVector4(matrix Matrix, value Vector4) Vector4 {

	return Vector4{
		matrix.M11*value.X + matrix.M21*value.Y + matrix.M31*value.Z + matrix.M41*value.W,
		matrix.M12*value.X + matrix.M22*value.Y + matrix.M32*value.Z + matrix.M42*value.W,
		matrix.M13*value.X + matrix.M23*value.Y + matrix.M33*value.Z + matrix.M43*value.W,
		matrix.M14*value.X + matrix.M24*value.Y + matrix.M34*value.Z + matrix.M44*value.W,
	}
}

////////////////////////////////////////////////////////////////////////////////

func Vector4FromPacked(packed int64) Vector4 {

	const a = 1 / 2097152.0
	const b = 1 / 1048576.0

	x := float64((packed<<0)>>42) * a
	y := float64((packed<<22)>>43) * b
	z := float64((packed<<43)>>43) * b

	wSquared := x*x + y*y + z*z

	w := 0.0
	if sysMath.Abs(wSquared-1) >= b {
		w = sysMath.Sqrt(1 - wSquared)
	}

	return Vector4{x, y, z, w}
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector4) Add(value Vector4) Vector4 {

	return Vector4{
		v.X + value.X,
		v.Y + value.Y,
		v.Z + value.Z,
		v.W + value.W,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector4) Sub(value Vector4) Vector4 {

	return Vector4{
		v.X - value.X,
		v.Y - value.Y,
		v.Z - value.Z,
		v.W - value.W,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector4) Mul(value Vector4) Vector4 {

	return Vector4{
		v.X * value.X,
		v.Y * value.Y,
		v.Z * value.Z,
		v.W * value.W,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector4) Div(value Vector4) Vector4 {

	return Vector4{
		v.X / value.X,
		v.Y / value.Y,
		v.Z / value.Z,
		v.W / value.W,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector4) AddScalar(scalar float64) Vector4 {

	return Vector4{
		v.X + scalar,
		v.Y + scalar,
		v.Z + scalar,
		v.W + scalar,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector4) SubScalar(scalar float64) Vector4 {

	return Vector4{
		v.X - scalar,
		v.Y - scalar,
		v.Z - scalar,
		v.W - scalar,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector4) MulScalar(scalar float64) Vector4 {

	return Vector4{
		v.X * scalar,
		v.Y * scalar,
		v.Z * scalar,
		v.W * scalar,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector4) DivScalar(scalar float64) Vector4 {

	return Vector4{
		v.X / scalar,
		v.Y / scalar,
		v.Z / scalar,
		v.W / scalar,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector4) Neg() Vector4 {

	return Vector4{
		-v.X,
		-v.Y,
		-v.Z,
		-v.W,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector4) Eq(value Vector4) bool {

	return v.X == value.X && v.Y == value.Y && v.Z == value.Z && v.W == value.W
}

////////////////////////////////////////////////////////////////////////////////

func (v Vector4) Ne(value Vector4) bool {

	return v.X != value.X || v.Y != value.Y || v.Z != value.Z || v.W != value.W
}
