package leech

import (
	"fmt"
)

////////////////////////////////////////////////////////////////////////////////

// Region represents a contiguous virtual memory region with
// consistent protection attributes. Regions are retrieved
// through `Memory.GetRegion` or `Memory.GetRegions`. All
// properties are public and mutable.
type Region struct {

	// Valid is true if the region is backed by a VAD or PTE
	// entry. False indicates an unmapped gap.
	Valid bool

	// Bound is true if the region's pages are committed and
	// backed by physical memory or the pagefile.
	Bound bool

	// Start is the first address in the region.
	Start uintptr

	// Stop is the address immediately after the last byte in
	// the region.
	Stop uintptr

	// Size is the total size of the region in bytes.
	Size uintptr

	// Readable is true if the region has read access.
	Readable bool

	// Writable is true if the region has write access.
	Writable bool

	// Executable is true if the region has execute access.
	Executable bool

	// Access is the raw Windows page protection attribute.
	// Use the AccessPage* constants to test individual flags.
	Access uint32

	// Private is true if the region is private memory (not
	// shared with other processes).
	Private bool

	// Guarded is true if the region has the `PageGuard` flag.
	// Any attempt to access a guard page causes the system to
	// raise an exception and turn off the guard page status.
	Guarded bool
}

////////////////////////////////////////////////////////////////////////////////

// String returns a formatted representation of the region
// showing validity, bounds, addresses, permissions, and flags.
func (r Region) String() string {

	valid := 0
	if r.Valid {
		valid = 1
	}
	bound := 0
	if r.Bound {
		bound = 1
	}

	access := ""
	if r.Readable {
		access += "R"
	} else {
		access += "-"
	}
	if r.Writable {
		access += "W"
	} else {
		access += "-"
	}
	if r.Executable {
		access += "X"
	} else {
		access += "-"
	}

	private := ""
	if r.Private {
		private += "P"
	} else {
		private += " "
	}
	guarded := ""
	if r.Guarded {
		guarded += "G"
	} else {
		guarded += " "
	}

	return fmt.Sprintf(
		"%d %d  %016X  %016X  %016X  %s [%s][%s]",
		valid,
		bound,
		r.Start,
		r.Stop,
		r.Size,
		access,
		private,
		guarded,
	)
}

////////////////////////////////////////////////////////////////////////////////

// Contains returns true if `address` is in the range
// [Start, Stop) of this region.
func (r Region) Contains(address uintptr) bool {

	return r.Start <= address && address < r.Stop
}

////////////////////////////////////////////////////////////////////////////////

// Compare returns -1, 0, or 1 comparing regions by start
// address.
func (r Region) Compare(value *Region) int {

	if r.Start < value.Start {
		return -1
	}

	if r.Start > value.Start {
		return 1
	}

	return 0
}

////////////////////////////////////////////////////////////////////////////////

// Lt returns true if this region's start is less than the
// other region's start.
func (r Region) Lt(value *Region) bool {
	return r.Start < value.Start
}

////////////////////////////////////////////////////////////////////////////////

// Gt returns true if this region's start is greater than the
// other region's start.
func (r Region) Gt(value *Region) bool {
	return r.Start > value.Start
}

////////////////////////////////////////////////////////////////////////////////

// Le returns true if this region's start is less than or equal
// to the other region's start.
func (r Region) Le(value *Region) bool {
	return r.Start <= value.Start
}

////////////////////////////////////////////////////////////////////////////////

// Ge returns true if this region's start is greater than or
// equal to the other region's start.
func (r Region) Ge(value *Region) bool {
	return r.Start >= value.Start
}

////////////////////////////////////////////////////////////////////////////////

// LtAddress returns true if this region's start is less than
// the given address.
func (r Region) LtAddress(address uintptr) bool {
	return r.Start < address
}

////////////////////////////////////////////////////////////////////////////////

// GtAddress returns true if this region's start is greater
// than the given address.
func (r Region) GtAddress(address uintptr) bool {
	return r.Start > address
}

////////////////////////////////////////////////////////////////////////////////

// LeAddress returns true if this region's start is less than
// or equal to the given address.
func (r Region) LeAddress(address uintptr) bool {
	return r.Start <= address
}

////////////////////////////////////////////////////////////////////////////////

// GeAddress returns true if this region's start is greater
// than or equal to the given address.
func (r Region) GeAddress(address uintptr) bool {
	return r.Start >= address
}

////////////////////////////////////////////////////////////////////////////////

// Eq returns true if both regions have identical fields.
func (r Region) Eq(value *Region) bool {

	return r.Valid == value.Valid &&
		r.Bound == value.Bound &&
		r.Start == value.Start &&
		r.Stop == value.Stop &&
		r.Size == value.Size &&
		r.Readable == value.Readable &&
		r.Writable == value.Writable &&
		r.Executable == value.Executable &&
		r.Access == value.Access &&
		r.Private == value.Private &&
		r.Guarded == value.Guarded
}

////////////////////////////////////////////////////////////////////////////////

// Ne returns true if the regions differ in any field.
func (r Region) Ne(value *Region) bool {

	return r.Valid != value.Valid ||
		r.Bound != value.Bound ||
		r.Start != value.Start ||
		r.Stop != value.Stop ||
		r.Size != value.Size ||
		r.Readable != value.Readable ||
		r.Writable != value.Writable ||
		r.Executable != value.Executable ||
		r.Access != value.Access ||
		r.Private != value.Private ||
		r.Guarded != value.Guarded
}
