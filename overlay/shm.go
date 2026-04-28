//go:build darwin || linux

package overlay

import (
	"encoding/binary"
	"sync/atomic"
	"unsafe"

	"github.com/dkrutsko/oasis/errors"
	"github.com/dkrutsko/oasis/shmem"
)

////////////////////////////////////////////////////////////////////////////////

const (
	// Fixed name for the POSIX shared memory segment.
	shmName = "/oasis_overlay"

	// Header occupies the first 24 bytes of the shared memory region.
	// Offset  0: writeIndex uint32 (atomic, 0 or 1)
	// Offset  4: width      uint32
	// Offset  8: height     uint32
	// Offset 12: x          int32
	// Offset 16: y          int32
	// Offset 20: dirty      uint32 (atomic)
	shmHeaderSize = 24
)

////////////////////////////////////////////////////////////////////////////////

// SharedMemory represents a double-buffered POSIX shared memory region
// used to pass RGBA pixel data between the overlay producer and a
// consumer such as Moonlight or the debug viewer.
type SharedMemory struct {
	seg     *shmem.Segment
	width   int
	height  int
	bufSize int
}

////////////////////////////////////////////////////////////////////////////////

// ShmOpen opens an existing shared memory segment created by the
// stream host (e.g. Moonlight). When readOnly is true the mapping
// is read-only. The dimensions are read from the header written
// by the segment owner.
func ShmOpen(readOnly bool) (*SharedMemory, error) {

	//----------------------------------------------------------------------------//

	seg := shmem.New(&shmem.Options{
		Name:      shmName,
		Technique: shmem.TechniqueSharedMemory,
	})

	err := seg.Open(readOnly)
	if err != nil {
		return nil, errors.New(
			"failed to open shared memory",
			errors.Error("error", err),
		)
	}

	//----------------------------------------------------------------------------//

	// Read dimensions from the header
	data := seg.GetData()
	width := int(binary.LittleEndian.Uint32(data[4:8]))
	height := int(binary.LittleEndian.Uint32(data[8:12]))

	if width <= 0 || height <= 0 {
		seg.Close()
		return nil, errors.New(
			"shared memory has invalid dimensions",
			errors.Int("width", width),
			errors.Int("height", height),
		)
	}

	return &SharedMemory{
		seg:     seg,
		width:   width,
		height:  height,
		bufSize: width * height * 4,
	}, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// Width returns the overlay width in pixels.
func (s *SharedMemory) Width() int {

	return s.width
}

////////////////////////////////////////////////////////////////////////////////

// Height returns the overlay height in pixels.
func (s *SharedMemory) Height() int {

	return s.height
}

////////////////////////////////////////////////////////////////////////////////

// GetBuffer returns the pixel buffer at the given index (0 or 1).
// The returned slice points directly into the mapped shared
// memory region.
func (s *SharedMemory) GetBuffer(index uint32) []byte {

	data := s.seg.GetData()
	offset := shmHeaderSize + int(index)*s.bufSize
	return data[offset : offset+s.bufSize]
}

////////////////////////////////////////////////////////////////////////////////

// GetInactiveIndex returns the index of the buffer that is not
// currently being read by the consumer. The producer should
// render into this buffer before calling `Flip`.
func (s *SharedMemory) GetInactiveIndex() uint32 {

	return 1 - atomic.LoadUint32(s.writeIndexPtr())
}

////////////////////////////////////////////////////////////////////////////////

// WriteBuffer returns the inactive pixel buffer that the producer
// should render into before calling `Flip`.
func (s *SharedMemory) WriteBuffer() []byte {

	return s.GetBuffer(s.GetInactiveIndex())
}

////////////////////////////////////////////////////////////////////////////////

// ReadBuffer returns the most recently completed pixel buffer.
func (s *SharedMemory) ReadBuffer() []byte {

	return s.GetBuffer(atomic.LoadUint32(s.writeIndexPtr()))
}

////////////////////////////////////////////////////////////////////////////////

// Flip atomically promotes the inactive buffer to active and sets the
// dirty flag. Call this after rendering into `WriteBuffer`.
func (s *SharedMemory) Flip() {

	idx := atomic.LoadUint32(s.writeIndexPtr())
	atomic.StoreUint32(s.writeIndexPtr(), 1-idx)
	atomic.StoreUint32(s.dirtyPtr(), 1)
}

////////////////////////////////////////////////////////////////////////////////

// IsUnlinked returns true if the segment's backing name has been
// removed while the mapping is still open. This indicates the
// stream host (e.g. Moonlight) ended the session.
func (s *SharedMemory) IsUnlinked() bool {

	if s.seg == nil {
		return false
	}

	return s.seg.IsUnlinked()
}

////////////////////////////////////////////////////////////////////////////////

// Close releases the underlying shared memory segment.
func (s *SharedMemory) Close() error {

	if s.seg != nil {
		s.seg.Close()
		s.seg = nil
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

func (s *SharedMemory) writeIndexPtr() *uint32 {

	return (*uint32)(unsafe.Pointer(&s.seg.GetData()[0]))
}

////////////////////////////////////////////////////////////////////////////////

func (s *SharedMemory) dirtyPtr() *uint32 {

	return (*uint32)(unsafe.Pointer(&s.seg.GetData()[20]))
}
