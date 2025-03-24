package leech

import (
	"regexp"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/dkrutsko/oasis/errors"
)

////////////////////////////////////////////////////////////////////////////////

type Process struct {
	leech *Leech

	pid     uint32
	name    string
	is64Bit bool

	peb   uint64
	peb32 uint32

	exited bool
}

////////////////////////////////////////////////////////////////////////////////

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

func (p *Process) GetPid() uint32 {
	return p.pid
}

////////////////////////////////////////////////////////////////////////////////

func (p *Process) GetName() string {
	return p.name
}

////////////////////////////////////////////////////////////////////////////////

func (p *Process) Is64Bit() bool {
	return p.is64Bit
}

////////////////////////////////////////////////////////////////////////////////

func (p *Process) GetPeb() uint64 {
	return p.peb
}

////////////////////////////////////////////////////////////////////////////////

func (p *Process) GetPeb32() uint32 {
	return p.peb32
}

////////////////////////////////////////////////////////////////////////////////

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

func (p *Process) GetModules(filter *regexp.Regexp) ([]*Module, error) {

	//----------------------------------------------------------------------------//

	var result []*Module

	p.leech.lock.RLock()
	defer p.leech.lock.RUnlock()

	// Make sure that there is a valid leech handle and PID
	if p.leech == nil || p.leech.handle == 0 || p.pid == 0 {
		return nil, errors.New("process is not valid")
	}

	//----------------------------------------------------------------------------//

	var mapPtr uintptr
	// Attempt to retrieve the module information
	success, _, err := vmmDll.mapGetModuleW.Call(
		p.leech.handle,
		uintptr(p.pid),
		uintptr(unsafe.Pointer(&mapPtr)),
		0,
	)
	if err.(windows.Errno) != 0 {
		return nil, errors.New(
			"failed to get modules",
			errors.Error("error", err),
		)
	}
	if success == 0 {
		return nil, errors.New(
			"failed to get modules",
		)
	}

	defer func() {
		// Try and free the allocated memory
		_, _, _ = vmmDll.memFree.Call(mapPtr)
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

			// Convert the current pointer to module structure
			modulePtr := unsafe.Pointer(base + uintptr(i)*size)
			module := (*vmmMapModuleEntry)(modulePtr)

			// Ensure that the module is loaded normally
			if module.moduleType == vmmModuleTypeNormal {

				// Convert names from wide string
				name := wideToString(module.name)
				path := wideToString(module.path)

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

func (p *Process) GetMemory() *Memory {

	// TODO: Create memory cache

	m := &Memory{
		leech: p.leech,
		proc:  p,
	}

	return m
}

////////////////////////////////////////////////////////////////////////////////

func (p *Process) Eq(value *Process) bool {
	return p.pid == value.pid
}

////////////////////////////////////////////////////////////////////////////////

func (p *Process) Ne(value *Process) bool {
	return p.pid != value.pid
}

////////////////////////////////////////////////////////////////////////////////

func (p *Process) EqPid(pid uint32) bool {
	return p.pid == pid
}

////////////////////////////////////////////////////////////////////////////////

func (p *Process) NePid(pid uint32) bool {
	return p.pid != pid
}
