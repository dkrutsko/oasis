//go:build darwin || linux

package input

import (
	"context"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/dkrutsko/oasis/logger"
)

////////////////////////////////////////////////////////////////////////////////

// Command types for input entries.
const (
	TypeKeyDown         uint8 = 0x01
	TypeKeyUp           uint8 = 0x02
	TypeMouseMoveRel    uint8 = 0x10
	TypeMouseMoveAbs    uint8 = 0x11
	TypeMouseButtonDown uint8 = 0x20
	TypeMouseButtonUp   uint8 = 0x21
	TypeMouseScrollV    uint8 = 0x30
	TypeMouseScrollH    uint8 = 0x31
)

////////////////////////////////////////////////////////////////////////////////

// Keyboard modifier flags (bitwise OR).
const (
	ModShift uint8 = 0x01
	ModCtrl  uint8 = 0x02
	ModAlt   uint8 = 0x04
	ModMeta  uint8 = 0x08
)

////////////////////////////////////////////////////////////////////////////////

// How often to check if the shared memory segment still
// exists. Moonlight unlinks it when the stream ends.
const shmCheckInterval = 2 * time.Second

////////////////////////////////////////////////////////////////////////////////

// InputEntry represents a single input event in the shared
// memory ring buffer.
type InputEntry struct {
	Type     uint8
	Flags    uint8
	X        int16
	Y        int16
	Reserved uint16
}

////////////////////////////////////////////////////////////////////////////////

// Input provides input synthesis via a shared memory ring
// buffer created by Moonlight. Keyboard and mouse events
// are enqueued and processed by Moonlight on the target
// machine. The connection to shared memory is managed
// lazily by the `Start` goroutine.
type Input struct {
	queue *InputQueue
	mu    sync.RWMutex
}

////////////////////////////////////////////////////////////////////////////////

// NewInput creates an Input ready to start its connection
// loop. The shared memory connection is established lazily
// once Moonlight creates the segment.
func NewInput() *Input {

	return &Input{}
}

////////////////////////////////////////////////////////////////////////////////

// Start launches the connection management goroutine. It
// connects to the shared memory segment when available and
// reconnects if Moonlight ends and restarts the stream.
func (i *Input) Start(group *errgroup.Group, ctx context.Context) {

	group.Go(func() error {
		logger.Dbg("starting input manager")
		i.connectionLoop(ctx)
		logger.Dbg("stopping input manager")
		return nil
	})
}

////////////////////////////////////////////////////////////////////////////////

