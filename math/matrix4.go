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
	// Matrix4Zero represents a matrix with all elements set to zero.
	Matrix4Zero = Matrix4{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	// Matrix4Identity represents the identity matrix.
	Matrix4Identity = Matrix4{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}
)

//----------------------------------------------------------------------------//
// Types                                                                      //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// Matrix4 represents a 4x4 transformation matrix stored in row-major order.
type Matrix4 struct {
	M11, M12, M13, M14 float64
	M21, M22, M23, M24 float64
	M31, M32, M33, M34 float64
	M41, M42, M43, M44 float64
}

//----------------------------------------------------------------------------//
// Methods                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// String returns the string representation of the matrix.
func (m Matrix4) String() string {

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

// IsZero returns whether all elements are zero.
func (m Matrix4) IsZero() bool {

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

// Transpose returns the transpose of the matrix.
func (m Matrix4) Transpose() Matrix4 {

	return Matrix4{
		m.M11, m.M21, m.M31, m.M41,
		m.M12, m.M22, m.M32, m.M42,
		m.M13, m.M23, m.M33, m.M43,
		m.M14, m.M24, m.M34, m.M44,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Invert returns the inverse of the matrix. Returns `Matrix4Zero` if the
// matrix is singular.
func (m Matrix4) Invert() Matrix4 {

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
		return Matrix4Zero
	}

	return Matrix4{
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

// Determinant returns the determinant of the matrix.
func (m Matrix4) Determinant() float64 {

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

// GetUp returns the up direction vector from the matrix.
func (m Matrix4) GetUp() Vector3 {

	return Vector3{m.M21, m.M22, m.M23}
}

////////////////////////////////////////////////////////////////////////////////

// SetUp returns the matrix with the up direction set to the given vector.
func (m Matrix4) SetUp(value Vector3) Matrix4 {

	return Matrix4{
		m.M11, m.M12, m.M13, m.M14,
		value.X, value.Y, value.Z, m.M24,
		m.M31, m.M32, m.M33, m.M34,
		m.M41, m.M42, m.M43, m.M44,
	}
}

////////////////////////////////////////////////////////////////////////////////

// GetDown returns the down direction vector from the matrix.
func (m Matrix4) GetDown() Vector3 {

	return Vector3{-m.M21, -m.M22, -m.M23}
}

////////////////////////////////////////////////////////////////////////////////

// SetDown returns the matrix with the down direction set to the given vector.
func (m Matrix4) SetDown(value Vector3) Matrix4 {

	return Matrix4{
		m.M11, m.M12, m.M13, m.M14,
		-value.X, -value.Y, -value.Z, m.M24,
		m.M31, m.M32, m.M33, m.M34,
		m.M41, m.M42, m.M43, m.M44,
	}
}

////////////////////////////////////////////////////////////////////////////////

// GetRight returns the right direction vector from the matrix.
func (m Matrix4) GetRight() Vector3 {

	return Vector3{m.M11, m.M12, m.M13}
}

////////////////////////////////////////////////////////////////////////////////

// SetRight returns the matrix with the right direction set to the given vector.
func (m Matrix4) SetRight(value Vector3) Matrix4 {

	return Matrix4{
		value.X, value.Y, value.Z, m.M14,
		m.M21, m.M22, m.M23, m.M24,
		m.M31, m.M32, m.M33, m.M34,
		m.M41, m.M42, m.M43, m.M44,
	}
}

////////////////////////////////////////////////////////////////////////////////

// GetLeft returns the left direction vector from the matrix.
func (m Matrix4) GetLeft() Vector3 {

	return Vector3{-m.M11, -m.M12, -m.M13}
}

////////////////////////////////////////////////////////////////////////////////

// SetLeft returns the matrix with the left direction set to the given vector.
func (m Matrix4) SetLeft(value Vector3) Matrix4 {

	return Matrix4{
		-value.X, -value.Y, -value.Z, m.M14,
		m.M21, m.M22, m.M23, m.M24,
		m.M31, m.M32, m.M33, m.M34,
		m.M41, m.M42, m.M43, m.M44,
	}
}

////////////////////////////////////////////////////////////////////////////////

// GetForward returns the forward direction vector from the matrix, negated
// for right-handed coordinates.
func (m Matrix4) GetForward() Vector3 {

	return Vector3{-m.M31, -m.M32, -m.M33}
}

////////////////////////////////////////////////////////////////////////////////

// SetForward returns the matrix with the forward direction set to the given
// vector, negated for right-handed coordinates.
func (m Matrix4) SetForward(value Vector3) Matrix4 {

	return Matrix4{
		m.M11, m.M12, m.M13, m.M14,
		m.M21, m.M22, m.M23, m.M24,
		-value.X, -value.Y, -value.Z, m.M34,
		m.M41, m.M42, m.M43, m.M44,
	}
}

////////////////////////////////////////////////////////////////////////////////

// GetBackward returns the backward direction vector from the matrix.
func (m Matrix4) GetBackward() Vector3 {

	return Vector3{m.M31, m.M32, m.M33}
}

////////////////////////////////////////////////////////////////////////////////

// SetBackward returns the matrix with the backward direction set to the given
// vector.
func (m Matrix4) SetBackward(value Vector3) Matrix4 {

	return Matrix4{
		m.M11, m.M12, m.M13, m.M14,
		m.M21, m.M22, m.M23, m.M24,
		value.X, value.Y, value.Z, m.M34,
		m.M41, m.M42, m.M43, m.M44,
	}
}

////////////////////////////////////////////////////////////////////////////////

// GetTranslation returns the translation vector from the matrix.
func (m Matrix4) GetTranslation() Vector3 {

	return Vector3{m.M41, m.M42, m.M43}
}

////////////////////////////////////////////////////////////////////////////////

// SetTranslation returns the matrix with the translation set to the given
// vector.
func (m Matrix4) SetTranslation(value Vector3) Matrix4 {

	return Matrix4{
		m.M11, m.M12, m.M13, m.M14,
		m.M21, m.M22, m.M23, m.M24,
		m.M31, m.M32, m.M33, m.M34,
		value.X, value.Y, value.Z, m.M44,
	}
}

////////////////////////////////////////////////////////////////////////////////

// GetScaling returns the scale factors extracted from each basis row.
func (m Matrix4) GetScaling() Vector3 {

	return Vector3{
		sysMath.Sqrt(m.M11*m.M11 + m.M12*m.M12 + m.M13*m.M13),
		sysMath.Sqrt(m.M21*m.M21 + m.M22*m.M22 + m.M23*m.M23),
		sysMath.Sqrt(m.M31*m.M31 + m.M32*m.M32 + m.M33*m.M33),
	}
}

////////////////////////////////////////////////////////////////////////////////

// GetRotation returns the rotation quaternion extracted from the matrix.
// The scale is divided out before extraction. For pure rotation matrices
// without scale, use `ToQuaternion` instead.
func (m Matrix4) GetRotation() Quaternion {

	s := m.GetScaling()

	invSX := 1 / s.X
	invSY := 1 / s.Y
	invSZ := 1 / s.Z

	rm := Matrix4{
		m.M11 * invSX, m.M12 * invSX, m.M13 * invSX, 0,
		m.M21 * invSY, m.M22 * invSY, m.M23 * invSY, 0,
		m.M31 * invSZ, m.M32 * invSZ, m.M33 * invSZ, 0,
		0, 0, 0, 1,
	}

	return rm.ToQuaternion()
}

////////////////////////////////////////////////////////////////////////////////

// GetMaxScaleOnAxis returns the largest scale factor across all three axes.
func (m Matrix4) GetMaxScaleOnAxis() float64 {

	x := m.M11*m.M11 + m.M12*m.M12 + m.M13*m.M13
	y := m.M21*m.M21 + m.M22*m.M22 + m.M23*m.M23
	z := m.M31*m.M31 + m.M32*m.M32 + m.M33*m.M33

	result := x
	if y > result {
		result = y
	}
	if z > result {
		result = z
	}

	return sysMath.Sqrt(result)
}

////////////////////////////////////////////////////////////////////////////////

// Decompose extracts the rotation, translation, and scale components
// from the matrix. If the determinant is negative, the x-axis scale
// is negated to preserve a valid rotation. The result is approximate
// if the matrix contains shear.
func (m Matrix4) Decompose() (rotation Quaternion, translation Vector3, scale Vector3) {

	sx := sysMath.Sqrt(m.M11*m.M11 + m.M12*m.M12 + m.M13*m.M13)
	sy := sysMath.Sqrt(m.M21*m.M21 + m.M22*m.M22 + m.M23*m.M23)
	sz := sysMath.Sqrt(m.M31*m.M31 + m.M32*m.M32 + m.M33*m.M33)

	// Flip sign for negative determinant
	if m.Determinant() < 0 {
		sx = -sx
	}

	translation = Vector3{m.M41, m.M42, m.M43}
	scale = Vector3{sx, sy, sz}

	// Divide out scale to get pure rotation
	invSX := 1 / sx
	invSY := 1 / sy
	invSZ := 1 / sz

	rm := Matrix4{
		m.M11 * invSX, m.M12 * invSX, m.M13 * invSX, 0,
		m.M21 * invSY, m.M22 * invSY, m.M23 * invSY, 0,
		m.M31 * invSZ, m.M32 * invSZ, m.M33 * invSZ, 0,
		0, 0, 0, 1,
	}

	rotation = rm.ToQuaternion()

	return rotation, translation, scale
}

////////////////////////////////////////////////////////////////////////////////

// Lerp performs element-wise linear interpolation toward the target matrix.
func (m Matrix4) Lerp(target Matrix4, amount float64) Matrix4 {

	return Matrix4{
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

// ToMatrix3 returns a `Matrix3` from the upper-left 3x3 portion
// of the matrix.
func (m Matrix4) ToMatrix3() Matrix3 {

	return Matrix3{
		m.M11, m.M12, m.M13,
		m.M21, m.M22, m.M23,
		m.M31, m.M32, m.M33,
	}
}

////////////////////////////////////////////////////////////////////////////////

// ToQuaternion returns a quaternion from the rotational part of the matrix
// using a numerically stable trace-based extraction. The matrix is assumed
// to be a pure rotation with no scaling. For matrices that may contain scale,
// use `GetRotation` instead.
func (m Matrix4) ToQuaternion() Quaternion {

	trace := m.M11 + m.M22 + m.M33

	if trace > 0 {
		s := sysMath.Sqrt(trace + 1)
		w := s * 0.5
		s = 0.5 / s
		return Quaternion{
			(m.M23 - m.M32) * s,
			(m.M31 - m.M13) * s,
			(m.M12 - m.M21) * s,
			w,
		}
	}

	if m.M11 >= m.M22 && m.M11 >= m.M33 {
		s := sysMath.Sqrt(1 + m.M11 - m.M22 - m.M33)
		half := 0.5 / s
		return Quaternion{
			0.5 * s,
			(m.M12 + m.M21) * half,
			(m.M13 + m.M31) * half,
			(m.M23 - m.M32) * half,
		}
	}

	if m.M22 > m.M33 {
		s := sysMath.Sqrt(1 + m.M22 - m.M11 - m.M33)
		half := 0.5 / s
		return Quaternion{
			(m.M21 + m.M12) * half,
			0.5 * s,
			(m.M32 + m.M23) * half,
			(m.M31 - m.M13) * half,
		}
	}

	s := sysMath.Sqrt(1 + m.M33 - m.M11 - m.M22)
	half := 0.5 / s
	return Quaternion{
		(m.M31 + m.M13) * half,
		(m.M32 + m.M23) * half,
		0.5 * s,
		(m.M12 - m.M21) * half,
	}
}

////////////////////////////////////////////////////////////////////////////////

// ToNormalMatrix3 computes the normal matrix from this 4x4 matrix. This
// is the inverse transpose of the upper-left 3x3. Returns `Matrix3Zero`
// if the matrix is singular.
func (m Matrix4) ToNormalMatrix3() Matrix3 {

	a00 := m.M11
	a01 := m.M12
	a02 := m.M13
	a03 := m.M14
	a10 := m.M21
	a11 := m.M22
	a12 := m.M23
	a13 := m.M24
	a20 := m.M31
	a21 := m.M32
	a22 := m.M33
	a23 := m.M34
	a30 := m.M41
	a31 := m.M42
	a32 := m.M43
	a33 := m.M44

	b00 := a00*a11 - a01*a10
	b01 := a00*a12 - a02*a10
	b02 := a00*a13 - a03*a10
	b03 := a01*a12 - a02*a11
	b04 := a01*a13 - a03*a11
	b05 := a02*a13 - a03*a12
	b06 := a20*a31 - a21*a30
	b07 := a20*a32 - a22*a30
	b08 := a20*a33 - a23*a30
	b09 := a21*a32 - a22*a31
	b10 := a21*a33 - a23*a31
	b11 := a22*a33 - a23*a32

	det := b00*b11 - b01*b10 + b02*b09 + b03*b08 - b04*b07 + b05*b06

	if det == 0 {
		return Matrix3Zero
	}

	inv := 1 / det

	return Matrix3{
		(a11*b11 - a12*b10 + a13*b09) * inv,
		(a12*b08 - a10*b11 - a13*b07) * inv,
		(a10*b10 - a11*b08 + a13*b06) * inv,

		(a02*b10 - a01*b11 - a03*b09) * inv,
		(a00*b11 - a02*b08 + a03*b07) * inv,
		(a01*b08 - a00*b10 - a03*b06) * inv,

		(a31*b05 - a32*b04 + a33*b03) * inv,
		(a32*b02 - a30*b05 - a33*b01) * inv,
		(a30*b04 - a31*b02 + a33*b00) * inv,
	}
}

////////////////////////////////////////////////////////////////////////////////

// ToSlice32 returns the elements as a float32 slice in row-major order.
func (m Matrix4) ToSlice32() []float32 {

	return []float32{
		float32(m.M11), float32(m.M12), float32(m.M13), float32(m.M14),
		float32(m.M21), float32(m.M22), float32(m.M23), float32(m.M24),
		float32(m.M31), float32(m.M32), float32(m.M33), float32(m.M34),
		float32(m.M41), float32(m.M42), float32(m.M43), float32(m.M44),
	}
}

////////////////////////////////////////////////////////////////////////////////

// ToSlice64 returns the elements as a float64 slice in row-major order.
func (m Matrix4) ToSlice64() []float64 {

	return []float64{
		m.M11, m.M12, m.M13, m.M14,
		m.M21, m.M22, m.M23, m.M24,
		m.M31, m.M32, m.M33, m.M34,
		m.M41, m.M42, m.M43, m.M44,
	}
}

////////////////////////////////////////////////////////////////////////////////

// ToBytes32 encodes the elements as 64 bytes of little-endian
// float32 values in row-major order.
func (m Matrix4) ToBytes32() []byte {

	b := make([]byte, 64)
	byteOrder.PutUint32(b[0:4], sysMath.Float32bits(float32(m.M11)))
	byteOrder.PutUint32(b[4:8], sysMath.Float32bits(float32(m.M12)))
	byteOrder.PutUint32(b[8:12], sysMath.Float32bits(float32(m.M13)))
	byteOrder.PutUint32(b[12:16], sysMath.Float32bits(float32(m.M14)))
	byteOrder.PutUint32(b[16:20], sysMath.Float32bits(float32(m.M21)))
	byteOrder.PutUint32(b[20:24], sysMath.Float32bits(float32(m.M22)))
	byteOrder.PutUint32(b[24:28], sysMath.Float32bits(float32(m.M23)))
	byteOrder.PutUint32(b[28:32], sysMath.Float32bits(float32(m.M24)))
	byteOrder.PutUint32(b[32:36], sysMath.Float32bits(float32(m.M31)))
	byteOrder.PutUint32(b[36:40], sysMath.Float32bits(float32(m.M32)))
	byteOrder.PutUint32(b[40:44], sysMath.Float32bits(float32(m.M33)))
	byteOrder.PutUint32(b[44:48], sysMath.Float32bits(float32(m.M34)))
	byteOrder.PutUint32(b[48:52], sysMath.Float32bits(float32(m.M41)))
	byteOrder.PutUint32(b[52:56], sysMath.Float32bits(float32(m.M42)))
	byteOrder.PutUint32(b[56:60], sysMath.Float32bits(float32(m.M43)))
	byteOrder.PutUint32(b[60:64], sysMath.Float32bits(float32(m.M44)))
	return b
}

////////////////////////////////////////////////////////////////////////////////

// ToBytes64 encodes the elements as 128 bytes of little-endian
// float64 values in row-major order.
func (m Matrix4) ToBytes64() []byte {

	b := make([]byte, 128)
	byteOrder.PutUint64(b[0:8], sysMath.Float64bits(m.M11))
	byteOrder.PutUint64(b[8:16], sysMath.Float64bits(m.M12))
	byteOrder.PutUint64(b[16:24], sysMath.Float64bits(m.M13))
	byteOrder.PutUint64(b[24:32], sysMath.Float64bits(m.M14))
	byteOrder.PutUint64(b[32:40], sysMath.Float64bits(m.M21))
	byteOrder.PutUint64(b[40:48], sysMath.Float64bits(m.M22))
	byteOrder.PutUint64(b[48:56], sysMath.Float64bits(m.M23))
	byteOrder.PutUint64(b[56:64], sysMath.Float64bits(m.M24))
	byteOrder.PutUint64(b[64:72], sysMath.Float64bits(m.M31))
	byteOrder.PutUint64(b[72:80], sysMath.Float64bits(m.M32))
	byteOrder.PutUint64(b[80:88], sysMath.Float64bits(m.M33))
	byteOrder.PutUint64(b[88:96], sysMath.Float64bits(m.M34))
	byteOrder.PutUint64(b[96:104], sysMath.Float64bits(m.M41))
	byteOrder.PutUint64(b[104:112], sysMath.Float64bits(m.M42))
	byteOrder.PutUint64(b[112:120], sysMath.Float64bits(m.M43))
	byteOrder.PutUint64(b[120:128], sysMath.Float64bits(m.M44))
	return b
}

//----------------------------------------------------------------------------//
// Static                                                                     //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// Matrix4Compose creates a transformation matrix from a quaternion rotation,
// translation vector, and scale vector.
func Matrix4Compose(rotation Quaternion, translation, scale Vector3) Matrix4 {

	x1 := rotation.X
	y1 := rotation.Y
	z1 := rotation.Z
	w1 := rotation.W

	x2 := x1 + x1
	y2 := y1 + y1
	z2 := z1 + z1

	xx := x1 * x2
	xy := x1 * y2
	xz := x1 * z2

	yy := y1 * y2
	yz := y1 * z2
	zz := z1 * z2

	wx := w1 * x2
	wy := w1 * y2
	wz := w1 * z2

	sx := scale.X
	sy := scale.Y
	sz := scale.Z

	return Matrix4{
		(1 - (yy + zz)) * sx, (xy + wz) * sx, (xz - wy) * sx, 0,
		(xy - wz) * sy, (1 - (xx + zz)) * sy, (yz + wx) * sy, 0,
		(xz + wy) * sz, (yz - wx) * sz, (1 - (xx + yy)) * sz, 0,
		translation.X, translation.Y, translation.Z, 1,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Matrix4FromSlice32 creates a Matrix4 from a float32 slice in row-major order.
func Matrix4FromSlice32(values []float32) (Matrix4, error) {

	if len(values) != 16 {
		return Matrix4Zero, ErrInvalidLength
	}

	m := Matrix4{
		float64(values[0]), float64(values[1]), float64(values[2]), float64(values[3]),
		float64(values[4]), float64(values[5]), float64(values[6]), float64(values[7]),
		float64(values[8]), float64(values[9]), float64(values[10]), float64(values[11]),
		float64(values[12]), float64(values[13]), float64(values[14]), float64(values[15]),
	}

	return m, nil
}

////////////////////////////////////////////////////////////////////////////////

// Matrix4FromSlice64 creates a Matrix4 from a float64 slice in row-major order.
func Matrix4FromSlice64(values []float64) (Matrix4, error) {

	if len(values) != 16 {
		return Matrix4Zero, ErrInvalidLength
	}

	m := Matrix4{
		values[0], values[1], values[2], values[3],
		values[4], values[5], values[6], values[7],
		values[8], values[9], values[10], values[11],
		values[12], values[13], values[14], values[15],
	}

	return m, nil
}

////////////////////////////////////////////////////////////////////////////////

// Matrix4FromBytes32 creates a Matrix4 from a byte slice of at
// least 64 bytes containing 16 little-endian float32 values in
// row-major order.
func Matrix4FromBytes32(data []byte) (Matrix4, error) {

	if len(data) < 64 {
		return Matrix4Zero, ErrInvalidLength
	}

	return Matrix4{
		M11: float64(sysMath.Float32frombits(byteOrder.Uint32(data[0:4]))),
		M12: float64(sysMath.Float32frombits(byteOrder.Uint32(data[4:8]))),
		M13: float64(sysMath.Float32frombits(byteOrder.Uint32(data[8:12]))),
		M14: float64(sysMath.Float32frombits(byteOrder.Uint32(data[12:16]))),
		M21: float64(sysMath.Float32frombits(byteOrder.Uint32(data[16:20]))),
		M22: float64(sysMath.Float32frombits(byteOrder.Uint32(data[20:24]))),
		M23: float64(sysMath.Float32frombits(byteOrder.Uint32(data[24:28]))),
		M24: float64(sysMath.Float32frombits(byteOrder.Uint32(data[28:32]))),
		M31: float64(sysMath.Float32frombits(byteOrder.Uint32(data[32:36]))),
		M32: float64(sysMath.Float32frombits(byteOrder.Uint32(data[36:40]))),
		M33: float64(sysMath.Float32frombits(byteOrder.Uint32(data[40:44]))),
		M34: float64(sysMath.Float32frombits(byteOrder.Uint32(data[44:48]))),
		M41: float64(sysMath.Float32frombits(byteOrder.Uint32(data[48:52]))),
		M42: float64(sysMath.Float32frombits(byteOrder.Uint32(data[52:56]))),
		M43: float64(sysMath.Float32frombits(byteOrder.Uint32(data[56:60]))),
		M44: float64(sysMath.Float32frombits(byteOrder.Uint32(data[60:64]))),
	}, nil
}

////////////////////////////////////////////////////////////////////////////////

// Matrix4FromBytes64 creates a Matrix4 from a byte slice of at
// least 128 bytes containing 16 little-endian float64 values in
// row-major order.
func Matrix4FromBytes64(data []byte) (Matrix4, error) {

	if len(data) < 128 {
		return Matrix4Zero, ErrInvalidLength
	}

	return Matrix4{
		M11: sysMath.Float64frombits(byteOrder.Uint64(data[0:8])),
		M12: sysMath.Float64frombits(byteOrder.Uint64(data[8:16])),
		M13: sysMath.Float64frombits(byteOrder.Uint64(data[16:24])),
		M14: sysMath.Float64frombits(byteOrder.Uint64(data[24:32])),
		M21: sysMath.Float64frombits(byteOrder.Uint64(data[32:40])),
		M22: sysMath.Float64frombits(byteOrder.Uint64(data[40:48])),
		M23: sysMath.Float64frombits(byteOrder.Uint64(data[48:56])),
		M24: sysMath.Float64frombits(byteOrder.Uint64(data[56:64])),
		M31: sysMath.Float64frombits(byteOrder.Uint64(data[64:72])),
		M32: sysMath.Float64frombits(byteOrder.Uint64(data[72:80])),
		M33: sysMath.Float64frombits(byteOrder.Uint64(data[80:88])),
		M34: sysMath.Float64frombits(byteOrder.Uint64(data[88:96])),
		M41: sysMath.Float64frombits(byteOrder.Uint64(data[96:104])),
		M42: sysMath.Float64frombits(byteOrder.Uint64(data[104:112])),
		M43: sysMath.Float64frombits(byteOrder.Uint64(data[112:120])),
		M44: sysMath.Float64frombits(byteOrder.Uint64(data[120:128])),
	}, nil
}

////////////////////////////////////////////////////////////////////////////////

// Matrix4FromAxisAngle creates a rotation matrix from an arbitrary axis
// and angle in radians.
func Matrix4FromAxisAngle(axis Vector3, angle float64) Matrix4 {

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

	return Matrix4{
		xx + (1-xx)*c, xy - xy*c + z*s, xz - xz*c - y*s, 0,
		xy - xy*c - z*s, yy + (1-yy)*c, yz - yz*c + x*s, 0,
		xz - xz*c + y*s, yz - yz*c - x*s, zz + (1-zz)*c, 0,
		0, 0, 0, 1,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Matrix4CreateProjection creates a perspective projection matrix from the
// given field of view, viewport dimensions, and near/far clipping planes.
// Returns `Matrix4Zero` if any parameter is invalid.
func Matrix4CreateProjection(fov float64, viewport Size, near, far float64) Matrix4 {

	// Do parameter check
	if fov <= 0 || fov >= sysMath.Pi || near <= 0 || far <= 0 || near >= far {
		return Matrix4Zero
	}

	w := float64(viewport.W)
	h := float64(viewport.H)

	dif := near - far
	cot := 1 / sysMath.Tan(fov*0.5)
	asp := cot / (w / h)

	m33 := far / dif
	m43 := near * far / dif

	return Matrix4{
		asp, 0.0, 0.0, 0.0,
		0.0, cot, 0.0, 0.0,
		0.0, 0.0, m33, -1.0,
		0.0, 0.0, m43, 0.0,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Matrix4CreateOrthographic creates an orthographic projection matrix from the
// given viewport dimensions and near/far clipping planes.
func Matrix4CreateOrthographic(viewport Size, near, far float64) Matrix4 {

	w := float64(viewport.W)
	h := float64(viewport.H)

	dif := near - far
	m11 := 2 / w
	m22 := 2 / h
	m33 := 1 / dif
	m43 := near / dif

	return Matrix4{
		m11, 0.0, 0.0, 0.0,
		0.0, m22, 0.0, 0.0,
		0.0, 0.0, m33, 0.0,
		0.0, 0.0, m43, 1.0,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Matrix4CreateView creates a view matrix from the camera position, target,
// and up direction.
func Matrix4CreateView(pos, target, up Vector3) Matrix4 {

	vz := pos.Sub(target).Normalize()
	vx := up.Cross(vz).Normalize()
	vy := vz.Cross(vx).Normalize()

	return Matrix4{
		vx.X, vy.X, vz.X, 0,
		vx.Y, vy.Y, vz.Y, 0,
		vx.Z, vy.Z, vz.Z, 0,
		-vx.Dot(pos),
		-vy.Dot(pos),
		-vz.Dot(pos), 1,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Matrix4CreateWorld creates a world transformation matrix from a position,
// forward direction, and up direction.
func Matrix4CreateWorld(pos, forward, up Vector3) Matrix4 {

	vz := forward.Neg().Normalize()
	vx := up.Cross(vz).Normalize()
	vy := vz.Cross(vx).Normalize()

	return Matrix4{
		vx.X, vx.Y, vx.Z, 0,
		vy.X, vy.Y, vy.Z, 0,
		vz.X, vz.Y, vz.Z, 0,
		pos.X, pos.Y, pos.Z, 1,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Matrix4CreateScale creates a scaling matrix with the given scale factors
// along each axis.
func Matrix4CreateScale(x, y, z float64) Matrix4 {

	return Matrix4{
		x, 0, 0, 0,
		0, y, 0, 0,
		0, 0, z, 0,
		0, 0, 0, 1,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Matrix4CreateRotationX creates a rotation matrix around the x-axis by the
// given angle in radians.
func Matrix4CreateRotationX(radians float64) Matrix4 {

	c := sysMath.Cos(radians)
	s := sysMath.Sin(radians)

	return Matrix4{
		1, 0, 0, 0,
		0, c, s, 0,
		0, -s, c, 0,
		0, 0, 0, 1,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Matrix4CreateRotationY creates a rotation matrix around the y-axis by the
// given angle in radians.
func Matrix4CreateRotationY(radians float64) Matrix4 {

	c := sysMath.Cos(radians)
	s := sysMath.Sin(radians)

	return Matrix4{
		c, 0, -s, 0,
		0, 1, 0, 0,
		s, 0, c, 0,
		0, 0, 0, 1,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Matrix4CreateRotationZ creates a rotation matrix around the z-axis by the
// given angle in radians.
func Matrix4CreateRotationZ(radians float64) Matrix4 {

	c := sysMath.Cos(radians)
	s := sysMath.Sin(radians)

	return Matrix4{
		c, s, 0, 0,
		-s, c, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Matrix4CreateTranslation creates a translation matrix with the given offsets
// along each axis.
func Matrix4CreateTranslation(x, y, z float64) Matrix4 {

	return Matrix4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		x, y, z, 1,
	}
}

//----------------------------------------------------------------------------//
// Operators                                                                  //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// Add returns the element-wise sum of two matrices.
func (m Matrix4) Add(value Matrix4) Matrix4 {

	return Matrix4{
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

// Sub returns the element-wise difference of two matrices.
func (m Matrix4) Sub(value Matrix4) Matrix4 {

	return Matrix4{
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

// Mul returns the matrix product of two matrices.
func (m Matrix4) Mul(value Matrix4) Matrix4 {

	return Matrix4{
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

// Div returns the element-wise quotient of two matrices.
func (m Matrix4) Div(value Matrix4) Matrix4 {

	return Matrix4{
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

// AddScalar adds a scalar to each element.
func (m Matrix4) AddScalar(scalar float64) Matrix4 {

	return Matrix4{
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

// SubScalar subtracts a scalar from each element.
func (m Matrix4) SubScalar(scalar float64) Matrix4 {

	return Matrix4{
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

// MulScalar multiplies each element by a scalar.
func (m Matrix4) MulScalar(scalar float64) Matrix4 {

	return Matrix4{
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

// DivScalar divides each element by a scalar.
func (m Matrix4) DivScalar(scalar float64) Matrix4 {

	return Matrix4{
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

// Neg returns the negation of the matrix.
func (m Matrix4) Neg() Matrix4 {

	return Matrix4{
		-m.M11, -m.M12, -m.M13, -m.M14,
		-m.M21, -m.M22, -m.M23, -m.M24,
		-m.M31, -m.M32, -m.M33, -m.M34,
		-m.M41, -m.M42, -m.M43, -m.M44,
	}
}
