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
	MatrixZero     = Matrix{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	MatrixIdentity = Matrix{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}
)

//----------------------------------------------------------------------------//
// Types                                                                      //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

type Matrix struct {
	M11, M12, M13, M14 float64
	M21, M22, M23, M24 float64
	M31, M32, M33, M34 float64
	M41, M42, M43, M44 float64
}

//----------------------------------------------------------------------------//
// Methods                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

func (m Matrix) String() string {

	return fmt.Sprintf(
		"["+
			"[%.2f, %.2f, %.2f, %.2f]"+
			"[%.2f, %.2f, %.2f, %.2f]"+
			"[%.2f, %.2f, %.2f, %.2f]"+
			"[%.2f, %.2f, %.2f, %.2f]"+
			"]",
		m.M11, m.M12, m.M13, m.M14,
		m.M21, m.M22, m.M23, m.M24,
		m.M31, m.M32, m.M33, m.M34,
		m.M41, m.M42, m.M43, m.M44,
	)
}

////////////////////////////////////////////////////////////////////////////////

func (m Matrix) IsZero() bool {

	return m.M11 == 0 &&
		m.M12 == 0 &&
		m.M13 == 0 &&
		m.M14 == 0 &&

		m.M21 == 0 &&
		m.M22 == 0 &&
		m.M23 == 0 &&
		m.M24 == 0 &&

		m.M31 == 0 &&
		m.M32 == 0 &&
		m.M33 == 0 &&
		m.M34 == 0 &&

		m.M41 == 0 &&
		m.M42 == 0 &&
		m.M43 == 0 &&
		m.M44 == 0
}

////////////////////////////////////////////////////////////////////////////////

