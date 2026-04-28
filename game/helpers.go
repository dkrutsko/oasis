package game

import (
	"encoding/binary"
	sysMath "math"

	"github.com/dkrutsko/oasis/errors"
	"github.com/dkrutsko/oasis/leech"
	"github.com/dkrutsko/oasis/math"
)

////////////////////////////////////////////////////////////////////////////////

func ReadVector2(memory *leech.Memory, address uintptr) (math.Vector2, error) {

	res, err := ReadVector2s(memory, address, 1, 0)
	if err != nil {
		return math.Vector2Zero, err
	}

	return res[0], nil
}

////////////////////////////////////////////////////////////////////////////////

func ReadVector2s(
	memory *leech.Memory,
	address uintptr,
	count, stride uint32,
) ([]math.Vector2, error) {

	res, err := memory.ReadTypes(address, leech.MemTypeBuffer, 4*2, count, stride)
	if err != nil {
		return nil, err
	}

	if len(res) != int(count) {
		return nil, errors.New("not enough results for read type")
	}

	out := make([]math.Vector2, count)
	for i, v := range res {
		val, ok := v.([]byte)
		if !ok {
			return nil, errors.New("failed to cast to final type")
		}

		x := sysMath.Float32frombits(binary.LittleEndian.Uint32(val[0:4]))
		y := sysMath.Float32frombits(binary.LittleEndian.Uint32(val[4:8]))

		out[i] = math.Vector2{
			X: float64(x),
			Y: float64(y),
		}
	}

	return out, nil
}

////////////////////////////////////////////////////////////////////////////////

func ReadVector3(memory *leech.Memory, address uintptr) (math.Vector3, error) {

	res, err := ReadVector3s(memory, address, 1, 0)
	if err != nil {
		return math.Vector3Zero, err
	}

	return res[0], nil
}

////////////////////////////////////////////////////////////////////////////////

func ReadVector3s(
	memory *leech.Memory,
	address uintptr,
	count, stride uint32,
) ([]math.Vector3, error) {

	res, err := memory.ReadTypes(address, leech.MemTypeBuffer, 4*3, count, stride)
	if err != nil {
		return nil, err
	}

	if len(res) != int(count) {
		return nil, errors.New("not enough results for read type")
	}

	out := make([]math.Vector3, count)
	for i, v := range res {
		val, ok := v.([]byte)
		if !ok {
			return nil, errors.New("failed to cast to final type")
		}

		x := sysMath.Float32frombits(binary.LittleEndian.Uint32(val[0:4]))
		y := sysMath.Float32frombits(binary.LittleEndian.Uint32(val[4:8]))
		z := sysMath.Float32frombits(binary.LittleEndian.Uint32(val[8:12]))

		out[i] = math.Vector3{
			X: float64(x),
			Y: float64(y),
			Z: float64(z),
		}
	}

	return out, nil
}

////////////////////////////////////////////////////////////////////////////////

func ReadVector4(memory *leech.Memory, address uintptr) (math.Vector4, error) {

	res, err := ReadVector4s(memory, address, 1, 0)
	if err != nil {
		return math.Vector4Zero, err
	}

	return res[0], nil
}

////////////////////////////////////////////////////////////////////////////////

func ReadVector4s(
	memory *leech.Memory,
	address uintptr,
	count, stride uint32,
) ([]math.Vector4, error) {

	res, err := memory.ReadTypes(address, leech.MemTypeBuffer, 4*4, count, stride)
	if err != nil {
		return nil, err
	}

	if len(res) != int(count) {
		return nil, errors.New("not enough results for read type")
	}

	out := make([]math.Vector4, count)
	for i, v := range res {
		val, ok := v.([]byte)
		if !ok {
			return nil, errors.New("failed to cast to final type")
		}

		x := sysMath.Float32frombits(binary.LittleEndian.Uint32(val[0:4]))
		y := sysMath.Float32frombits(binary.LittleEndian.Uint32(val[4:8]))
		z := sysMath.Float32frombits(binary.LittleEndian.Uint32(val[8:12]))
		w := sysMath.Float32frombits(binary.LittleEndian.Uint32(val[12:16]))

		out[i] = math.Vector4{
			X: float64(x),
			Y: float64(y),
			Z: float64(z),
			W: float64(w),
		}
	}

	return out, nil
}

