package leech

import (
	"fmt"
)

////////////////////////////////////////////////////////////////////////////////

type Region struct {
	Valid bool
	Bound bool

	Start uintptr
	Stop  uintptr
	Size  uintptr

	Readable   bool
	Writable   bool
	Executable bool
	Access     uint32

	Private bool
	Guarded bool
}

////////////////////////////////////////////////////////////////////////////////

func (r *Region) String() string {

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
		"%d %d  %08X  %08X  %08X  %s [%s][%s]",
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

func (r *Region) Contains(address uintptr) bool {

	return r.Start <= address && address < r.Stop
}

////////////////////////////////////////////////////////////////////////////////

func (r *Region) Compare(value *Region) int {

	if r.Start < value.Start {
		return -1
	}

	if r.Start > value.Start {
		return 1
	}

	return 0
}

////////////////////////////////////////////////////////////////////////////////

func (r *Region) Lt(value *Region) bool {
	return r.Start < value.Start
}

////////////////////////////////////////////////////////////////////////////////

func (r *Region) Gt(value *Region) bool {
	return r.Start > value.Start
}

////////////////////////////////////////////////////////////////////////////////

func (r *Region) Le(value *Region) bool {
	return r.Start <= value.Start
}

////////////////////////////////////////////////////////////////////////////////

func (r *Region) Ge(value *Region) bool {
	return r.Start >= value.Start
}

////////////////////////////////////////////////////////////////////////////////

func (r *Region) LtAddress(address uintptr) bool {
	return r.Start < address
}

////////////////////////////////////////////////////////////////////////////////

func (r *Region) GtAddress(address uintptr) bool {
	return r.Start > address
}

////////////////////////////////////////////////////////////////////////////////

func (r *Region) LeAddress(address uintptr) bool {
	return r.Start <= address
}

////////////////////////////////////////////////////////////////////////////////

func (r *Region) GeAddress(address uintptr) bool {
	return r.Start >= address
}

////////////////////////////////////////////////////////////////////////////////

func (r *Region) Eq(value *Region) bool {

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

func (r *Region) Ne(value *Region) bool {

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
