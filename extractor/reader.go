package main

import (
	"bytes"
	"encoding/binary"

	"github.com/dkrutsko/oasis/errors"
)

////////////////////////////////////////////////////////////////////////////////

// binaryReader reads little-endian values from the front of a
// byte slice. Every read checks the bytes left, so truncated
// data returns an error rather than panicking.
type binaryReader struct {
	data []byte
	pos  int
}

////////////////////////////////////////////////////////////////////////////////

// GetPosition returns the number of bytes read so far.
func (r *binaryReader) GetPosition() int {
	return r.pos
}

////////////////////////////////////////////////////////////////////////////////

// GetRemaining returns the number of bytes left to read.
func (r *binaryReader) GetRemaining() int {
	return len(r.data) - r.pos
}

////////////////////////////////////////////////////////////////////////////////

// Align skips ahead to the next multiple of `alignment`, which
// must be a power of two. It stops at the end of the data.
func (r *binaryReader) Align(alignment int) {

	r.pos = (r.pos + alignment - 1) &^ (alignment - 1)

	if r.pos > len(r.data) {
		r.pos = len(r.data)
	}
}

////////////////////////////////////////////////////////////////////////////////

// ReadBytes returns the next `n` bytes without copying them.
func (r *binaryReader) ReadBytes(n int) ([]byte, error) {

	if n < 0 || n > r.GetRemaining() {
		return nil, errors.New(
			"unexpected end of data",
			errors.Int("size", n),
			errors.Int("remaining", r.GetRemaining()),
		)
	}

	data := r.data[r.pos : r.pos+n]
	r.pos += n

	return data, nil
}

////////////////////////////////////////////////////////////////////////////////

// ReadUint8 reads a single byte.
func (r *binaryReader) ReadUint8() (uint8, error) {

	data, err := r.ReadBytes(1)
	if err != nil {
		return 0, err
	}

	return data[0], nil
}

////////////////////////////////////////////////////////////////////////////////

// ReadUint16 reads a little-endian `uint16`.
func (r *binaryReader) ReadUint16() (uint16, error) {

	data, err := r.ReadBytes(2)
	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint16(data), nil
}

////////////////////////////////////////////////////////////////////////////////

// ReadUint32 reads a little-endian `uint32`.
func (r *binaryReader) ReadUint32() (uint32, error) {

	data, err := r.ReadBytes(4)
	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint32(data), nil
}

////////////////////////////////////////////////////////////////////////////////

// ReadUint64 reads a little-endian `uint64`.
func (r *binaryReader) ReadUint64() (uint64, error) {

	data, err := r.ReadBytes(8)
	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint64(data), nil
}

////////////////////////////////////////////////////////////////////////////////

// ReadInt32 reads a little-endian `int32`.
func (r *binaryReader) ReadInt32() (int32, error) {

	value, err := r.ReadUint32()
	if err != nil {
		return 0, err
	}

	return int32(value), nil
}

////////////////////////////////////////////////////////////////////////////////

// ReadString reads a null-terminated string.
func (r *binaryReader) ReadString() (string, error) {

	end := bytes.IndexByte(r.data[r.pos:], 0)
	if end < 0 {
		return "", errors.New(
			"string is not null-terminated",
			errors.Int("position", r.pos),
		)
	}

	value := string(r.data[r.pos : r.pos+end])
	r.pos += end + 1

	return value, nil
}