////////////////////////////////////////////////////////////////////////////////

func ReadQuaternion(memory *leech.Memory, address uintptr) (math.Quaternion, error) {

	res, err := ReadQuaternions(memory, address, 1, 0)
	if err != nil {
		return math.QuaternionZero, err
	}

	return res[0], nil
}

////////////////////////////////////////////////////////////////////////////////

func ReadQuaternions(
	memory *leech.Memory,
	address uintptr,
	count, stride uint32,
) ([]math.Quaternion, error) {

	res, err := memory.ReadTypes(address, leech.MemTypeBuffer, 4*4, count, stride)
	if err != nil {
		return nil, err
	}

	if len(res) != int(count) {
		return nil, errors.New("not enough results for read type")
	}

	out := make([]math.Quaternion, count)
	for i, v := range res {
		val, ok := v.([]byte)
		if !ok {
			return nil, errors.New("failed to cast to final type")
		}

		x := sysMath.Float32frombits(binary.LittleEndian.Uint32(val[0:4]))
		y := sysMath.Float32frombits(binary.LittleEndian.Uint32(val[4:8]))
		z := sysMath.Float32frombits(binary.LittleEndian.Uint32(val[8:12]))
		w := sysMath.Float32frombits(binary.LittleEndian.Uint32(val[12:16]))

		out[i] = math.Quaternion{
			X: float64(x),
			Y: float64(y),
			Z: float64(z),
			W: float64(w),
		}
	}

	return out, nil
}

////////////////////////////////////////////////////////////////////////////////

func ReadMatrix3(memory *leech.Memory, address uintptr) (math.Matrix3, error) {

	res, err := ReadMatrix3s(memory, address, 1, 0)
	if err != nil {
		return math.Matrix3Zero, err
	}

	return res[0], nil
}

////////////////////////////////////////////////////////////////////////////////

func ReadMatrix3s(
	memory *leech.Memory,
	address uintptr,
	count, stride uint32,
) ([]math.Matrix3, error) {

	res, err := memory.ReadTypes(address, leech.MemTypeBuffer, 4*9, count, stride)
	if err != nil {
		return nil, err
	}

	if len(res) != int(count) {
		return nil, errors.New("not enough results for read type")
	}

	out := make([]math.Matrix3, count)
	for i, v := range res {
		val, ok := v.([]byte)
		if !ok {
			return nil, errors.New("failed to cast to final type")
		}

		// Decode the buffer data
		var data [9]float32
		for j := 0; j < 9; j++ {
			bits := binary.LittleEndian.Uint32(val[j*4 : j*4+4])
			data[j] = sysMath.Float32frombits(bits)
		}

		out[i] = math.Matrix3{
			M11: float64(data[0]), M12: float64(data[1]), M13: float64(data[2]),
			M21: float64(data[3]), M22: float64(data[4]), M23: float64(data[5]),
			M31: float64(data[6]), M32: float64(data[7]), M33: float64(data[8]),
		}
	}

	return out, nil
}

////////////////////////////////////////////////////////////////////////////////

func ReadMatrix4(memory *leech.Memory, address uintptr) (math.Matrix4, error) {

	res, err := ReadMatrix4s(memory, address, 1, 0)
	if err != nil {
		return math.Matrix4Zero, err
	}

	return res[0], nil
}

////////////////////////////////////////////////////////////////////////////////

