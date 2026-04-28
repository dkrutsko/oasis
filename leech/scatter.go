package leech

import (
	"encoding/binary"
	"unsafe"

	"github.com/dkrutsko/oasis/errors"
)

////////////////////////////////////////////////////////////////////////////////

const (
	// ScatterFlagNoCache disables the VMMDLL internal read
	// cache for scatter operations.
	ScatterFlagNoCache uint32 = 0x0001

	// ScatterFlagNoCachePut prevents scatter results from
	// being inserted into the VMMDLL internal read cache.
	ScatterFlagNoCachePut uint32 = 0x0100

	// ScatterFlagZeroPadOnFail zero-fills requested regions
	// that fail to read instead of returning an error.
	ScatterFlagZeroPadOnFail uint32 = 0x0002

	// ScatterFlagDefault combines NoCache, NoCachePut, and
	// ZeroPadOnFail for typical DMA scatter operations.
	ScatterFlagDefault uint32 = 0x0103
)

////////////////////////////////////////////////////////////////////////////////

// Scatter provides batched read and write access to the
// virtual address space of a process via DMA. Multiple
// addresses are queued with `Prepare` and executed in a
// single DMA transaction with `Execute` or `ExecuteRead`.
// Results are retrieved with `Read` or the typed helpers.
// Use `Clear` to reset the handle between passes without
// closing it. Scatter is NOT safe for concurrent use.
type Scatter struct {
	leech  *Leech
	proc   *Process
	handle uintptr
}

////////////////////////////////////////////////////////////////////////////////

// Prepare queues a read of `size` bytes at the given virtual
// address. The read is not performed until `Execute` or
// `ExecuteRead` is called. Use `Read` to retrieve results
// afterward.
func (s *Scatter) Prepare(address uintptr, size uint32) error {

	//----------------------------------------------------------------------------//

	s.leech.lock.RLock()
	defer s.leech.lock.RUnlock()

	if s.handle == 0 {
		return errors.New("scatter handle is not valid")
	}

	//----------------------------------------------------------------------------//

	success, err := vmmCall(
		vmmDll.scatterPrepare,
		s.handle,
		address,
		uintptr(size),
	)
	if err != nil {
		return errors.New(
			"failed to prepare scatter read",
			errors.Error("error", err),
		)
	}
	if success == 0 {
		return errors.New(
			"failed to prepare scatter read",
		)
	}

	//----------------------------------------------------------------------------//

	return nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// PrepareEx queues a read at the given virtual address with
// a pre-allocated output buffer. On `Execute`, data is
// written directly into `buf` without requiring a separate
// `Read` call.
func (s *Scatter) PrepareEx(address uintptr, buf []byte) error {

	//----------------------------------------------------------------------------//

	s.leech.lock.RLock()
	defer s.leech.lock.RUnlock()

	if s.handle == 0 {
		return errors.New("scatter handle is not valid")
	}

	if len(buf) == 0 {
		return errors.New("buffer must not be empty")
	}

	//----------------------------------------------------------------------------//

	var cbRead uint32

	success, err := vmmCall(
		vmmDll.scatterPrepareEx,
		s.handle,
		address,
		uintptr(len(buf)),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&cbRead)),
	)
	if err != nil {
		return errors.New(
			"failed to prepare scatter read",
			errors.Error("error", err),
		)
	}
	if success == 0 {
		return errors.New(
			"failed to prepare scatter read",
		)
	}

	//----------------------------------------------------------------------------//

	return nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// PrepareWrite queues a write of `data` to the given virtual
