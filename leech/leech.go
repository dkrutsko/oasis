package leech

import (
	"regexp"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/dkrutsko/oasis/errors"
	"github.com/dkrutsko/oasis/logger"
)

////////////////////////////////////////////////////////////////////////////////

type Leech struct {
	args   []string
	handle uintptr
	lock   sync.RWMutex
}

////////////////////////////////////////////////////////////////////////////////

func New(args ...string) *Leech {

	l := &Leech{
		args:   args,
		handle: 0,
	}

	return l
}

////////////////////////////////////////////////////////////////////////////////

func (l *Leech) Create() error {

	//----------------------------------------------------------------------------//

	// Load VMM library
	err := loadVmmDll()
	if err != nil {
		return err
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	// Close existing
	if l.handle != 0 {
		err := l.Close()
		if err != nil {
			return err
		}
	}

	//----------------------------------------------------------------------------//

	// Count number of arguments
	argc := uintptr(len(l.args))

	// Convert into C-style string array
	argv := make([]uintptr, len(l.args))
	for i, arg := range l.args {

		// Try and convert string to C-style bytes
		cStr, err := syscall.BytePtrFromString(arg)
		if err != nil {
			return err
		}

		// Store the C-style bytes in arguments
		argv[i] = uintptr(unsafe.Pointer(cStr))
	}

	// Attempt to perform VMM initialization
	handle, _, err := vmmDll.initializeEx.Call(
		argc, uintptr(unsafe.Pointer(&argv[0])), 0,
	)
	if err.(windows.Errno) != 0 {
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
	_, _, err := vmmDll.close.Call(l.handle)
	if err.(windows.Errno) != 0 {
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

func (l *Leech) IsValid() bool {

	l.lock.RLock()
	defer l.lock.RUnlock()

	return l.handle != 0
}

////////////////////////////////////////////////////////////////////////////////

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
	success, _, err := vmmDll.processGetInformation.Call(
		l.handle,
		uintptr(pid),
		uintptr(unsafe.Pointer(&process)),
		uintptr(unsafe.Pointer(&size)),
	)
	if err.(windows.Errno) != 0 {
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
	success, _, err := vmmDll.processGetInformationAll.Call(
		l.handle,
		uintptr(unsafe.Pointer(&items)),
		uintptr(unsafe.Pointer(&count)),
	)
	if err.(windows.Errno) != 0 {
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
		_, _, _ = vmmDll.memFree.Call(items)
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