func ReadMatrix4s(
	memory *leech.Memory,
	address uintptr,
	count, stride uint32,
) ([]math.Matrix4, error) {

	res, err := memory.ReadTypes(address, leech.MemTypeBuffer, 4*16, count, stride)
	if err != nil {
		return nil, err
	}

	if len(res) != int(count) {
		return nil, errors.New("not enough results for read type")
	}

	out := make([]math.Matrix4, count)
	for i, v := range res {
		val, ok := v.([]byte)
		if !ok {
			return nil, errors.New("failed to cast to final type")
		}

		// Decode the buffer data
		var data [16]float32
		for j := 0; j < 16; j++ {
			bits := binary.LittleEndian.Uint32(val[j*4 : j*4+4])
			data[j] = sysMath.Float32frombits(bits)
		}

		out[i] = math.Matrix4{
			M11: float64(data[0x0]), M12: float64(data[0x1]), M13: float64(data[0x2]), M14: float64(data[0x3]),
			M21: float64(data[0x4]), M22: float64(data[0x5]), M23: float64(data[0x6]), M24: float64(data[0x7]),
			M31: float64(data[0x8]), M32: float64(data[0x9]), M33: float64(data[0xA]), M34: float64(data[0xB]),
			M41: float64(data[0xC]), M42: float64(data[0xD]), M43: float64(data[0xE]), M44: float64(data[0xF]),
		}
	}

	return out, nil
}

////////////////////////////////////////////////////////////////////////////////

func WriteVector2(memory *leech.Memory, address uintptr, value math.Vector2) error {

	var buf [4 * 2]byte
	binary.LittleEndian.PutUint32(buf[0:4], sysMath.Float32bits(float32(value.X)))
	binary.LittleEndian.PutUint32(buf[4:8], sysMath.Float32bits(float32(value.Y)))

	return memory.WriteData(address, buf[:])
}

////////////////////////////////////////////////////////////////////////////////