// Close releases the shared memory segment.
func (i *Input) Close() error {

	i.mu.Lock()
	defer i.mu.Unlock()

	if i.queue != nil {
		i.queue.Close()
		i.queue = nil
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

// IsConnected returns true if the input queue is attached
// to the shared memory segment.
func (i *Input) IsConnected() bool {

	i.mu.RLock()
	defer i.mu.RUnlock()

	return i.queue != nil
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//                                 Keyboard                                   //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// KeyClick enqueues a key press followed by a key release.
// Returns false if either enqueue fails.
func (i *Input) KeyClick(key Key) bool {

	if !i.KeyPress(key) {
		return false
	}

	return i.KeyRelease(key)
}

////////////////////////////////////////////////////////////////////////////////

// KeyPress enqueues a key-down event. Returns false if the
// queue is not connected or full.
func (i *Input) KeyPress(key Key) bool {

	return i.enqueue(InputEntry{
		Type: TypeKeyDown,
		X:    int16(uint16(key) | 0x8000),
	})
}

////////////////////////////////////////////////////////////////////////////////

// KeyRelease enqueues a key-up event. Returns false if the
// queue is not connected or full.
func (i *Input) KeyRelease(key Key) bool {

	return i.enqueue(InputEntry{
		Type: TypeKeyUp,
		X:    int16(uint16(key) | 0x8000),
	})
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//                                  Mouse                                     //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// MouseClick enqueues a mouse button press followed by a
// release. Returns false if either enqueue fails.
func (i *Input) MouseClick(button Button) bool {

	if !i.MousePress(button) {
		return false
	}

	return i.MouseRelease(button)
}

////////////////////////////////////////////////////////////////////////////////

// MousePress enqueues a mouse button press. Returns false
// if the queue is not connected or full.
func (i *Input) MousePress(button Button) bool {

	return i.enqueue(InputEntry{
		Type:  TypeMouseButtonDown,
		Flags: 1 << button,
	})
}

////////////////////////////////////////////////////////////////////////////////

// MouseRelease enqueues a mouse button release. Returns
// false if the queue is not connected or full.
func (i *Input) MouseRelease(button Button) bool {

	return i.enqueue(InputEntry{
		Type:  TypeMouseButtonUp,
		Flags: 1 << button,
	})
}

////////////////////////////////////////////////////////////////////////////////

// MouseScrollH enqueues a horizontal scroll event. The
// amount is pre-multiplied by 120 (one notch = 120,
// positive is right). Returns false if the queue is not
// connected or full.
func (i *Input) MouseScrollH(amount int16) bool {

	return i.enqueue(InputEntry{
		Type: TypeMouseScrollH,
		X:    amount,
	})
}

////////////////////////////////////////////////////////////////////////////////

// MouseScrollV enqueues a vertical scroll event. The
// amount is pre-multiplied by 120 (one notch = 120,
// positive is up). Returns false if the queue is not
// connected or full.
func (i *Input) MouseScrollV(amount int16) bool {

	return i.enqueue(InputEntry{
		Type: TypeMouseScrollV,
		X:    amount,
	})
}

////////////////////////////////////////////////////////////////////////////////

// MouseMove enqueues a relative mouse movement. Send many
// small deltas for smooth aiming rather than one large one.
// Returns false if the queue is not connected or full.
func (i *Input) MouseMove(dx, dy int16) bool {

	return i.enqueue(InputEntry{
		Type: TypeMouseMoveRel,
		X:    dx,
		Y:    dy,
	})
}

////////////////////////////////////////////////////////////////////////////////

// MouseMoveTo enqueues an absolute mouse movement. X and Y
// are in stream resolution coordinates (e.g. 0-1919 for
// 1920px wide). Returns false if the queue is not connected
// or full.
func (i *Input) MouseMoveTo(x, y int16) bool {

	return i.enqueue(InputEntry{
		Type: TypeMouseMoveAbs,
		X:    x,
		Y:    y,
	})
}

////////////////////////////////////////////////////////////////////////////////

func (i *Input) enqueue(entry InputEntry) bool {

	i.mu.RLock()
	defer i.mu.RUnlock()

	if i.queue == nil {
		return false
	}

	return i.queue.Enqueue(entry)
}

////////////////////////////////////////////////////////////////////////////////

func (i *Input) connectionLoop(ctx context.Context) {

	var lastCheck time.Time

	for {
		if ctx.Err() != nil {
			return
		}

		//--------------------------------------------------------------------//

		i.mu.RLock()
		connected := i.queue != nil
		i.mu.RUnlock()

		// Try to attach if not connected
		if !connected {
			queue, err := ShmOpen()
			if err != nil {
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Second):
				}
				continue
			}

			i.mu.Lock()
			i.queue = queue
			i.mu.Unlock()

			lastCheck = time.Now()
			logger.Dbg("input queue attached")
		}

		//--------------------------------------------------------------------//

		// Check if the segment has been unlinked
		if time.Since(lastCheck) > shmCheckInterval {
			lastCheck = time.Now()

			i.mu.RLock()
			unlinked := i.queue != nil && i.queue.IsUnlinked()
			i.mu.RUnlock()

			if unlinked {
				logger.Dbg("input queue unlinked, waiting for reconnect")

				i.mu.Lock()
				i.queue.Close()
				i.queue = nil
				i.mu.Unlock()

				continue
			}
		}

		//--------------------------------------------------------------------//

		// Sleep until the next check
		select {
		case <-ctx.Done():
			return
		case <-time.After(shmCheckInterval):
		}
	}
}
