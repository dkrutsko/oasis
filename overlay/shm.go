//go:build darwin || linux

package overlay

import (
	"encoding/binary"
	"sync/atomic"
	"unsafe"

	"github.com/dkrutsko/oasis/errors"
	"github.com/dkrutsko/oasis/shm"
)

////////////////////////////////////////////////////////////////////////////////

const (
	// Fixed name for the POSIX shared memory segment.
	shmName = "/oasis_overlay"

	// Header occupies the first 32 bytes of the shared memory
	// region. Moonlight creates the segment and writes the
	// dimensions. The ready/reading fields are used for
	// lock-free triple buffering.
	//
	// Offset  0: width     uint32
	// Offset  4: height    uint32
	// Offset  8: ready     uint32 (atomic, 0-2, producer sets after render)
	// Offset 12: reading   uint32 (atomic, 0-2 or 0xFF, consumer sets while reading)
	// Offset 16: reserved  [16]byte
	shmHeaderSize = 32

	// Number of pixel buffers for triple buffering.
	shmBufferCount = 3

	// Sentinel value indicating the consumer is not
	// currently reading any buffer.
	shmReadingNone uint32 = 0xFF
)

////////////////////////////////////////////////////////////////////////////////

// SharedMemory represents a triple-buffered POSIX shared memory
// region used to pass RGBA pixel data between the overlay
// producer (oasis) and consumer (Moonlight).
type SharedMemory struct {
	seg     *shm.Segment
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

	seg := shm.New(&shm.Options{
		Name:      shmName,
		Technique: shm.TechniqueSharedMemory,
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
	width := int(binary.LittleEndian.Uint32(data[0:4]))
	height := int(binary.LittleEndian.Uint32(data[4:8]))

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

// GetBuffer returns the pixel buffer at the given index (0, 1,
// or 2). The returned slice points directly into the mapped
// shared memory region.
func (s *SharedMemory) GetBuffer(index uint32) []byte {

	data := s.seg.GetData()
	offset := shmHeaderSize + int(index)*s.bufSize
	return data[offset : offset+s.bufSize]
}

////////////////////////////////////////////////////////////////////////////////

// GetWriteIndex returns the index of a buffer that is safe for
// the producer to write into. This is a buffer that is neither
// the latest completed frame (ready) nor the one the consumer
// is currently reading.
func (s *SharedMemory) GetWriteIndex() uint32 {

	ready := atomic.LoadUint32(s.readyPtr())
	reading := atomic.LoadUint32(s.readingPtr())

	for i := uint32(0); i < shmBufferCount; i++ {
		if i != ready && i != reading {
			return i
		}
	}

	// Fallback: if ready == reading, two buffers are free.
	// Pick one that isn't ready.
	if ready == 0 {
		return 1
	}
	return 0
}

////////////////////////////////////////////////////////////////////////////////

// WriteBuffer returns a pixel buffer that the producer can
// safely render into without conflicting with the consumer.
func (s *SharedMemory) WriteBuffer() []byte {

	return s.GetBuffer(s.GetWriteIndex())
}

////////////////////////////////////////////////////////////////////////////////

// Publish atomically marks the given buffer index as the
// latest completed frame. Call this after rendering into the
// buffer returned by `GetWriteIndex`.
func (s *SharedMemory) Publish(index uint32) {

	atomic.StoreUint32(s.readyPtr(), index)
}

////////////////////////////////////////////////////////////////////////////////

// ReadBuffer returns the most recently completed pixel buffer.
func (s *SharedMemory) ReadBuffer() []byte {

	return s.GetBuffer(atomic.LoadUint32(s.readyPtr()))
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

func (s *SharedMemory) readyPtr() *uint32 {

	return (*uint32)(unsafe.Pointer(&s.seg.GetData()[8]))
}

////////////////////////////////////////////////////////////////////////////////

func (s *SharedMemory) readingPtr() *uint32 {

	return (*uint32)(unsafe.Pointer(&s.seg.GetData()[12]))
}
