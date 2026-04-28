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
	// EulerZero represents an Euler with all angles set to zero.
	EulerZero = Euler{0, 0, 0, RotationOrderYXZ}
)

//----------------------------------------------------------------------------//
// Types                                                                      //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// RotationOrderType defines the axis sequence for Euler angle decomposition.
type RotationOrderType uint8

const (
	// RotationOrderXYZ applies rotations in X then Y then Z (pitch, yaw, roll).
	RotationOrderXYZ RotationOrderType = iota

	// RotationOrderYXZ applies rotations in Y then X then Z (yaw, pitch, roll).
	RotationOrderYXZ

	// RotationOrderZXY applies rotations in Z then X then Y (roll, pitch, yaw).
	RotationOrderZXY

	// RotationOrderZYX applies rotations in Z then Y then X (roll, yaw, pitch).
	RotationOrderZYX

	// RotationOrderYZX applies rotations in Y then Z then X (yaw, roll, pitch).
	RotationOrderYZX

	// RotationOrderXZY applies rotations in X then Z then Y (pitch, roll, yaw).
	RotationOrderXZY
)

////////////////////////////////////////////////////////////////////////////////

// String returns the string representation of the rotation order.
func (o RotationOrderType) String() string {

	switch o {
	case RotationOrderXYZ:
		return "xyz"
	case RotationOrderYXZ:
		return "yxz"
	case RotationOrderZXY:
		return "zxy"
	case RotationOrderZYX:
		return "zyx"
	case RotationOrderYZX:
		return "yzx"
	case RotationOrderXZY:
		return "xzy"
	default:
		return ""
	}
}

//----------------------------------------------------------------------------//
// Types                                                                      //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// Euler represents a rotation as three angles with a specified rotation order.
type Euler struct {
	X     float64
	Y     float64
	Z     float64
	Order RotationOrderType
}

//----------------------------------------------------------------------------//
// Methods                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// String returns the string representation of the Euler angles.
func (e Euler) String() string {

	return fmt.Sprintf("[%.2f, %.2f, %.2f, %s]", e.X, e.Y, e.Z, e.Order)
}

////////////////////////////////////////////////////////////////////////////////

// IsZero returns whether all angles are zero.
func (e Euler) IsZero() bool {

	return e.X == 0 && e.Y == 0 && e.Z == 0
}

////////////////////////////////////////////////////////////////////////////////

// ToQuaternion returns a rotation quaternion from the Euler angles,
// respecting the rotation order.
func (e Euler) ToQuaternion() Quaternion {

	cx := sysMath.Cos(e.X * 0.5)
	sx := sysMath.Sin(e.X * 0.5)
	cy := sysMath.Cos(e.Y * 0.5)
	sy := sysMath.Sin(e.Y * 0.5)
	cz := sysMath.Cos(e.Z * 0.5)
	sz := sysMath.Sin(e.Z * 0.5)

	switch e.Order {
	case RotationOrderXYZ:
		return Quaternion{
			sx*cy*cz + cx*sy*sz,
			cx*sy*cz - sx*cy*sz,
			cx*cy*sz + sx*sy*cz,
			cx*cy*cz - sx*sy*sz,
		}

	case RotationOrderYXZ:
		return Quaternion{
			sx*cy*cz + cx*sy*sz,
			cx*sy*cz - sx*cy*sz,
			cx*cy*sz - sx*sy*cz,
			cx*cy*cz + sx*sy*sz,
		}

	case RotationOrderZXY:
		return Quaternion{
			sx*cy*cz - cx*sy*sz,
			cx*sy*cz + sx*cy*sz,
			cx*cy*sz + sx*sy*cz,
			cx*cy*cz - sx*sy*sz,
		}

	case RotationOrderZYX:
		return Quaternion{
			sx*cy*cz - cx*sy*sz,
			cx*sy*cz + sx*cy*sz,
			cx*cy*sz - sx*sy*cz,
			cx*cy*cz + sx*sy*sz,
		}

	case RotationOrderYZX:
		return Quaternion{
			sx*cy*cz + cx*sy*sz,
			cx*sy*cz + sx*cy*sz,
			cx*cy*sz - sx*sy*cz,
			cx*cy*cz - sx*sy*sz,
		}

	case RotationOrderXZY:
		return Quaternion{
			sx*cy*cz - cx*sy*sz,
			cx*sy*cz - sx*cy*sz,
			cx*cy*sz + sx*sy*cz,
			cx*cy*cz + sx*sy*sz,
		}

	default:
		return QuaternionZero
	}
}

