package leech

import (
	"context"
	"regexp"
	"unsafe"

	"github.com/dkrutsko/oasis/errors"
)

////////////////////////////////////////////////////////////////////////////////

// Process represents a single Windows process on the DMA
// target. Processes are retrieved through `Leech.GetProcess`
// or `Leech.GetProcessList`. All information except the path
// is cached at creation time.
type Process struct {
	leech *Leech

	pid     uint32
	name    string
	is64Bit bool

	peb   uint64
	peb32 uint32

	exited bool
	path   string
}

////////////////////////////////////////////////////////////////////////////////

// IsValid returns true if the process still exists and is
// accessible on the target system. This performs a live
// query to VMMDLL each time it is called.
func (p *Process) IsValid() bool {

	// Try and retrieve the process using PID
	process, err := p.leech.GetProcess(p.pid)
	if err != nil {
		return false
	}

	// Process still valid
	return process != nil
}

////////////////////////////////////////////////////////////////////////////////

// Is64Bit returns true if the process is a native 64-bit
// process. Returns false for WOW64 (32-bit on 64-bit)
// processes.
func (p *Process) Is64Bit() bool {
	return p.is64Bit
}

////////////////////////////////////////////////////////////////////////////////

// IsDebugged returns true if the process has an active
// debugger attached. This is determined by reading the
// `BeingDebugged` byte from the PEB via DMA. Returns
// false with an error if the read fails.
func (p *Process) IsDebugged() (bool, error) {

	//----------------------------------------------------------------------------//

	// Lock for retrieval
	p.leech.lock.RLock()
	defer p.leech.lock.RUnlock()

	// Make sure handle and PID are valid
	if p.leech.handle == 0 || p.pid == 0 {
		return false, errors.New("process is not valid")
	}

	//----------------------------------------------------------------------------//

	// Determine the PEB address based on process bitness.
	// `BeingDebugged` is at offset 0x2 in both 32-bit and
	// 64-bit PEB structures.
	var pebAddr uintptr
	if p.is64Bit {
		pebAddr = uintptr(p.peb) + 2
	} else {
		pebAddr = uintptr(p.peb32) + 2
	}

	//----------------------------------------------------------------------------//

	var bytesRead uint32
	var result [1]byte

	// Read the `BeingDebugged` byte from the PEB
	success, err := vmmCall(
		vmmDll.memReadEx,
		p.leech.handle,
		uintptr(p.pid),
		pebAddr,
		uintptr(unsafe.Pointer(&result[0])),
		1,
		uintptr(unsafe.Pointer(&bytesRead)),
		uintptr(0x1), // VMMDLL_FLAG_NOCACHE
	)
	if err != nil {
		return false, errors.New(
			"failed to read peb",
			errors.Uint32("pid", p.pid),
			errors.Error("error", err),
		)
	}
	if success == 0 {
		return false, errors.New(
			"failed to read peb",
			errors.Uint32("pid", p.pid),
		)
	}

	//----------------------------------------------------------------------------//

	return result[0] != 0, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// GetPid returns the unique PID of the process. Returns zero
// if no process was selected.
func (p *Process) GetPid() uint32 {
	return p.pid
}

////////////////////////////////////////////////////////////////////////////////

// GetName returns the executable image name of the process
// as a UTF-8 encoded string.
func (p *Process) GetName() string {
	return p.name
}

////////////////////////////////////////////////////////////////////////////////

// GetPath returns the full path of the process executable
// as a UTF-8 encoded string. The path is lazily loaded from
// VMMDLL on first access and cached for subsequent calls.
// Returns an empty string if no path is available.
func (p *Process) GetPath() (string, error) {

	//----------------------------------------------------------------------------//

	// Return cached path if available
	if p.path != "" {
		return p.path, nil
	}

	// Lock for retrieval
	p.leech.lock.RLock()
	defer p.leech.lock.RUnlock()

	// Make sure handle and PID are valid
	if p.leech.handle == 0 || p.pid == 0 {
		return "", errors.New("process is not valid")
	}

	//----------------------------------------------------------------------------//

	var strPtr uintptr

	// Option 2 = PATH_USER_IMAGE (full exe path)
	success, err := vmmCall(
		vmmDll.processGetInformationString,
		p.leech.handle,
		uintptr(p.pid),
		uintptr(2),
		uintptr(unsafe.Pointer(&strPtr)),
	)
	if err != nil {
		return "", errors.New(
			"failed to get process path",
			errors.Uint32("pid", p.pid),
			errors.Error("error", err),
		)
	}
	if success == 0 {
		return "", errors.New(
			"failed to get process path",
			errors.Uint32("pid", p.pid),
		)
	}

	//----------------------------------------------------------------------------//

	// Convert the returned string pointer
	if strPtr != 0 {
		p.path = ptrToString(strPtr)

		// Free the allocated string memory
		_, _ = vmmCall(vmmDll.memFree, strPtr)
	}

	//----------------------------------------------------------------------------//

	return p.path, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// GetPeb returns the 64-bit PEB (Process Environment Block)
// address for the process.
func (p *Process) GetPeb() uint64 {
	return p.peb
}

////////////////////////////////////////////////////////////////////////////////

// GetPeb32 returns the 32-bit PEB address for WOW64 processes.
// Returns zero for native 64-bit processes.
func (p *Process) GetPeb32() uint32 {
	return p.peb32
}

////////////////////////////////////////////////////////////////////////////////

// HasExited returns true if the process has terminated. If
// the process cannot be queried, it is assumed to have
// exited.
func (p *Process) HasExited() bool {

	// Try and retrieve the process using PID
	process, err := p.leech.GetProcess(p.pid)
	if err != nil {
		return true
	}

	// If process valid
	if process != nil {
		return process.exited
	}

	return true
}

////////////////////////////////////////////////////////////////////////////////

// GetModules returns the list of loaded modules for this
// process. `filter` accepts a regular expression which is
// matched against module names. Pass nil to include all
// modules. Only normally loaded modules are included.
// The resulting list is sorted by base address. Returns
// an error if the process is not valid.
func (p *Process) GetModules(ctx context.Context, filter *regexp.Regexp) ([]*Module, error) {

	//----------------------------------------------------------------------------//

	var result []*Module

	// Lock for retrieval
	p.leech.lock.RLock()
	defer p.leech.lock.RUnlock()

	// Make sure handle and PID are valid
	if p.leech.handle == 0 || p.pid == 0 {
		return nil, errors.New("process is not valid")
	}

	//----------------------------------------------------------------------------//

	var mapPtr uintptr
	// Attempt to retrieve the module information
	success, err := vmmCall(
		vmmDll.mapGetModule,
		p.leech.handle,
		uintptr(p.pid),
		uintptr(unsafe.Pointer(&mapPtr)),
		0,
	)
	if err != nil {
		return nil, errors.New(
			"failed to get modules",
			errors.Uint32("pid", p.pid),
			errors.Error("error", err),
		)
	}
	if success == 0 {
		return nil, errors.New(
			"failed to get modules",
			errors.Uint32("pid", p.pid),
		)
	}

	defer func() {
		// Try and free the allocated memory
		_, _ = vmmCall(vmmDll.memFree, mapPtr)
	}()

	//----------------------------------------------------------------------------//

	{
		// Convert returned pointer into a module structure
		moduleMap := (*vmmMapModule)(unsafe.Pointer(mapPtr))

		// Ensure that version value matches constant
		if moduleMap.version != vmmMapModuleVersion {
			return nil, errors.New(
				"failed to verify module version value",
			)
		}

		base := uintptr(unsafe.Pointer(&moduleMap.modules))
		size := unsafe.Sizeof(vmmMapModuleEntry{})

		// Iterate through number of returned modules
		for i := uint32(0); i < moduleMap.count; i++ {

			if ctx.Err() != nil {
				return result, ctx.Err()
			}

			// Convert the current pointer to module structure
			modulePtr := unsafe.Pointer(base + uintptr(i)*size)
			module := (*vmmMapModuleEntry)(modulePtr)

			// Ensure that the module is loaded normally
			if module.moduleType == vmmModuleTypeNormal {

				// Convert names from native string
				name := ptrToString(module.name)
				path := ptrToString(module.path)

				// Check the name if name filter was specified
				if filter == nil || filter.MatchString(name) {

					m := &Module{
						proc:  p,
						name:  name,
						path:  path,
						base:  uintptr(module.base),
						size:  uintptr(module.size),
						entry: uintptr(module.entry),
					}

					result = append(result, m)
				}
			}
		}
	}

	//----------------------------------------------------------------------------//

	return result, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// GetMemory returns a new Memory instance for reading and
// writing the virtual address space of this process. Each
// call creates an independent instance. Use `CreateCache`
// on the returned Memory to enable caching.
func (p *Process) GetMemory() *Memory {

	return &Memory{
		leech: p.leech,
		proc:  p,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Eq returns true if both processes have the same PID.
func (p *Process) Eq(value *Process) bool {
	return p.pid == value.pid
}

////////////////////////////////////////////////////////////////////////////////

// Ne returns true if the processes have different PIDs.
func (p *Process) Ne(value *Process) bool {
	return p.pid != value.pid
}

////////////////////////////////////////////////////////////////////////////////

// EqPid returns true if the process PID matches the given PID.
func (p *Process) EqPid(pid uint32) bool {
	return p.pid == pid
}

////////////////////////////////////////////////////////////////////////////////

// NePid returns true if the process PID differs from the given
// PID.
func (p *Process) NePid(pid uint32) bool {
	return p.pid != pid
}
