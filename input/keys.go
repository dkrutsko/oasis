//go:build darwin || linux

package input

import (
	"context"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/dkrutsko/oasis/logger"
	"github.com/dkrutsko/oasis/shm"
)

////////////////////////////////////////////////////////////////////////////////

const (
	// Fixed name for the key state shared memory segment.
	// Moonlight writes the current keyboard and mouse button
	// state here. Oasis reads it to detect hotkeys.
	keysShmName = "/oasis_keys"

	// Layout of the shared memory region:
	// Offset  0: version   uint32 (protocol version, must be 1)
	// Offset  4: scrollY   int32  (accumulated scroll ticks, atomic)
	// Offset  8: keyboard  [32]byte (256-bit bitfield)
	// Offset 40: mouse     uint8 (bit 0=left, 1=mid, 2=right, 3=x1, 4=x2)
	keysVersion      = 1
	keysScrollOffset = 4
	keysHeaderSize   = 8
	keysKeyboardSize = 32
	keysMouseOffset  = keysHeaderSize + keysKeyboardSize
	keysTotalSize    = keysMouseOffset + 1
)

////////////////////////////////////////////////////////////////////////////////

// Mouse button bit positions within the mouse byte.
const (
	KeysMouseLeft   = 0
	KeysMouseMiddle = 1
	KeysMouseRight  = 2
	KeysMouseX1     = 3
	KeysMouseX2     = 4
)

////////////////////////////////////////////////////////////////////////////////

// Keys provides read access to the key/mouse state shared
// memory segment written by Moonlight. The connection is
// managed lazily and reconnects automatically.
type Keys struct {
	seg       *shm.Segment
	mu        sync.RWMutex
	lastScroll int32
}

////////////////////////////////////////////////////////////////////////////////

// NewKeys creates a Keys reader ready to start its
// connection loop.
func NewKeys() *Keys {

	return &Keys{}
}

////////////////////////////////////////////////////////////////////////////////

// Start launches the connection management goroutine.
func (k *Keys) Start(group *errgroup.Group, ctx context.Context) {

	group.Go(func() error {
		logger.Dbg("starting keys reader")
		k.connectionLoop(ctx)
		logger.Dbg("stopping keys reader")
		return nil
	})
}

////////////////////////////////////////////////////////////////////////////////

// Close releases the shared memory segment.
func (k *Keys) Close() error {

	k.mu.Lock()
	defer k.mu.Unlock()

	if k.seg != nil {
		k.seg.Close()
		k.seg = nil
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

// GetScrollDelta reads the scroll accumulator at offset
// 4 in the shared memory and returns the change since the
// last call. Moonlight writes a cumulative scroll counter.
// The segment is read-only so we track the delta ourselves
// rather than resetting the value.
func (k *Keys) GetScrollDelta() int32 {

	k.mu.RLock()
	defer k.mu.RUnlock()

	if k.seg == nil {
		return 0
	}

	data := k.seg.GetData()
	if len(data) < keysHeaderSize {
		return 0
	}

	val := int32(data[keysScrollOffset]) |
		int32(data[keysScrollOffset+1])<<8 |
		int32(data[keysScrollOffset+2])<<16 |
		int32(data[keysScrollOffset+3])<<24

	delta := val - k.lastScroll
	k.lastScroll = val
	return delta
}

////////////////////////////////////////////////////////////////////////////////

// IsMouseDown returns true if the given mouse button bit
// is currently pressed.
func (k *Keys) IsMouseDown(button uint8) bool {

	k.mu.RLock()
	defer k.mu.RUnlock()

	if k.seg == nil {
		return false
	}

	data := k.seg.GetData()
	if len(data) < keysTotalSize {
		return false
	}

	return data[keysMouseOffset]&(1<<button) != 0
}

////////////////////////////////////////////////////////////////////////////////

// IsKeyDown returns true if the given scancode is
// currently pressed.
func (k *Keys) IsKeyDown(scancode uint8) bool {

	k.mu.RLock()
	defer k.mu.RUnlock()

	if k.seg == nil {
		return false
	}

	data := k.seg.GetData()
	if len(data) < keysTotalSize {
		return false
	}

	byteIdx := keysHeaderSize + int(scancode/8)
	bitIdx := scancode % 8
	return data[byteIdx]&(1<<bitIdx) != 0
}

////////////////////////////////////////////////////////////////////////////////

func (k *Keys) connectionLoop(ctx context.Context) {

	var lastCheck time.Time

	for {
		if ctx.Err() != nil {
			return
		}

		//--------------------------------------------------------------------//

		k.mu.RLock()
		connected := k.seg != nil
		k.mu.RUnlock()

		if !connected {
			seg := shm.New(&shm.Options{
				Name:      keysShmName,
				Technique: shm.TechniqueSharedMemory,
			})

			err := seg.Open(true)
			if err != nil {
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Second):
				}
				continue
			}

			k.mu.Lock()
			k.seg = seg
			k.mu.Unlock()

			lastCheck = time.Now()
			logger.Dbg("keys reader attached")
		}

		//--------------------------------------------------------------------//

		if time.Since(lastCheck) > shmCheckInterval {
			lastCheck = time.Now()

			k.mu.RLock()
			unlinked := k.seg != nil && k.seg.IsUnlinked()
			k.mu.RUnlock()

			if unlinked {
				logger.Dbg("keys segment unlinked, waiting for reconnect")

				k.mu.Lock()
				k.seg.Close()
				k.seg = nil
				k.mu.Unlock()

				continue
			}
		}

		//--------------------------------------------------------------------//

		select {
		case <-ctx.Done():
			return
		case <-time.After(shmCheckInterval):
		}
	}
}