// address. The data is captured at call time. The write is
// not performed until `Execute` is called.
func (s *Scatter) PrepareWrite(address uintptr, data []byte) error {

	//----------------------------------------------------------------------------//

	s.leech.lock.RLock()
	defer s.leech.lock.RUnlock()

	if s.handle == 0 {
		return errors.New("scatter handle is not valid")
	}

	if len(data) == 0 {
		return errors.New("data must not be empty")
	}

	//----------------------------------------------------------------------------//

	success, err := vmmCall(
		vmmDll.scatterPrepareWrite,
		s.handle,
		address,
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(len(data)),
	)
	if err != nil {
		return errors.New(
			"failed to prepare scatter write",
			errors.Error("error", err),
		)
	}
	if success == 0 {
		return errors.New(
			"failed to prepare scatter write",
		)
	}

	//----------------------------------------------------------------------------//

	return nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// Execute performs all queued reads and writes in a single
// DMA transaction. After this call, read results can be
// retrieved with `Read` or the typed convenience methods.
func (s *Scatter) Execute() error {

	//----------------------------------------------------------------------------//

	s.leech.lock.RLock()
	defer s.leech.lock.RUnlock()

	if s.handle == 0 {
		return errors.New("scatter handle is not valid")
	}

	//----------------------------------------------------------------------------//

	success, err := vmmCall(
		vmmDll.scatterExecute,
		s.handle,
	)
	if err != nil {
		return errors.New(
			"failed to execute scatter",
			errors.Error("error", err),
		)
	}
	if success == 0 {
		return errors.New(
			"failed to execute scatter",
		)
	}

	//----------------------------------------------------------------------------//

	return nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// ExecuteRead performs all queued reads in a single DMA
// transaction, skipping any pending writes. After this call,
// read results can be retrieved with `Read` or the typed
// convenience methods.
func (s *Scatter) ExecuteRead() error {

	//----------------------------------------------------------------------------//

	s.leech.lock.RLock()
	defer s.leech.lock.RUnlock()

	if s.handle == 0 {
		return errors.New("scatter handle is not valid")
	}

	//----------------------------------------------------------------------------//

	success, err := vmmCall(
		vmmDll.scatterExecuteRead,
		s.handle,
	)
	if err != nil {
		return errors.New(
			"failed to execute scatter read",
			errors.Error("error", err),
		)
	}
	if success == 0 {
		return errors.New(
			"failed to execute scatter read",
		)
	}

	//----------------------------------------------------------------------------//

	return nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// Read retrieves data for a previously prepared and executed
// address. Returns a byte slice of the requested size. Must
// be called after `Execute` or `ExecuteRead`.
func (s *Scatter) Read(address uintptr, size uint32) ([]byte, error) {

	//----------------------------------------------------------------------------//

	s.leech.lock.RLock()
	defer s.leech.lock.RUnlock()

	if s.handle == 0 {
		return nil, errors.New("scatter handle is not valid")
	}

	//----------------------------------------------------------------------------//

	result := make([]byte, size)
	var cbRead uint32

	success, err := vmmCall(
		vmmDll.scatterRead,
		s.handle,
		address,
		uintptr(size),
		uintptr(unsafe.Pointer(&result[0])),
		uintptr(unsafe.Pointer(&cbRead)),
	)
	if err != nil {
		return nil, errors.New(
			"failed to read scatter result",
			errors.Error("error", err),
		)
	}
	if success == 0 {
		return nil, errors.New(
			"failed to read scatter result",
		)
	}

	//----------------------------------------------------------------------------//

	return result, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// ReadPtr retrieves a pointer-sized value (8 bytes) from a
// previously prepared and executed address.
func (s *Scatter) ReadPtr(address uintptr) (uintptr, error) {

	data, err := s.Read(address, 8)
	if err != nil {
		return 0, err
	}

	return uintptr(binary.LittleEndian.Uint64(data)), nil
}

////////////////////////////////////////////////////////////////////////////////

// ReadInt32 retrieves a signed 32-bit integer from a
// previously prepared and executed address.
func (s *Scatter) ReadInt32(address uintptr) (int32, error) {

	data, err := s.Read(address, 4)
	if err != nil {
		return 0, err
	}

	return int32(binary.LittleEndian.Uint32(data)), nil
}

////////////////////////////////////////////////////////////////////////////////

// Clear resets the scatter handle for reuse. All prepared
// reads, writes, and cached results are discarded. The
// handle itself is preserved. Optionally updates the target
// PID and flags for the next batch.
func (s *Scatter) Clear(pid uint32, flags uint32) error {

	//----------------------------------------------------------------------------//

	s.leech.lock.RLock()
	defer s.leech.lock.RUnlock()

	if s.handle == 0 {
		return errors.New("scatter handle is not valid")
	}

	//----------------------------------------------------------------------------//

	success, err := vmmCall(
		vmmDll.scatterClear,
		s.handle,
		uintptr(pid),
		uintptr(flags),
	)
	if err != nil {
		return errors.New(
			"failed to clear scatter handle",
			errors.Error("error", err),
		)
	}
	if success == 0 {
		return errors.New(
			"failed to clear scatter handle",
		)
	}

	//----------------------------------------------------------------------------//

	return nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// Close releases the scatter handle. Safe to call multiple
// times. Does nothing if the handle is already closed.
func (s *Scatter) Close() {

	if s.handle != 0 {
		vmmCall(vmmDll.scatterCloseHandle, s.handle)
		s.handle = 0
	}
}