////////////////////////////////////////////////////////////////////////////////

// ToMatrix4 returns a 4x4 rotation matrix from the Euler angles,
// respecting the rotation order.
func (e Euler) ToMatrix4() Matrix4 {

	return e.ToQuaternion().ToMatrix4()
}

////////////////////////////////////////////////////////////////////////////////

// ToSlice32 returns the angles as a float32 slice.
func (e Euler) ToSlice32() []float32 {

	return []float32{
		float32(e.X), float32(e.Y), float32(e.Z),
	}
}

////////////////////////////////////////////////////////////////////////////////

// ToSlice64 returns the angles as a float64 slice.
func (e Euler) ToSlice64() []float64 {

	return []float64{e.X, e.Y, e.Z}
}

////////////////////////////////////////////////////////////////////////////////

// ToBytes32 encodes the angles as 12 bytes of little-endian
// float32 values. The rotation order is not included.
func (e Euler) ToBytes32() []byte {

	b := make([]byte, 12)
	byteOrder.PutUint32(b[0:4], sysMath.Float32bits(float32(e.X)))
	byteOrder.PutUint32(b[4:8], sysMath.Float32bits(float32(e.Y)))
	byteOrder.PutUint32(b[8:12], sysMath.Float32bits(float32(e.Z)))
	return b
}

////////////////////////////////////////////////////////////////////////////////

// ToBytes64 encodes the angles as 24 bytes of little-endian
// float64 values. The rotation order is not included.
func (e Euler) ToBytes64() []byte {

	b := make([]byte, 24)
	byteOrder.PutUint64(b[0:8], sysMath.Float64bits(e.X))
	byteOrder.PutUint64(b[8:16], sysMath.Float64bits(e.Y))
	byteOrder.PutUint64(b[16:24], sysMath.Float64bits(e.Z))
	return b
}

//----------------------------------------------------------------------------//
// Static                                                                     //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// EulerFromSlice32 creates an Euler from a float32 slice and
// a rotation order. The slice must have exactly 3 elements.
func EulerFromSlice32(values []float32, order RotationOrderType) (Euler, error) {

	if len(values) != 3 {
		return EulerZero, ErrInvalidLength
	}

	return Euler{
		X:     float64(values[0]),
		Y:     float64(values[1]),
		Z:     float64(values[2]),
		Order: order,
	}, nil
}

////////////////////////////////////////////////////////////////////////////////

// EulerFromSlice64 creates an Euler from a float64 slice and
// a rotation order. The slice must have exactly 3 elements.
func EulerFromSlice64(values []float64, order RotationOrderType) (Euler, error) {

	if len(values) != 3 {
		return EulerZero, ErrInvalidLength
	}

	return Euler{
		X:     values[0],
		Y:     values[1],
		Z:     values[2],
		Order: order,
	}, nil
}

////////////////////////////////////////////////////////////////////////////////

// EulerFromBytes32 creates an Euler from a byte slice of at
// least 12 bytes containing 3 little-endian float32 values
// and a rotation order.
func EulerFromBytes32(data []byte, order RotationOrderType) (Euler, error) {

	if len(data) < 12 {
		return EulerZero, ErrInvalidLength
	}

	return Euler{
		X:     float64(sysMath.Float32frombits(byteOrder.Uint32(data[0:4]))),
		Y:     float64(sysMath.Float32frombits(byteOrder.Uint32(data[4:8]))),
		Z:     float64(sysMath.Float32frombits(byteOrder.Uint32(data[8:12]))),
		Order: order,
	}, nil
}

////////////////////////////////////////////////////////////////////////////////

