//go:build darwin || linux

package input

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
	shmName = "/oasis_input"

	// Header occupies the first 16 bytes of the shared memory
	// region.
	// Offset  0: capacity  uint32 (number of entry slots)
	// Offset  4: entrySize uint32 (sizeof entry, must be 8)
	// Offset  8: writePos  uint32 (atomic, producer)
	// Offset 12: readPos   uint32 (atomic, consumer)
	shmHeaderSize = 16

	// Expected size of each entry in the ring buffer.
	shmEntrySize = 8
)

////////////////////////////////////////////////////////////////////////////////

// InputQueue provides access to the shared memory ring buffer
// used to send input events to Moonlight. Moonlight creates
// the segment when a stream starts and unlinks it when the
// stream ends.
type InputQueue struct {
	seg      *shm.Segment
	capacity uint32
}

////////////////////////////////////////////////////////////////////////////////

// ShmOpen opens the existing input queue shared memory
// segment created by Moonlight. Returns an error if the
// segment does not exist or the header is invalid.
func ShmOpen() (*InputQueue, error) {

	//----------------------------------------------------------------------------//

	seg := shm.New(&shm.Options{
		Name:      shmName,
		Technique: shm.TechniqueSharedMemory,
	})

	err := seg.Open(false)
	if err != nil {
		return nil, errors.New(
			"failed to open input shared memory",
			errors.Error("error", err),
		)
	}

	//----------------------------------------------------------------------------//

	data := seg.GetData()
	if len(data) < shmHeaderSize {
		seg.Close()
		return nil, errors.New(
			"input shared memory segment too small",
		)
	}

	capacity := binary.LittleEndian.Uint32(data[0:4])
	entrySize := binary.LittleEndian.Uint32(data[4:8])

	if capacity == 0 {
		seg.Close()
		return nil, errors.New(
			"input queue has zero capacity",
		)
	}

	if entrySize != shmEntrySize {
		seg.Close()
		return nil, errors.New(
			"input queue entry size mismatch",
			errors.Uint32("expected", shmEntrySize),
			errors.Uint32("actual", entrySize),
		)
	}

	//----------------------------------------------------------------------------//

	return &InputQueue{
		seg:      seg,
		capacity: capacity,
	}, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// Enqueue writes an input entry to the next available slot
// in the ring buffer. Returns false if the queue is full.
func (q *InputQueue) Enqueue(entry InputEntry) bool {

	writePos := atomic.LoadUint32(q.writePosPtr())
	readPos := atomic.LoadUint32(q.readPosPtr())

	// Queue is full
	if writePos-readPos >= q.capacity {
		return false
	}

	// Write entry to the slot
	slot := writePos % q.capacity
	q.writeEntry(slot, entry)

	// Advance writePos with a store-release so the
	// consumer sees the entry before the updated position
	atomic.StoreUint32(q.writePosPtr(), writePos+1)
	return true
}

////////////////////////////////////////////////////////////////////////////////

// IsUnlinked returns true if the segment's backing name has
// been removed. This indicates the stream host ended the
// session.
func (q *InputQueue) IsUnlinked() bool {

	if q.seg == nil {
		return false
	}

	return q.seg.IsUnlinked()
}

////////////////////////////////////////////////////////////////////////////////

// Close releases the underlying shared memory segment.
func (q *InputQueue) Close() error {

	if q.seg != nil {
		q.seg.Close()
		q.seg = nil
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

func (q *InputQueue) writePosPtr() *uint32 {

	return (*uint32)(unsafe.Pointer(&q.seg.GetData()[8]))
}

////////////////////////////////////////////////////////////////////////////////

func (q *InputQueue) readPosPtr() *uint32 {

	return (*uint32)(unsafe.Pointer(&q.seg.GetData()[12]))
}

////////////////////////////////////////////////////////////////////////////////

func (q *InputQueue) writeEntry(slot uint32, entry InputEntry) {

	data := q.seg.GetData()
	offset := shmHeaderSize + slot*shmEntrySize

	data[offset+0] = entry.Type
	data[offset+1] = entry.Flags
	data[offset+2] = byte(entry.X)
	data[offset+3] = byte(entry.X >> 8)
	data[offset+4] = byte(entry.Y)
	data[offset+5] = byte(entry.Y >> 8)
	data[offset+6] = byte(entry.Reserved)
	data[offset+7] = byte(entry.Reserved >> 8)
}
