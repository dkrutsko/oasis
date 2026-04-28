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

// ShmCreate opens or creates a shared memory segment for writing and
// initializes the header with the given dimensions.
func ShmCreate(width, height int) (*SharedMemory, error) {

	//----------------------------------------------------------------------------//

	bufSize := width * height * 4
	totalSize := shmHeaderSize + bufSize*2

	seg := shmem.New(&shmem.Options{
		Name:      shmName,
		Size:      totalSize,
		Technique: shmem.TechniqueSharedMemory,
	})

	err := seg.Create()
	if err != nil {
		return nil, errors.New(
			"failed to create shared memory",
			errors.Error("error", err),
		)
	}

	//----------------------------------------------------------------------------//

	// Initialize header fields
	data := seg.GetData()
	binary.LittleEndian.PutUint32(data[0:4], 0)
	binary.LittleEndian.PutUint32(data[4:8], uint32(width))
	binary.LittleEndian.PutUint32(data[8:12], uint32(height))
	binary.LittleEndian.PutUint32(data[12:16], 0)
	binary.LittleEndian.PutUint32(data[16:20], 0)
	binary.LittleEndian.PutUint32(data[20:24], 0)

	return &SharedMemory{
		seg:     seg,
		width:   width,
		height:  height,
		bufSize: bufSize,
	}, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// ShmOpen opens an existing shared memory segment for reading. The
// dimensions are read from the header written by the producer.
func ShmOpen() (*SharedMemory, error) {

	//----------------------------------------------------------------------------//

	seg := shmem.New(&shmem.Options{
		Name:      shmName,
		Technique: shmem.TechniqueSharedMemory,
	})

	err := seg.Open(true)
	if err != nil {
		return nil, errors.New(
			"failed to open shared memory",
			errors.Error("error", err),
		)
	}

	//----------------------------------------------------------------------------//

	data := seg.GetData()
	width := int(binary.LittleEndian.Uint32(data[4:8]))
	height := int(binary.LittleEndian.Uint32(data[8:12]))

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

// WriteBuffer returns the inactive pixel buffer that the producer
// should render into before calling `Flip`.
func (s *SharedMemory) WriteBuffer() []byte {

	data := s.seg.GetData()
	idx := atomic.LoadUint32(s.writeIndexPtr())
	inactive := 1 - idx
	offset := shmHeaderSize + int(inactive)*s.bufSize
	return data[offset : offset+s.bufSize]
}

////////////////////////////////////////////////////////////////////////////////

// ReadBuffer returns the most recently completed pixel buffer.
func (s *SharedMemory) ReadBuffer() []byte {

	data := s.seg.GetData()
	idx := atomic.LoadUint32(s.writeIndexPtr())
	offset := shmHeaderSize + int(idx)*s.bufSize
	return data[offset : offset+s.bufSize]
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
