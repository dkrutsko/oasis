package leech

import (
	"fmt"
)

////////////////////////////////////////////////////////////////////////////////

// Module represents a single module (DLL or executable) loaded
// within a process on the DMA target. Modules are retrieved
// through `Process.GetModules`. All information is cached and
// immutable.
type Module struct {
	proc *Process

	name string
	path string

	base  uintptr
	size  uintptr
	entry uintptr
}

////////////////////////////////////////////////////////////////////////////////

// String returns a formatted representation of the module
// showing base address, size, entry point, and path.
func (m Module) String() string {

	return fmt.Sprintf(
		"%016X  %016X  %016X  %s",
		m.base,
		m.size,
		m.entry,
		m.path,
	)
}

////////////////////////////////////////////////////////////////////////////////

// GetProcess returns the process that owns this module.
func (m Module) GetProcess() *Process {
	return m.proc
}

////////////////////////////////////////////////////////////////////////////////

// GetName returns the name of this module as a UTF-8
// encoded string.
func (m Module) GetName() string {
	return m.name
}

////////////////////////////////////////////////////////////////////////////////

// GetPath returns the full path of this module as a UTF-8
// encoded string.
func (m Module) GetPath() string {
	return m.path
}

////////////////////////////////////////////////////////////////////////////////

// GetBase returns the load address of this module within
// the process.
func (m Module) GetBase() uintptr {
	return m.base
}

////////////////////////////////////////////////////////////////////////////////

// GetSize returns the size this module occupies within
// the process, in bytes.
func (m Module) GetSize() uintptr {
	return m.size
}

////////////////////////////////////////////////////////////////////////////////

// GetEntry returns the entry point address of the module.
func (m Module) GetEntry() uintptr {
	return m.entry
}

////////////////////////////////////////////////////////////////////////////////

// Contains returns true if `address` is in the range
// [GetBase, GetBase + GetSize) of this module.
func (m Module) Contains(address uintptr) bool {

	return m.base <= address && address < (m.base+m.size)
}

////////////////////////////////////////////////////////////////////////////////

// Compare returns -1, 0, or 1 comparing modules by base
// address.
func (m Module) Compare(value *Module) int {

	if m.base < value.base {
		return -1
	}

	if m.base > value.base {
		return 1
	}

	return 0
}

////////////////////////////////////////////////////////////////////////////////

// Lt returns true if this module's base is less than the
// other module's base.
func (m Module) Lt(value *Module) bool {
	return m.base < value.base
}

////////////////////////////////////////////////////////////////////////////////

// Gt returns true if this module's base is greater than the
// other module's base.
func (m Module) Gt(value *Module) bool {
	return m.base > value.base
}

////////////////////////////////////////////////////////////////////////////////

// Le returns true if this module's base is less than or equal
// to the other module's base.
func (m Module) Le(value *Module) bool {
	return m.base <= value.base
}

////////////////////////////////////////////////////////////////////////////////

// Ge returns true if this module's base is greater than or
// equal to the other module's base.
func (m Module) Ge(value *Module) bool {
	return m.base >= value.base
}

////////////////////////////////////////////////////////////////////////////////

// LtAddress returns true if this module's base is less than
// the given address.
func (m Module) LtAddress(address uintptr) bool {
	return m.base < address
}

////////////////////////////////////////////////////////////////////////////////

// GtAddress returns true if this module's base is greater than
// the given address.
func (m Module) GtAddress(address uintptr) bool {
	return m.base > address
}

////////////////////////////////////////////////////////////////////////////////

// LeAddress returns true if this module's base is less than or
// equal to the given address.
func (m Module) LeAddress(address uintptr) bool {
	return m.base <= address
}

////////////////////////////////////////////////////////////////////////////////

// GeAddress returns true if this module's base is greater than
// or equal to the given address.
func (m Module) GeAddress(address uintptr) bool {
	return m.base >= address
}

////////////////////////////////////////////////////////////////////////////////

// Eq returns true if both modules have the same process, base,
// size, and entry point.
func (m Module) Eq(value *Module) bool {

	return m.proc == value.proc &&
		m.base == value.base &&
		m.size == value.size &&
		m.entry == value.entry
}

////////////////////////////////////////////////////////////////////////////////

// Ne returns true if the modules differ in process, base,
// size, or entry point.
func (m Module) Ne(value *Module) bool {

	return m.proc != value.proc ||
		m.base != value.base ||
		m.size != value.size ||
		m.entry != value.entry
}
