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
	// Matrix3Zero represents a 3x3 matrix with all elements set to zero.
	Matrix3Zero = Matrix3{0, 0, 0, 0, 0, 0, 0, 0, 0}

	// Matrix3Identity represents the 3x3 identity matrix.
	Matrix3Identity = Matrix3{1, 0, 0, 0, 1, 0, 0, 0, 1}
)

//----------------------------------------------------------------------------//
// Types                                                                      //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// Matrix3 represents a 3x3 matrix stored in row-major order.
type Matrix3 struct {
	M11, M12, M13 float64
	M21, M22, M23 float64
	M31, M32, M33 float64
}

//----------------------------------------------------------------------------//
// Methods                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// String returns the string representation of the matrix.
func (m Matrix3) String() string {

	return fmt.Sprintf(
		"["+
			"[%.2f, %.2f, %.2f]"+
			"[%.2f, %.2f, %.2f]"+
			"[%.2f, %.2f, %.2f]"+
			"]",
		m.M11, m.M12, m.M13,
		m.M21, m.M22, m.M23,
		m.M31, m.M32, m.M33,
	)
}

////////////////////////////////////////////////////////////////////////////////

// IsZero returns whether all elements are zero.
func (m Matrix3) IsZero() bool {

	return m.M11 == 0 &&
		m.M12 == 0 &&
		m.M13 == 0 &&

		m.M21 == 0 &&
		m.M22 == 0 &&
		m.M23 == 0 &&

		m.M31 == 0 &&
		m.M32 == 0 &&
		m.M33 == 0
}

////////////////////////////////////////////////////////////////////////////////