func (m Matrix) Transpose() Matrix {

	return Matrix{
		m.M11, m.M21, m.M31, m.M41,
		m.M12, m.M22, m.M32, m.M42,
		m.M13, m.M23, m.M33, m.M43,
		m.M14, m.M24, m.M34, m.M44,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (m Matrix) Invert() Matrix {

	v01 := m.M11*m.M22 - m.M12*m.M21
	v02 := m.M11*m.M23 - m.M13*m.M21
	v03 := m.M11*m.M24 - m.M14*m.M21
	v04 := m.M12*m.M23 - m.M13*m.M22
	v05 := m.M12*m.M24 - m.M14*m.M22
	v06 := m.M13*m.M24 - m.M14*m.M23
	v07 := m.M31*m.M42 - m.M32*m.M41
	v08 := m.M31*m.M43 - m.M33*m.M41
	v09 := m.M31*m.M44 - m.M34*m.M41
	v10 := m.M32*m.M43 - m.M33*m.M42
	v11 := m.M32*m.M44 - m.M34*m.M42
	v12 := m.M33*m.M44 - m.M34*m.M43

	det := v01*v12 - v02*v11 + v03*v10 + v04*v09 - v05*v08 + v06*v07

	if det == 0 {
		return MatrixZero
	}

	return Matrix{
		(+m.M22*v12 - m.M23*v11 + m.M24*v10) / det,
		(-m.M12*v12 + m.M13*v11 - m.M14*v10) / det,
		(+m.M42*v06 - m.M43*v05 + m.M44*v04) / det,
		(-m.M32*v06 + m.M33*v05 - m.M34*v04) / det,

		(-m.M21*v12 + m.M23*v09 - m.M24*v08) / det,
		(+m.M11*v12 - m.M13*v09 + m.M14*v08) / det,
		(-m.M41*v06 + m.M43*v03 - m.M44*v02) / det,
		(+m.M31*v06 - m.M33*v03 + m.M34*v02) / det,

		(+m.M21*v11 - m.M22*v09 + m.M24*v07) / det,
		(-m.M11*v11 + m.M12*v09 - m.M14*v07) / det,
		(+m.M41*v05 - m.M42*v03 + m.M44*v01) / det,
		(-m.M31*v05 + m.M32*v03 - m.M34*v01) / det,

		(-m.M21*v10 + m.M22*v08 - m.M23*v07) / det,
		(+m.M11*v10 - m.M12*v08 + m.M13*v07) / det,
		(-m.M41*v04 + m.M42*v02 - m.M43*v01) / det,
		(+m.M31*v04 - m.M32*v02 + m.M33*v01) / det,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (m Matrix) Determinant() float64 {

	v1 := m.M33*m.M44 - m.M34*m.M43
	v2 := m.M32*m.M44 - m.M34*m.M42
	v3 := m.M32*m.M43 - m.M33*m.M42
	v4 := m.M31*m.M44 - m.M34*m.M41
	v5 := m.M31*m.M43 - m.M33*m.M41
	v6 := m.M31*m.M42 - m.M32*m.M41

	return (m.M11 * (m.M22*v1 - m.M23*v2 + m.M24*v3)) -
		(m.M12 * (m.M21*v1 - m.M23*v4 + m.M24*v5)) +
		(m.M13 * (m.M21*v2 - m.M22*v4 + m.M24*v6)) -
		(m.M14 * (m.M21*v3 - m.M22*v5 + m.M23*v6))
}

////////////////////////////////////////////////////////////////////////////////

func (m Matrix) GetUp() Vector3 {

	return Vector3{m.M21, m.M22, m.M23}
}

////////////////////////////////////////////////////////////////////////////////

func (m Matrix) SetUp(value Vector3) Matrix {

	m.M21 = value.X
	m.M22 = value.Y
	m.M23 = value.Z
	return m
}

////////////////////////////////////////////////////////////////////////////////

func (m Matrix) GetDown() Vector3 {

	return Vector3{-m.M21, -m.M22, -m.M23}
}

////////////////////////////////////////////////////////////////////////////////

func (m Matrix) SetDown(value Vector3) Matrix {

	m.M21 = -value.X
	m.M22 = -value.Y
	m.M23 = -value.Z
	return m
}

////////////////////////////////////////////////////////////////////////////////

func (m Matrix) GetRight() Vector3 {

	return Vector3{m.M11, m.M12, m.M13}
}

////////////////////////////////////////////////////////////////////////////////

func (m Matrix) SetRight(value Vector3) Matrix {

	m.M11 = value.X
	m.M12 = value.Y
	m.M13 = value.Z
	return m
}

////////////////////////////////////////////////////////////////////////////////

func (m Matrix) GetLeft() Vector3 {

	return Vector3{-m.M11, -m.M12, -m.M13}
}

////////////////////////////////////////////////////////////////////////////////

func (m Matrix) SetLeft(value Vector3) Matrix {

	m.M11 = -value.X
	m.M12 = -value.Y
	m.M13 = -value.Z
	return m
}

////////////////////////////////////////////////////////////////////////////////

func (m Matrix) GetForward() Vector3 {

	return Vector3{-m.M31, -m.M32, -m.M33}
}

////////////////////////////////////////////////////////////////////////////////

func (m Matrix) SetForward(value Vector3) Matrix {

	m.M31 = -value.X
	m.M32 = -value.Y
	m.M33 = -value.Z
	return m
}

////////////////////////////////////////////////////////////////////////////////

func (m Matrix) GetBackward() Vector3 {

	return Vector3{m.M31, m.M32, m.M33}
}

////////////////////////////////////////////////////////////////////////////////

func (m Matrix) SetBackward(value Vector3) Matrix {

	m.M31 = value.X
	m.M32 = value.Y
	m.M33 = value.Z
	return m
}

////////////////////////////////////////////////////////////////////////////////

func (m Matrix) GetTranslation() Vector3 {

	return Vector3{m.M41, m.M42, m.M43}
}

////////////////////////////////////////////////////////////////////////////////

func (m Matrix) SetTranslation(value Vector3) Matrix {

	m.M41 = value.X
	m.M42 = value.Y
	m.M43 = value.Z
	return m
}

////////////////////////////////////////////////////////////////////////////////

func (m Matrix) Lerp(target Matrix, amount float64) Matrix {

	return Matrix{
		m.M11 + (target.M11-m.M11)*amount,
		m.M12 + (target.M12-m.M12)*amount,
		m.M13 + (target.M13-m.M13)*amount,
		m.M14 + (target.M14-m.M14)*amount,

		m.M21 + (target.M21-m.M21)*amount,
		m.M22 + (target.M22-m.M22)*amount,
		m.M23 + (target.M23-m.M23)*amount,
		m.M24 + (target.M24-m.M24)*amount,

		m.M31 + (target.M31-m.M31)*amount,
		m.M32 + (target.M32-m.M32)*amount,
		m.M33 + (target.M33-m.M33)*amount,
		m.M34 + (target.M34-m.M34)*amount,

		m.M41 + (target.M41-m.M41)*amount,
		m.M42 + (target.M42-m.M42)*amount,
		m.M43 + (target.M43-m.M43)*amount,
		m.M44 + (target.M44-m.M44)*amount,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (m Matrix) ToSlice32() []float32 {

	return []float32{
		float32(m.M11), float32(m.M12), float32(m.M13), float32(m.M14),
		float32(m.M21), float32(m.M22), float32(m.M23), float32(m.M24),
		float32(m.M31), float32(m.M32), float32(m.M33), float32(m.M34),
		float32(m.M41), float32(m.M42), float32(m.M43), float32(m.M44),
	}
}

////////////////////////////////////////////////////////////////////////////////

func (m Matrix) ToSlice64() []float64 {

	return []float64{
		m.M11, m.M12, m.M13, m.M14,
		m.M21, m.M22, m.M23, m.M24,
		m.M31, m.M32, m.M33, m.M34,
		m.M41, m.M42, m.M43, m.M44,
	}
}

//----------------------------------------------------------------------------//
// Static                                                                     //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

func MatrixFromSlice32(values []float32) (Matrix, error) {

	if len(values) != 16 {
		return MatrixZero, errors.New("not enough values")
	}

	m := Matrix{
		float64(values[0]),  float64(values[1]),  float64(values[2]),  float64(values[3]),
		float64(values[4]),  float64(values[5]),  float64(values[6]),  float64(values[7]),
		float64(values[8]),  float64(values[9]),  float64(values[10]), float64(values[11]),
		float64(values[12]), float64(values[13]), float64(values[14]), float64(values[15]),
	}

	return m, nil
}

////////////////////////////////////////////////////////////////////////////////

func MatrixFromSlice64(values []float64) (Matrix, error) {

	if len(values) != 16 {
		return MatrixZero, errors.New("not enough values")
	}

	m := Matrix{
		values[0],  values[1],  values[2],  values[3],
		values[4],  values[5],  values[6],  values[7],
		values[8],  values[9],  values[10], values[11],
		values[12], values[13], values[14], values[15],
	}

	return m, nil
}

////////////////////////////////////////////////////////////////////////////////

func MatrixCreateProj(fov float64, width, height int, near, far float64) Matrix {

	// Do parameter check
	if fov <= 0 || fov >= sysMath.Pi || near <= 0 || far <= 0 || near >= far {
		return MatrixZero
	}

	w := float64(width)
	h := float64(height)

	dif := near - far
	cot := 1 / sysMath.Tan(fov*0.5)
	asp := cot / (w / h)

	m33 := far / dif
	m43 := near * far / dif

	return Matrix{
		asp, 0.0, 0.0, 0.0,
		0.0, cot, 0.0, 0.0,
		0.0, 0.0, m33, -1.0,
		0.0, 0.0, m43, 0.0,
	}
}

////////////////////////////////////////////////////////////////////////////////

func MatrixCreateView(pos, target, up Vector3) Matrix {

	vz := pos.Sub(target).Normalize()
	vx := up.Cross(vz).Normalize()
	vy := vz.Cross(vx).Normalize()

	return Matrix{
		vx.X, vy.X, vz.X, 0,
		vx.Y, vy.Y, vz.Y, 0,
		vx.Z, vy.Z, vz.Z, 0,
		-vx.Dot(pos),
		-vy.Dot(pos),
		-vz.Dot(pos), 1,
	}
}

////////////////////////////////////////////////////////////////////////////////

func MatrixProject(pos Vector3, width, height int, model, view, proj Matrix) Vector3 {

	return MatrixProjectWithMvp(pos, width, height, model.Mul(view).Mul(proj))
}

////////////////////////////////////////////////////////////////////////////////

func MatrixProjectWithMvp(pos Vector3, width, height int, mvp Matrix) Vector3 {

	const minZ = 0.0
	const maxZ = 1.0

	transform := Vector4TransformVector3(mvp, pos)

	return Vector3{
		(1 + transform.X / transform.W) * float64(width) * 0.5,
		(1 - transform.Y / transform.W) * float64(height) * 0.5,
		minZ + (transform.Z / transform.W) * (maxZ - minZ),
	}
}

////////////////////////////////////////////////////////////////////////////////

func MatrixCreateScale(x, y, z float64) Matrix {

	return Matrix{
		x, 0, 0, 0,
		0, y, 0, 0,
		0, 0, z, 0,
		0, 0, 0, 1,
	}
}

////////////////////////////////////////////////////////////////////////////////

func MatrixCreateRotationX(radians float64) Matrix {

	c := sysMath.Cos(radians)
	s := sysMath.Sin(radians)

	return Matrix{
		1, 0, 0, 0,
		0, c, s, 0,
		0, -s, c, 0,
		0, 0, 0, 1,
	}
}

////////////////////////////////////////////////////////////////////////////////

func MatrixCreateRotationY(radians float64) Matrix {

	c := sysMath.Cos(radians)
	s := sysMath.Sin(radians)

	return Matrix{
		c, 0, -s, 0,
		0, 1, 0, 0,
		s, 0, c, 0,
		0, 0, 0, 1,
	}
}

////////////////////////////////////////////////////////////////////////////////

func MatrixCreateRotationZ(radians float64) Matrix {

	c := sysMath.Cos(radians)
	s := sysMath.Sin(radians)

	return Matrix{
		c, s, 0, 0,
		-s, c, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

////////////////////////////////////////////////////////////////////////////////

func MatrixCreateTranslation(x, y, z float64) Matrix {

	return Matrix{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		x, y, z, 1,
	}
}

////////////////////////////////////////////////////////////////////////////////

func MatrixCreateFromAxisAngle(axis Vector3, angle float64) Matrix {

	x := axis.X
	y := axis.Y
	z := axis.Z

	s := sysMath.Sin(angle)
	c := sysMath.Cos(angle)

	xx := x * x
	yy := y * y
	zz := z * z
	xy := x * y
	xz := x * z
	yz := y * z

	return Matrix{
		xx + (1-xx)*c, xy - xy*c + z*s, xz - xz*c - y*s, 0,
		xy - xy*c - z*s, yy + (1-yy)*c, yz - yz*c + x*s, 0,
		xz - xz*c + y*s, yz - yz*c - x*s, zz + (1-zz)*c, 0,
		0, 0, 0, 1,
	}
}

////////////////////////////////////////////////////////////////////////////////

func MatrixCreateFromYawPitchRoll(yaw, pitch, roll float64) Matrix {

	return MatrixCreateRotationY(yaw).
		Mul(MatrixCreateRotationX(pitch)).
		Mul(MatrixCreateRotationZ(roll))
}

//----------------------------------------------------------------------------//
// Operators                                                                  //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

func (m Matrix) Add(value Matrix) Matrix {

	return Matrix{
		m.M11 + value.M11,
		m.M12 + value.M12,
		m.M13 + value.M13,
		m.M14 + value.M14,

		m.M21 + value.M21,
		m.M22 + value.M22,
		m.M23 + value.M23,
		m.M24 + value.M24,

		m.M31 + value.M31,
		m.M32 + value.M32,
		m.M33 + value.M33,
		m.M34 + value.M34,

		m.M41 + value.M41,
		m.M42 + value.M42,
		m.M43 + value.M43,
		m.M44 + value.M44,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (m Matrix) Sub(value Matrix) Matrix {

	return Matrix{
		m.M11 - value.M11,
		m.M12 - value.M12,
		m.M13 - value.M13,
		m.M14 - value.M14,

		m.M21 - value.M21,
		m.M22 - value.M22,
		m.M23 - value.M23,
		m.M24 - value.M24,

		m.M31 - value.M31,
		m.M32 - value.M32,
		m.M33 - value.M33,
		m.M34 - value.M34,

		m.M41 - value.M41,
		m.M42 - value.M42,
		m.M43 - value.M43,
		m.M44 - value.M44,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (m Matrix) Mul(value Matrix) Matrix {

	return Matrix{
		m.M11*value.M11 + m.M12*value.M21 + m.M13*value.M31 + m.M14*value.M41,
		m.M11*value.M12 + m.M12*value.M22 + m.M13*value.M32 + m.M14*value.M42,
		m.M11*value.M13 + m.M12*value.M23 + m.M13*value.M33 + m.M14*value.M43,
		m.M11*value.M14 + m.M12*value.M24 + m.M13*value.M34 + m.M14*value.M44,

		m.M21*value.M11 + m.M22*value.M21 + m.M23*value.M31 + m.M24*value.M41,
		m.M21*value.M12 + m.M22*value.M22 + m.M23*value.M32 + m.M24*value.M42,
		m.M21*value.M13 + m.M22*value.M23 + m.M23*value.M33 + m.M24*value.M43,
		m.M21*value.M14 + m.M22*value.M24 + m.M23*value.M34 + m.M24*value.M44,

		m.M31*value.M11 + m.M32*value.M21 + m.M33*value.M31 + m.M34*value.M41,
		m.M31*value.M12 + m.M32*value.M22 + m.M33*value.M32 + m.M34*value.M42,
		m.M31*value.M13 + m.M32*value.M23 + m.M33*value.M33 + m.M34*value.M43,
		m.M31*value.M14 + m.M32*value.M24 + m.M33*value.M34 + m.M34*value.M44,

		m.M41*value.M11 + m.M42*value.M21 + m.M43*value.M31 + m.M44*value.M41,
		m.M41*value.M12 + m.M42*value.M22 + m.M43*value.M32 + m.M44*value.M42,
		m.M41*value.M13 + m.M42*value.M23 + m.M43*value.M33 + m.M44*value.M43,
		m.M41*value.M14 + m.M42*value.M24 + m.M43*value.M34 + m.M44*value.M44,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (m Matrix) Div(value Matrix) Matrix {

	return Matrix{
		m.M11 / value.M11,
		m.M12 / value.M12,
		m.M13 / value.M13,
		m.M14 / value.M14,

		m.M21 / value.M21,
		m.M22 / value.M22,
		m.M23 / value.M23,
		m.M24 / value.M24,

		m.M31 / value.M31,
		m.M32 / value.M32,
		m.M33 / value.M33,
		m.M34 / value.M34,

		m.M41 / value.M41,
		m.M42 / value.M42,
		m.M43 / value.M43,
		m.M44 / value.M44,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (m Matrix) AddScalar(scalar float64) Matrix {

	return Matrix{
		m.M11 + scalar,
		m.M12 + scalar,
		m.M13 + scalar,
		m.M14 + scalar,

		m.M21 + scalar,
		m.M22 + scalar,
		m.M23 + scalar,
		m.M24 + scalar,

		m.M31 + scalar,
		m.M32 + scalar,
		m.M33 + scalar,
		m.M34 + scalar,

		m.M41 + scalar,
		m.M42 + scalar,
		m.M43 + scalar,
		m.M44 + scalar,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (m Matrix) SubScalar(scalar float64) Matrix {

	return Matrix{
		m.M11 - scalar,
		m.M12 - scalar,
		m.M13 - scalar,
		m.M14 - scalar,

		m.M21 - scalar,
		m.M22 - scalar,
		m.M23 - scalar,
		m.M24 - scalar,

		m.M31 - scalar,
		m.M32 - scalar,
		m.M33 - scalar,
		m.M34 - scalar,

		m.M41 - scalar,
		m.M42 - scalar,
		m.M43 - scalar,
		m.M44 - scalar,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (m Matrix) MulScalar(scalar float64) Matrix {

	return Matrix{
		m.M11 * scalar,
		m.M12 * scalar,
		m.M13 * scalar,
		m.M14 * scalar,

		m.M21 * scalar,
		m.M22 * scalar,
		m.M23 * scalar,
		m.M24 * scalar,

		m.M31 * scalar,
		m.M32 * scalar,
		m.M33 * scalar,
		m.M34 * scalar,

		m.M41 * scalar,
		m.M42 * scalar,
		m.M43 * scalar,
		m.M44 * scalar,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (m Matrix) DivScalar(scalar float64) Matrix {

	return Matrix{
		m.M11 / scalar,
		m.M12 / scalar,
		m.M13 / scalar,
		m.M14 / scalar,

		m.M21 / scalar,
		m.M22 / scalar,
		m.M23 / scalar,
		m.M24 / scalar,

		m.M31 / scalar,
		m.M32 / scalar,
		m.M33 / scalar,
		m.M34 / scalar,

		m.M41 / scalar,
		m.M42 / scalar,
		m.M43 / scalar,
		m.M44 / scalar,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (m Matrix) Neg() Matrix {

	return Matrix{
		-m.M11, -m.M12, -m.M13, -m.M14,
		-m.M21, -m.M22, -m.M23, -m.M24,
		-m.M31, -m.M32, -m.M33, -m.M34,
		-m.M41, -m.M42, -m.M43, -m.M44,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (m Matrix) Eq(value Matrix) bool {

	return m.M11 == value.M11 &&
		m.M12 == value.M12 &&
		m.M13 == value.M13 &&
		m.M14 == value.M14 &&

		m.M21 == value.M21 &&
		m.M22 == value.M22 &&
		m.M23 == value.M23 &&
		m.M24 == value.M24 &&

		m.M31 == value.M31 &&
		m.M32 == value.M32 &&
		m.M33 == value.M33 &&
		m.M34 == value.M34 &&

		m.M41 == value.M41 &&
		m.M42 == value.M42 &&
		m.M43 == value.M43 &&
		m.M44 == value.M44
}

////////////////////////////////////////////////////////////////////////////////

func (m Matrix) Ne(value Matrix) bool {

	return m.M11 != value.M11 ||
		m.M12 != value.M12 ||
		m.M13 != value.M13 ||
		m.M14 != value.M14 ||

		m.M21 != value.M21 ||
		m.M22 != value.M22 ||
		m.M23 != value.M23 ||
		m.M24 != value.M24 ||

		m.M31 != value.M31 ||
		m.M32 != value.M32 ||
		m.M33 != value.M33 ||
		m.M34 != value.M34 ||

		m.M41 != value.M41 ||
		m.M42 != value.M42 ||
		m.M43 != value.M43 ||
		m.M44 != value.M44
}
