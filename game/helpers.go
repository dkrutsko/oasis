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

func ReadMatrix(memory *leech.Memory, address uintptr) (math.Matrix, error) {

	res, err := ReadMatrices(memory, address, 1, 0)
	if err != nil {
		return math.MatrixZero, err
	}

	return res[0], nil
}

////////////////////////////////////////////////////////////////////////////////

func ReadMatrices(
	memory *leech.Memory,
	address uintptr,
	count, stride uint32,
) ([]math.Matrix, error) {

	res, err := memory.ReadTypes(address, leech.MemTypeBuffer, 4*16, count, stride)
	if err != nil {
		return nil, err
	}

	if len(res) != int(count) {
		return nil, errors.New("not enough results for read type")
	}

	out := make([]math.Matrix, count)
	for i, v := range res {
		val, ok := v.([]byte)
		if !ok {
			return nil, errors.New("failed to cast to final type")
		}

		// Decode the buffer data
		data := make([]float32, 16)
		for j := 0; j < 16; j++ {
			bits := binary.LittleEndian.Uint32(val[j*4 : j*4+4])
			data[j] = sysMath.Float32frombits(bits)
		}

		out[i] = math.Matrix{
			float64(data[0x0]), float64(data[0x1]), float64(data[0x2]), float64(data[0x3]),
			float64(data[0x4]), float64(data[0x5]), float64(data[0x6]), float64(data[0x7]),
			float64(data[0x8]), float64(data[0x9]), float64(data[0xA]), float64(data[0xB]),
			float64(data[0xC]), float64(data[0xD]), float64(data[0xE]), float64(data[0xF]),
		}
	}

	return out, nil
}