// Transpose returns the transpose of the matrix.
func (m Matrix3) Transpose() Matrix3 {

	return Matrix3{
		m.M11, m.M21, m.M31,
		m.M12, m.M22, m.M32,
		m.M13, m.M23, m.M33,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Invert returns the inverse of the matrix. Returns `Matrix3Zero` if the
// matrix is singular.
func (m Matrix3) Invert() Matrix3 {

	b01 := m.M33*m.M22 - m.M23*m.M32
	b11 := -m.M33*m.M21 + m.M23*m.M31
	b21 := m.M32*m.M21 - m.M22*m.M31

	det := m.M11*b01 + m.M12*b11 + m.M13*b21

	if det == 0 {
		return Matrix3Zero
	}

	inv := 1 / det

	return Matrix3{
		b01 * inv,
		(-m.M33*m.M12 + m.M13*m.M32) * inv,
		(m.M23*m.M12 - m.M13*m.M22) * inv,

		b11 * inv,
		(m.M33*m.M11 - m.M13*m.M31) * inv,
		(-m.M23*m.M11 + m.M13*m.M21) * inv,

		b21 * inv,
		(-m.M32*m.M11 + m.M12*m.M31) * inv,
		(m.M22*m.M11 - m.M12*m.M21) * inv,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Determinant returns the determinant of the matrix.
func (m Matrix3) Determinant() float64 {

	return m.M11*(m.M33*m.M22-m.M23*m.M32) -
		m.M12*(m.M33*m.M21-m.M23*m.M31) +
		m.M13*(m.M32*m.M21-m.M22*m.M31)
}

////////////////////////////////////////////////////////////////////////////////

// Lerp performs element-wise linear interpolation toward the target matrix.
func (m Matrix3) Lerp(target Matrix3, amount float64) Matrix3 {

	return Matrix3{
		m.M11 + (target.M11-m.M11)*amount,
		m.M12 + (target.M12-m.M12)*amount,
		m.M13 + (target.M13-m.M13)*amount,

		m.M21 + (target.M21-m.M21)*amount,
		m.M22 + (target.M22-m.M22)*amount,
		m.M23 + (target.M23-m.M23)*amount,

		m.M31 + (target.M31-m.M31)*amount,
		m.M32 + (target.M32-m.M32)*amount,
		m.M33 + (target.M33-m.M33)*amount,
	}
}

////////////////////////////////////////////////////////////////////////////////

// ToMatrix4 returns a `Matrix4` with the 3x3 elements embedded in
// the upper-left and the fourth row and column set to identity.
func (m Matrix3) ToMatrix4() Matrix4 {

	return Matrix4{
		m.M11, m.M12, m.M13, 0,
		m.M21, m.M22, m.M23, 0,
		m.M31, m.M32, m.M33, 0,
		0, 0, 0, 1,
	}
}

////////////////////////////////////////////////////////////////////////////////

// ToSlice32 returns the elements as a float32 slice in row-major order.
func (m Matrix3) ToSlice32() []float32 {

	return []float32{
		float32(m.M11), float32(m.M12), float32(m.M13),
		float32(m.M21), float32(m.M22), float32(m.M23),
		float32(m.M31), float32(m.M32), float32(m.M33),
	}
}

////////////////////////////////////////////////////////////////////////////////

// ToSlice64 returns the elements as a float64 slice in row-major order.
func (m Matrix3) ToSlice64() []float64 {

	return []float64{
		m.M11, m.M12, m.M13,
		m.M21, m.M22, m.M23,
		m.M31, m.M32, m.M33,
	}
}

////////////////////////////////////////////////////////////////////////////////

// ToBytes32 encodes the elements as 36 bytes of little-endian
// float32 values in row-major order.
func (m Matrix3) ToBytes32() []byte {

	b := make([]byte, 36)
	byteOrder.PutUint32(b[0:4], sysMath.Float32bits(float32(m.M11)))
	byteOrder.PutUint32(b[4:8], sysMath.Float32bits(float32(m.M12)))
	byteOrder.PutUint32(b[8:12], sysMath.Float32bits(float32(m.M13)))
	byteOrder.PutUint32(b[12:16], sysMath.Float32bits(float32(m.M21)))
	byteOrder.PutUint32(b[16:20], sysMath.Float32bits(float32(m.M22)))
	byteOrder.PutUint32(b[20:24], sysMath.Float32bits(float32(m.M23)))
	byteOrder.PutUint32(b[24:28], sysMath.Float32bits(float32(m.M31)))
	byteOrder.PutUint32(b[28:32], sysMath.Float32bits(float32(m.M32)))
	byteOrder.PutUint32(b[32:36], sysMath.Float32bits(float32(m.M33)))
	return b
}

////////////////////////////////////////////////////////////////////////////////

// ToBytes64 encodes the elements as 72 bytes of little-endian
// float64 values in row-major order.
func (m Matrix3) ToBytes64() []byte {

	b := make([]byte, 72)
	byteOrder.PutUint64(b[0:8], sysMath.Float64bits(m.M11))
	byteOrder.PutUint64(b[8:16], sysMath.Float64bits(m.M12))
	byteOrder.PutUint64(b[16:24], sysMath.Float64bits(m.M13))
	byteOrder.PutUint64(b[24:32], sysMath.Float64bits(m.M21))
	byteOrder.PutUint64(b[32:40], sysMath.Float64bits(m.M22))
	byteOrder.PutUint64(b[40:48], sysMath.Float64bits(m.M23))
	byteOrder.PutUint64(b[48:56], sysMath.Float64bits(m.M31))
	byteOrder.PutUint64(b[56:64], sysMath.Float64bits(m.M32))
	byteOrder.PutUint64(b[64:72], sysMath.Float64bits(m.M33))
	return b
}

//----------------------------------------------------------------------------//
// Static                                                                     //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// Matrix3FromSlice32 creates a Matrix3 from a float32 slice in row-major order.
func Matrix3FromSlice32(values []float32) (Matrix3, error) {

	if len(values) != 9 {
		return Matrix3Zero, ErrInvalidLength
	}

	m := Matrix3{
		float64(values[0]), float64(values[1]), float64(values[2]),
		float64(values[3]), float64(values[4]), float64(values[5]),
		float64(values[6]), float64(values[7]), float64(values[8]),
	}

	return m, nil
}

////////////////////////////////////////////////////////////////////////////////

// Matrix3FromSlice64 creates a Matrix3 from a float64 slice in row-major order.
func Matrix3FromSlice64(values []float64) (Matrix3, error) {

	if len(values) != 9 {
		return Matrix3Zero, ErrInvalidLength
	}

	m := Matrix3{
		values[0], values[1], values[2],
		values[3], values[4], values[5],
		values[6], values[7], values[8],
	}

	return m, nil
}

////////////////////////////////////////////////////////////////////////////////

// Matrix3FromBytes32 creates a Matrix3 from a byte slice of at
// least 36 bytes containing 9 little-endian float32 values in
// row-major order.
func Matrix3FromBytes32(data []byte) (Matrix3, error) {

	if len(data) < 36 {
		return Matrix3Zero, ErrInvalidLength
	}

	return Matrix3{
		M11: float64(sysMath.Float32frombits(byteOrder.Uint32(data[0:4]))),
		M12: float64(sysMath.Float32frombits(byteOrder.Uint32(data[4:8]))),
		M13: float64(sysMath.Float32frombits(byteOrder.Uint32(data[8:12]))),
		M21: float64(sysMath.Float32frombits(byteOrder.Uint32(data[12:16]))),
		M22: float64(sysMath.Float32frombits(byteOrder.Uint32(data[16:20]))),
		M23: float64(sysMath.Float32frombits(byteOrder.Uint32(data[20:24]))),
		M31: float64(sysMath.Float32frombits(byteOrder.Uint32(data[24:28]))),
		M32: float64(sysMath.Float32frombits(byteOrder.Uint32(data[28:32]))),
		M33: float64(sysMath.Float32frombits(byteOrder.Uint32(data[32:36]))),
	}, nil
}

////////////////////////////////////////////////////////////////////////////////

// Matrix3FromBytes64 creates a Matrix3 from a byte slice of at
// least 72 bytes containing 9 little-endian float64 values in
// row-major order.
func Matrix3FromBytes64(data []byte) (Matrix3, error) {

	if len(data) < 72 {
		return Matrix3Zero, ErrInvalidLength
	}

	return Matrix3{
		M11: sysMath.Float64frombits(byteOrder.Uint64(data[0:8])),
		M12: sysMath.Float64frombits(byteOrder.Uint64(data[8:16])),
		M13: sysMath.Float64frombits(byteOrder.Uint64(data[16:24])),
		M21: sysMath.Float64frombits(byteOrder.Uint64(data[24:32])),
		M22: sysMath.Float64frombits(byteOrder.Uint64(data[32:40])),
		M23: sysMath.Float64frombits(byteOrder.Uint64(data[40:48])),
		M31: sysMath.Float64frombits(byteOrder.Uint64(data[48:56])),
		M32: sysMath.Float64frombits(byteOrder.Uint64(data[56:64])),
		M33: sysMath.Float64frombits(byteOrder.Uint64(data[64:72])),
	}, nil
}

////////////////////////////////////////////////////////////////////////////////

// Matrix3FromBasis creates a 3x3 matrix from three row basis vectors.
func Matrix3FromBasis(a, b, c Vector3) Matrix3 {

	return Matrix3{
		a.X, a.Y, a.Z,
		b.X, b.Y, b.Z,
		c.X, c.Y, c.Z,
	}
}

//----------------------------------------------------------------------------//
// Operators                                                                  //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// Add returns the element-wise sum of two matrices.
func (m Matrix3) Add(value Matrix3) Matrix3 {

	return Matrix3{
		m.M11 + value.M11,
		m.M12 + value.M12,
		m.M13 + value.M13,

		m.M21 + value.M21,
		m.M22 + value.M22,
		m.M23 + value.M23,

		m.M31 + value.M31,
		m.M32 + value.M32,
		m.M33 + value.M33,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Sub returns the element-wise difference of two matrices.
func (m Matrix3) Sub(value Matrix3) Matrix3 {

	return Matrix3{
		m.M11 - value.M11,
		m.M12 - value.M12,
		m.M13 - value.M13,

		m.M21 - value.M21,
		m.M22 - value.M22,
		m.M23 - value.M23,

		m.M31 - value.M31,
		m.M32 - value.M32,
		m.M33 - value.M33,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Mul returns the matrix product of two matrices.
func (m Matrix3) Mul(value Matrix3) Matrix3 {

	return Matrix3{
		m.M11*value.M11 + m.M12*value.M21 + m.M13*value.M31,
		m.M11*value.M12 + m.M12*value.M22 + m.M13*value.M32,
		m.M11*value.M13 + m.M12*value.M23 + m.M13*value.M33,

		m.M21*value.M11 + m.M22*value.M21 + m.M23*value.M31,
		m.M21*value.M12 + m.M22*value.M22 + m.M23*value.M32,
		m.M21*value.M13 + m.M22*value.M23 + m.M23*value.M33,

		m.M31*value.M11 + m.M32*value.M21 + m.M33*value.M31,
		m.M31*value.M12 + m.M32*value.M22 + m.M33*value.M32,
		m.M31*value.M13 + m.M32*value.M23 + m.M33*value.M33,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Div returns the element-wise quotient of two matrices.
func (m Matrix3) Div(value Matrix3) Matrix3 {

	return Matrix3{
		m.M11 / value.M11,
		m.M12 / value.M12,
		m.M13 / value.M13,

		m.M21 / value.M21,
		m.M22 / value.M22,
		m.M23 / value.M23,

		m.M31 / value.M31,
		m.M32 / value.M32,
		m.M33 / value.M33,
	}
}

////////////////////////////////////////////////////////////////////////////////

// AddScalar adds a scalar to each element.
func (m Matrix3) AddScalar(scalar float64) Matrix3 {

	return Matrix3{
		m.M11 + scalar,
		m.M12 + scalar,
		m.M13 + scalar,

		m.M21 + scalar,
		m.M22 + scalar,
		m.M23 + scalar,

		m.M31 + scalar,
		m.M32 + scalar,
		m.M33 + scalar,
	}
}

////////////////////////////////////////////////////////////////////////////////

// SubScalar subtracts a scalar from each element.
func (m Matrix3) SubScalar(scalar float64) Matrix3 {

	return Matrix3{
		m.M11 - scalar,
		m.M12 - scalar,
		m.M13 - scalar,

		m.M21 - scalar,
		m.M22 - scalar,
		m.M23 - scalar,

		m.M31 - scalar,
		m.M32 - scalar,
		m.M33 - scalar,
	}
}

////////////////////////////////////////////////////////////////////////////////

// MulScalar multiplies each element by a scalar.
func (m Matrix3) MulScalar(scalar float64) Matrix3 {

	return Matrix3{
		m.M11 * scalar,
		m.M12 * scalar,
		m.M13 * scalar,

		m.M21 * scalar,
		m.M22 * scalar,
		m.M23 * scalar,

		m.M31 * scalar,
		m.M32 * scalar,
		m.M33 * scalar,
	}
}

////////////////////////////////////////////////////////////////////////////////

// DivScalar divides each element by a scalar.
func (m Matrix3) DivScalar(scalar float64) Matrix3 {

	return Matrix3{
		m.M11 / scalar,
		m.M12 / scalar,
		m.M13 / scalar,

		m.M21 / scalar,
		m.M22 / scalar,
		m.M23 / scalar,

		m.M31 / scalar,
		m.M32 / scalar,
		m.M33 / scalar,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Neg returns the negation of the matrix.
func (m Matrix3) Neg() Matrix3 {

	return Matrix3{
		-m.M11, -m.M12, -m.M13,
		-m.M21, -m.M22, -m.M23,
		-m.M31, -m.M32, -m.M33,
	}
}

