package leech

import (
	"regexp"
	"sync"
	"syscall"
	"unsafe"

	"github.com/dkrutsko/oasis/errors"
	"github.com/dkrutsko/oasis/logger"
)

////////////////////////////////////////////////////////////////////////////////

// Options configures a Leech connection. Pass the configured
// options to `New` to create a new instance.
type Options struct {

	// Args are the raw VMMDLL initialization arguments
	// passed to the underlying library. For example:
	// ["-device", "fpga"].
	Args []string
}

////////////////////////////////////////////////////////////////////////////////

// Leech manages a VMMDLL handle for reading and writing
// memory on a remote Windows system via DMA. Use `New` to
// create an instance and `Create` to establish the connection.
// Use `Close` to release the handle when done. Leech is safe
// for concurrent use by multiple goroutines.
type Leech struct {
	opts   *Options
	handle uintptr
	lock   sync.RWMutex
}

////////////////////////////////////////////////////////////////////////////////

// New returns a new Leech configured with the given options.
// Call Create to establish the connection.
func New(opts *Options) *Leech {

	return &Leech{opts: opts}
}

////////////////////////////////////////////////////////////////////////////////

// Create loads the VMM library and initializes the DMA
// connection using the arguments provided to New. If a
// connection is already open, it is closed first.
func (l *Leech) Create() error {

	//----------------------------------------------------------------------------//

	// Load VMM library
	err := loadVmmDll()
	if err != nil {
		return err
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	// Close existing handle without calling Close
	// to avoid deadlock since we already hold the lock
	if l.handle != 0 {
		_, err := vmmCall(vmmDll.close, l.handle)
		if err != nil {
			return errors.New(
				"failed to close vmm handle",
				errors.Error("error", err),
			)
		}
		l.handle = 0
	}

	//----------------------------------------------------------------------------//

	// Count number of arguments
	argc := uintptr(len(l.opts.Args))

	// Convert into C-style string array
	argv := make([]uintptr, len(l.opts.Args))
	for i, arg := range l.opts.Args {

		// Try and convert string to C-style bytes
		cStr, err := syscall.BytePtrFromString(arg)
		if err != nil {
			return errors.New(
				"failed to convert argument",
				errors.String("arg", arg),
				errors.Error("error", err),
			)
		}

		// Store the C-style bytes in arguments
		argv[i] = uintptr(unsafe.Pointer(cStr))
	}

	// Attempt to perform VMM initialization
	handle, err := vmmCall(
		vmmDll.initializeEx,
		argc, uintptr(unsafe.Pointer(&argv[0])), 0,
	)
	if err != nil {
		return errors.New(
			"failed to initialize vmm",
			errors.Error("error", err),
		)
	}
	if handle == 0 {
		return errors.New(
			"failed to initialize vmm",
		)
	}

	//----------------------------------------------------------------------------//

	l.handle = handle
	logger.Dbg("vmm was initialized")

	return nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// Close releases the VMMDLL handle and disconnects from the
// DMA device. Does nothing if no connection is open. Safe to
// call multiple times.
func (l *Leech) Close() error {

	//----------------------------------------------------------------------------//

	l.lock.Lock()
	defer l.lock.Unlock()

	// If handle valid
	if l.handle == 0 {
		return nil
	}

	//----------------------------------------------------------------------------//

	// Try and close the initialized handle
	_, err := vmmCall(vmmDll.close, l.handle)
	if err != nil {
		return errors.New(
			"failed to close vmm handle",
			errors.Error("error", err),
		)
	}

	//----------------------------------------------------------------------------//

	l.handle = 0
	logger.Dbg("vmm handle was closed")

	return nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// IsValid returns true if the VMMDLL handle is initialized.
func (l *Leech) IsValid() bool {

	l.lock.RLock()
	defer l.lock.RUnlock()

	return l.handle != 0
}

////////////////////////////////////////////////////////////////////////////////

// GetConfig retrieves a VMMDLL configuration value. Use the
// Config* constants to specify which option to query. Returns
// an error if the handle is not initialized or the query
// fails.
func (l *Leech) GetConfig(option uint64) (uint64, error) {

	//----------------------------------------------------------------------------//

	l.lock.RLock()
	defer l.lock.RUnlock()

	if l.handle == 0 {
		return 0, errors.New("leech not initialized")
	}

	//----------------------------------------------------------------------------//

	var value uint64

	success, err := vmmCall(
		vmmDll.configGet,
		l.handle,
		uintptr(option),
		uintptr(unsafe.Pointer(&value)),
	)
	if err != nil {
		return 0, errors.New(
			"failed to get config",
			errors.Error("error", err),
		)
	}
	if success == 0 {
		return 0, errors.New(
			"failed to get config",
		)
	}

	//----------------------------------------------------------------------------//

	return value, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// SetConfig sets a VMMDLL configuration value. Use the
// Config* constants to specify which option to set. Returns
// an error if the handle is not initialized or the operation
// fails.
func (l *Leech) SetConfig(option uint64, value uint64) error {

	//----------------------------------------------------------------------------//

	l.lock.RLock()
	defer l.lock.RUnlock()

	if l.handle == 0 {
		return errors.New("leech not initialized")
	}

	//----------------------------------------------------------------------------//

	success, err := vmmCall(
		vmmDll.configSet,
		l.handle,
		uintptr(option),
		uintptr(value),
	)
	if err != nil {
		return errors.New(
			"failed to set config",
			errors.Error("error", err),
		)
	}
	if success == 0 {
		return errors.New(
			"failed to set config",
		)
	}

	//----------------------------------------------------------------------------//

	return nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// GetProcess retrieves information about a single process
// identified by `pid`. Returns an error if the handle is not
// initialized, the PID is not found, or the VMMDLL response
// fails validation.
func (l *Leech) GetProcess(pid uint32) (*Process, error) {

	//----------------------------------------------------------------------------//

	var result *Process = nil

	l.lock.RLock()
	defer l.lock.RUnlock()

	// If handle valid
	if l.handle == 0 {
		return nil, errors.New("leech not initialized")
	}

	//----------------------------------------------------------------------------//

	// Output needs magic and version
	process := vmmProcessInformation{
		magic:   vmmProcessInformationMagic,
		version: vmmProcessInformationVersion,
	}

	// Calculate the size of the output structure
	size := unsafe.Sizeof(vmmProcessInformation{})

	// Attempt to retrieve information about the process
	success, err := vmmCall(
		vmmDll.processGetInformation,
		l.handle,
		uintptr(pid),
		uintptr(unsafe.Pointer(&process)),
		uintptr(unsafe.Pointer(&size)),
	)
	if err != nil {
		return nil, errors.New(
			"failed to get process info",
			errors.Uint32("pid", pid),
			errors.Error("error", err),
		)
	}
	if success == 0 {
		return nil, errors.New(
			"failed to get process info",
			errors.Uint32("pid", pid),
		)
	}

	//----------------------------------------------------------------------------//

	// Ensure that the magic number matches constant
	if process.magic != vmmProcessInformationMagic {
		return nil, errors.New(
			"failed to verify process magic number",
		)
	}

	// Ensure that the version value matches constant
	if process.version != vmmProcessInformationVersion {
		return nil, errors.New(
			"failed to verify process version value",
		)
	}

	// Check if structure size matches
	if process.size != uint16(size) {
		return nil, errors.New(
			"process structure has a size mismatch",
		)
	}

	// Perform sanity check
	if process.pid != pid {
		return nil, errors.New(
			"failed to query for requested process",
		)
	}

	// Convert the name from a C-style string
	name := cStrToString(process.nameLong[:])

	is64Bit := false
	// If 32-bit app is running on 64-bit system
	if process.system == vmmSystemTypeWindows64 {
		is64Bit = process.win.wow64 == 0
	}

	result = &Process{
		leech:   l,
		pid:     process.pid,
		name:    name,
		is64Bit: is64Bit,
		peb:     process.win.peb,
		peb32:   process.win.peb32,
		exited:  process.state != 0,
	}

	//----------------------------------------------------------------------------//

	return result, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// GetProcessList retrieves information about all processes on
// the target system. `filter` accepts a regular expression
// which is matched against process names. Pass nil to include
// all processes. If `onlyActive` is true, exited processes are
// excluded. Returns an error if the handle is not initialized.
func (l *Leech) GetProcessList(filter *regexp.Regexp, onlyActive bool) ([]*Process, error) {

	//----------------------------------------------------------------------------//

	var result []*Process

	l.lock.RLock()
	defer l.lock.RUnlock()

	// If handle valid
	if l.handle == 0 {
		return nil, errors.New("leech not initialized")
	}

	//----------------------------------------------------------------------------//

	var items uintptr
	var count uint32

	// Attempt to retrieve information about all processes
	success, err := vmmCall(
		vmmDll.processGetInformationAll,
		l.handle,
		uintptr(unsafe.Pointer(&items)),
		uintptr(unsafe.Pointer(&count)),
	)
	if err != nil {
		return nil, errors.New(
			"failed to list process info",
			errors.Error("error", err),
		)
	}
	if success == 0 {
		return nil, errors.New(
			"failed to list process info",
		)
	}

	defer func() {
		// Try and free the allocated memory
		_, _ = vmmCall(vmmDll.memFree, items)
	}()

	//----------------------------------------------------------------------------//

	size := unsafe.Sizeof(vmmProcessInformation{})

	// Iterate through number of items
	for i := uint32(0); i < count; i++ {

		// Convert the current pointer into process structure
		processPtr := unsafe.Pointer(items + uintptr(i)*size)
		process := (*vmmProcessInformation)(processPtr)

		// Ensure that the magic number matches constant
		if process.magic != vmmProcessInformationMagic {
			return nil, errors.New(
				"failed to verify process magic number",
			)
		}

		// Ensure that the version value matches constant
		if process.version != vmmProcessInformationVersion {
			return nil, errors.New(
				"failed to verify process version value",
			)
		}

		// Check if structure size matches
		if process.size != uint16(size) {
			return nil, errors.New(
				"process structure has a size mismatch",
			)
		}

		// Skip processes that aren't running
		if onlyActive && process.state != 0 {
			continue
		}

		// Convert the name from a C-style string
		name := cStrToString(process.nameLong[:])

		is64Bit := false
		// If 32-bit app is running on 64-bit system
		if process.system == vmmSystemTypeWindows64 {
			is64Bit = process.win.wow64 == 0
		}

		// Check the name if name filter was specified
		if filter == nil || filter.MatchString(name) {

			p := &Process{
				leech:   l,
				pid:     process.pid,
				name:    name,
				is64Bit: is64Bit,
				peb:     process.win.peb,
				peb32:   process.win.peb32,
				exited:  process.state != 0,
			}

			result = append(result, p)
		}
	}

	//----------------------------------------------------------------------------//

	return result, nil

	//----------------------------------------------------------------------------//
}