func WriteVector2s(
	memory *leech.Memory,
	address uintptr,
	values []math.Vector2,
	stride uint32,
) error {

	step := uintptr(4*2 + stride)
	for i, value := range values {
		err := WriteVector2(memory, address+uintptr(i)*step, value)
		if err != nil {
			return err
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

func WriteVector3(memory *leech.Memory, address uintptr, value math.Vector3) error {

	var buf [4 * 3]byte
	binary.LittleEndian.PutUint32(buf[0:4], sysMath.Float32bits(float32(value.X)))
	binary.LittleEndian.PutUint32(buf[4:8], sysMath.Float32bits(float32(value.Y)))
	binary.LittleEndian.PutUint32(buf[8:12], sysMath.Float32bits(float32(value.Z)))

	return memory.WriteData(address, buf[:])
}

////////////////////////////////////////////////////////////////////////////////

func WriteVector3s(
	memory *leech.Memory,
	address uintptr,
	values []math.Vector3,
	stride uint32,
) error {

	step := uintptr(4*3 + stride)
	for i, value := range values {
		err := WriteVector3(memory, address+uintptr(i)*step, value)
		if err != nil {
			return err
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

func WriteVector4(memory *leech.Memory, address uintptr, value math.Vector4) error {

	var buf [4 * 4]byte
	binary.LittleEndian.PutUint32(buf[0:4], sysMath.Float32bits(float32(value.X)))
	binary.LittleEndian.PutUint32(buf[4:8], sysMath.Float32bits(float32(value.Y)))
	binary.LittleEndian.PutUint32(buf[8:12], sysMath.Float32bits(float32(value.Z)))
	binary.LittleEndian.PutUint32(buf[12:16], sysMath.Float32bits(float32(value.W)))

	return memory.WriteData(address, buf[:])
}

////////////////////////////////////////////////////////////////////////////////

func WriteVector4s(
	memory *leech.Memory,
	address uintptr,
	values []math.Vector4,
	stride uint32,
) error {

	step := uintptr(4*4 + stride)
	for i, value := range values {
		err := WriteVector4(memory, address+uintptr(i)*step, value)
		if err != nil {
			return err
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

func WriteQuaternion(memory *leech.Memory, address uintptr, value math.Quaternion) error {

	var buf [4 * 4]byte
	binary.LittleEndian.PutUint32(buf[0:4], sysMath.Float32bits(float32(value.X)))
	binary.LittleEndian.PutUint32(buf[4:8], sysMath.Float32bits(float32(value.Y)))
	binary.LittleEndian.PutUint32(buf[8:12], sysMath.Float32bits(float32(value.Z)))
	binary.LittleEndian.PutUint32(buf[12:16], sysMath.Float32bits(float32(value.W)))

	return memory.WriteData(address, buf[:])
}

////////////////////////////////////////////////////////////////////////////////

func WriteQuaternions(
	memory *leech.Memory,
	address uintptr,
	values []math.Quaternion,
	stride uint32,
) error {

	step := uintptr(4*4 + stride)
	for i, value := range values {
		err := WriteQuaternion(memory, address+uintptr(i)*step, value)
		if err != nil {
			return err
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

func WriteMatrix3(memory *leech.Memory, address uintptr, value math.Matrix3) error {

	var buf [4 * 9]byte
	binary.LittleEndian.PutUint32(buf[0:4], sysMath.Float32bits(float32(value.M11)))
	binary.LittleEndian.PutUint32(buf[4:8], sysMath.Float32bits(float32(value.M12)))
	binary.LittleEndian.PutUint32(buf[8:12], sysMath.Float32bits(float32(value.M13)))
	binary.LittleEndian.PutUint32(buf[12:16], sysMath.Float32bits(float32(value.M21)))
	binary.LittleEndian.PutUint32(buf[16:20], sysMath.Float32bits(float32(value.M22)))
	binary.LittleEndian.PutUint32(buf[20:24], sysMath.Float32bits(float32(value.M23)))
	binary.LittleEndian.PutUint32(buf[24:28], sysMath.Float32bits(float32(value.M31)))
	binary.LittleEndian.PutUint32(buf[28:32], sysMath.Float32bits(float32(value.M32)))
	binary.LittleEndian.PutUint32(buf[32:36], sysMath.Float32bits(float32(value.M33)))

	return memory.WriteData(address, buf[:])
}

////////////////////////////////////////////////////////////////////////////////

func WriteMatrix3s(
	memory *leech.Memory,
	address uintptr,
	values []math.Matrix3,
	stride uint32,
) error {

	step := uintptr(4*9 + stride)
	for i, value := range values {
		err := WriteMatrix3(memory, address+uintptr(i)*step, value)
		if err != nil {
			return err
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

func WriteMatrix4(memory *leech.Memory, address uintptr, value math.Matrix4) error {

	var buf [4 * 16]byte
	binary.LittleEndian.PutUint32(buf[0:4], sysMath.Float32bits(float32(value.M11)))
	binary.LittleEndian.PutUint32(buf[4:8], sysMath.Float32bits(float32(value.M12)))
	binary.LittleEndian.PutUint32(buf[8:12], sysMath.Float32bits(float32(value.M13)))
	binary.LittleEndian.PutUint32(buf[12:16], sysMath.Float32bits(float32(value.M14)))
	binary.LittleEndian.PutUint32(buf[16:20], sysMath.Float32bits(float32(value.M21)))
	binary.LittleEndian.PutUint32(buf[20:24], sysMath.Float32bits(float32(value.M22)))
	binary.LittleEndian.PutUint32(buf[24:28], sysMath.Float32bits(float32(value.M23)))
	binary.LittleEndian.PutUint32(buf[28:32], sysMath.Float32bits(float32(value.M24)))
	binary.LittleEndian.PutUint32(buf[32:36], sysMath.Float32bits(float32(value.M31)))
	binary.LittleEndian.PutUint32(buf[36:40], sysMath.Float32bits(float32(value.M32)))
	binary.LittleEndian.PutUint32(buf[40:44], sysMath.Float32bits(float32(value.M33)))
	binary.LittleEndian.PutUint32(buf[44:48], sysMath.Float32bits(float32(value.M34)))
	binary.LittleEndian.PutUint32(buf[48:52], sysMath.Float32bits(float32(value.M41)))
	binary.LittleEndian.PutUint32(buf[52:56], sysMath.Float32bits(float32(value.M42)))
	binary.LittleEndian.PutUint32(buf[56:60], sysMath.Float32bits(float32(value.M43)))
	binary.LittleEndian.PutUint32(buf[60:64], sysMath.Float32bits(float32(value.M44)))

	return memory.WriteData(address, buf[:])
}

////////////////////////////////////////////////////////////////////////////////

func WriteMatrix4s(
	memory *leech.Memory,
	address uintptr,
	values []math.Matrix4,
	stride uint32,
) error {

	step := uintptr(4*16 + stride)
	for i, value := range values {
		err := WriteMatrix4(memory, address+uintptr(i)*step, value)
		if err != nil {
			return err
		}
	}

	return nil
}

