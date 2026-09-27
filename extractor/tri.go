package main

import (
	"bufio"
	"encoding/binary"
	sysMath "math"
	"os"

	"github.com/dkrutsko/oasis/errors"
	"github.com/dkrutsko/oasis/math"
)

////////////////////////////////////////////////////////////////////////////////

// Size of a single triangle in the .tri file format. Each
// triangle is 9 consecutive float32 values (3 vertices x 3
// components) stored in little-endian, as `maps.Load` reads.
const triSize = 36

////////////////////////////////////////////////////////////////////////////////

// triWriter writes triangles to a temporary file that replaces
// the `.tri` file on `Commit`, so a failed or interrupted run
// never leaves a partial `.tri` file behind.
type triWriter struct {
	path   string
	file   *os.File
	buffer *bufio.Writer
	count  int

	// Reused for every triangle, since a local array would be
	// moved to the heap on each call
	triangle [triSize]byte
}

////////////////////////////////////////////////////////////////////////////////

// createTriWriter creates the temporary file for a `.tri` file
// at the given path.
func createTriWriter(path string) (*triWriter, error) {

	file, err := os.Create(path + ".tmp")
	if err != nil {
		return nil, errors.New(
			"failed to create tri file",
			errors.String("path", path),
			errors.Error("error", err),
		)
	}

	return &triWriter{
		path:   path,
		file:   file,
		buffer: bufio.NewWriterSize(file, 1<<20),
	}, nil
}

////////////////////////////////////////////////////////////////////////////////

// GetCount returns the number of triangles written.
func (w *triWriter) GetCount() int {
	return w.count
}

////////////////////////////////////////////////////////////////////////////////

// Write adds a triangle. The buffer keeps the first write error,
// which `Commit` returns.
func (w *triWriter) Write(v0, v1, v2 math.Vector3) {

	encodeVec3(w.triangle[0:12], v0)
	encodeVec3(w.triangle[12:24], v1)
	encodeVec3(w.triangle[24:36], v2)

	w.buffer.Write(w.triangle[:])
	w.count++
}

////////////////////////////////////////////////////////////////////////////////

// Commit writes out the remaining triangles and moves the file
// into place. The temporary file is removed if anything fails.
func (w *triWriter) Commit() error {

	err := w.buffer.Flush()
	if err != nil {
		w.Abort()
		return errors.New(
			"failed to write tri file",
			errors.String("path", w.path),
			errors.Error("error", err),
		)
	}

	err = w.file.Close()
	if err != nil {
		os.Remove(w.file.Name())
		return errors.New(
			"failed to close tri file",
			errors.String("path", w.path),
			errors.Error("error", err),
		)
	}

	err = os.Rename(w.file.Name(), w.path)
	if err != nil {
		os.Remove(w.file.Name())
		return errors.New(
			"failed to move tri file into place",
			errors.String("path", w.path),
			errors.Error("error", err),
		)
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

// Abort closes and removes the temporary file.
func (w *triWriter) Abort() {

	w.file.Close()
	os.Remove(w.file.Name())
}

////////////////////////////////////////////////////////////////////////////////

func encodeVec3(data []byte, v math.Vector3) {

	binary.LittleEndian.PutUint32(data[0:4], sysMath.Float32bits(float32(v.X)))
	binary.LittleEndian.PutUint32(data[4:8], sysMath.Float32bits(float32(v.Y)))
	binary.LittleEndian.PutUint32(data[8:12], sysMath.Float32bits(float32(v.Z)))
}