// EulerFromBytes64 creates an Euler from a byte slice of at
// least 24 bytes containing 3 little-endian float64 values
// and a rotation order.
func EulerFromBytes64(data []byte, order RotationOrderType) (Euler, error) {

	if len(data) < 24 {
		return EulerZero, ErrInvalidLength
	}

	return Euler{
		X:     sysMath.Float64frombits(byteOrder.Uint64(data[0:8])),
		Y:     sysMath.Float64frombits(byteOrder.Uint64(data[8:16])),
		Z:     sysMath.Float64frombits(byteOrder.Uint64(data[16:24])),
		Order: order,
	}, nil
}

////////////////////////////////////////////////////////////////////////////////

// EulerFromRotationMatrix extracts Euler angles from the upper-left 3x3 of
// the given rotation matrix using the specified rotation order. The matrix
// is assumed to contain a pure rotation with no scaling. Results will be
// incorrect if the matrix contains scale or shear. Returns `EulerZero` if
// the rotation order is not recognized.
func EulerFromRotationMatrix(matrix Matrix4, order RotationOrderType) Euler {

	var x, y, z float64

	switch order {
	case RotationOrderXYZ:
		y = sysMath.Asin(Clamp(matrix.M13, -1, 1))

		if sysMath.Abs(matrix.M13) < 0.99999 {
			x = sysMath.Atan2(-matrix.M23, matrix.M33)
			z = sysMath.Atan2(-matrix.M12, matrix.M11)
		} else {
			x = sysMath.Atan2(matrix.M32, matrix.M22)
			z = 0
		}

	case RotationOrderYXZ:
		x = sysMath.Asin(-Clamp(matrix.M23, -1, 1))

		if sysMath.Abs(matrix.M23) < 0.99999 {
			y = sysMath.Atan2(matrix.M13, matrix.M33)
			z = sysMath.Atan2(matrix.M21, matrix.M22)
		} else {
			y = sysMath.Atan2(-matrix.M31, matrix.M11)
			z = 0
		}

	case RotationOrderZXY:
		x = sysMath.Asin(Clamp(matrix.M32, -1, 1))

		if sysMath.Abs(matrix.M32) < 0.99999 {
			y = sysMath.Atan2(-matrix.M31, matrix.M33)
			z = sysMath.Atan2(-matrix.M12, matrix.M22)
		} else {
			y = 0
			z = sysMath.Atan2(matrix.M21, matrix.M11)
		}

	case RotationOrderZYX:
		y = sysMath.Asin(-Clamp(matrix.M31, -1, 1))

		if sysMath.Abs(matrix.M31) < 0.99999 {
			x = sysMath.Atan2(matrix.M32, matrix.M33)
			z = sysMath.Atan2(matrix.M21, matrix.M11)
		} else {
			x = 0
			z = sysMath.Atan2(-matrix.M12, matrix.M22)
		}

	case RotationOrderYZX:
		z = sysMath.Asin(Clamp(matrix.M21, -1, 1))

		if sysMath.Abs(matrix.M21) < 0.99999 {
			x = sysMath.Atan2(-matrix.M23, matrix.M22)
			y = sysMath.Atan2(-matrix.M31, matrix.M11)
		} else {
			x = 0
			y = sysMath.Atan2(matrix.M13, matrix.M33)
		}

	case RotationOrderXZY:
		z = sysMath.Asin(-Clamp(matrix.M12, -1, 1))

		if sysMath.Abs(matrix.M12) < 0.99999 {
			x = sysMath.Atan2(matrix.M32, matrix.M22)
			y = sysMath.Atan2(matrix.M13, matrix.M11)
		} else {
			x = sysMath.Atan2(-matrix.M23, matrix.M33)
			y = 0
		}

	default:
		return EulerZero
	}

	return Euler{x, y, z, order}
}

////////////////////////////////////////////////////////////////////////////////

// EulerFromQuaternion creates Euler angles from a quaternion using the
// specified rotation order.
func EulerFromQuaternion(q Quaternion, order RotationOrderType) Euler {

	return EulerFromRotationMatrix(q.ToMatrix4(), order)
}

