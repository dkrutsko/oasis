package leech

import (
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/dkrutsko/oasis/errors"
)

////////////////////////////////////////////////////////////////////////////////

type Memory struct {
	leech *Leech
	proc  *Process

	//cache []byte
	//pages PageMap
	//next  uintptr

	//blockLength uintptr
	//blockBuffer uintptr

	//cacheSize   uintptr
	//enlargeSize uintptr
	//maximumSize uintptr
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) GetProcess() *Process {
	return m.proc
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) GetRegions(start uintptr, stop uintptr) ([]*Region, error) {

	//----------------------------------------------------------------------------//

	var result []*Region

	m.leech.lock.RLock()
	defer m.leech.lock.RUnlock()

	// Make sure that there is a valid leech handle and process ID
	if m.leech == nil || m.leech.handle == 0 || m.proc.pid == 0 {
		return nil, errors.New("process is not valid")
	}

	//----------------------------------------------------------------------------//

	// Ensure start is in range
	if start < m.GetMinAddress() {
		start = m.GetMinAddress()
	}

	// Ensure stop is in range
	if stop > m.GetMaxAddress() {
		stop = m.GetMaxAddress()
	}

	oldStop := stop
	// Align range to the page size
	start &= ^(m.GetPageSize() - 1)
	stop &= ^(m.GetPageSize() - 1)
	if oldStop != stop {
		stop += m.GetPageSize()
	}

	// Make sure the range is correct
	if start >= stop {
		return result, nil
	}

	//----------------------------------------------------------------------------//

	var mapVadPtr uintptr
	// Attempt to retrieve the VAD information
	success, _, err := vmmDll.mapGetVadW.Call(
		m.leech.handle,
		uintptr(m.proc.pid),
		0,
		uintptr(unsafe.Pointer(&mapVadPtr)),
	)
	if err.(windows.Errno) != 0 {
		return nil, errors.New(
			"failed to get vad info",
			errors.Error("error", err),
		)
	}
	if success == 0 {
		return nil, errors.New(
			"failed to get vad info",
		)
	}

	defer func() {
		// Attempt to free the allocated memory
		_, _, _ = vmmDll.memFree.Call(mapVadPtr)
	}()

	//----------------------------------------------------------------------------//

	{
		// Convert returned pointer into VAD structure
		vadMap := (*vmmMapVad)(unsafe.Pointer(mapVadPtr))

		// Ensure version value matches constant
		if vadMap.version != vmmMapVadVersion {
			return nil, errors.New(
				"failed to verify vad version value",
			)
		}

		base := uintptr(unsafe.Pointer(&vadMap.vads))
		size := unsafe.Sizeof(vmmMapVadEntry{})

		addr := start
		// Iterate through number of returned VADs
		for i := uint32(0); i < vadMap.count; i++ {

			// Convert the current pointer to VAD structure
			vadPtr := unsafe.Pointer(base + uintptr(i)*size)
			vad := (*vmmMapVadEntry)(vadPtr)

			vadStart := uintptr(vad.start) + 0
			vadStop := uintptr(vad.end) + 1

			// Skip VADs outside the requested range
			if vadStop <= start {
				continue
			}

			if vadStart >= stop {
				break
			}

			// Fill any gaps before the current VAD
			if addr < vadStart {

				region := &Region{
					Start: addr,
					Stop:  vadStart,
					Size:  vadStart - addr,
				}

				// Add invalid region to result
				result = append(result, region)
				addr = vadStart
			}

			// Decode various VAD flags
			flags0 := vad.flags0.Decode()
			flags1 := vad.flags1.Decode()

			// In Windows, it's possible to allocate virtual memory that is not
			// initially backed by physical pages — for example, memory that is
			// only committed or paged in upon first access.
			//
			// The presence of physical backing can be determined using
			// `MiGetWorkingSetInfoList`, which inspects PFNs (Page Frame Numbers)
			// associated with each virtual page.
			//
			// Some applications use this type of lazy allocation intentionally
			// to detect memory scanners or debuggers, altering behavior if the
			// memory is accessed prematurely.
			//
			// Since `VMMDLL_MemRead` does *not* fault in or back uncommitted memory
			// upon access, the `bound` flag is still valid for most scanning use cases.
			//
			// If more accurate tracking of physical residency is required,
			// a future implementation could read the PFNs and validate them manually.
			//
			// Demo: https://gist.github.com/dkrutsko/d6118638b0ef711b30bfcfe5b083d067
			bound := flags1.isCommitted

			region := &Region{
				Valid: true,
				Bound: bound,

				Start: vadStart,
				Stop:  vadStop,
				Size:  vadStop - vadStart,

				Readable:   (flags0.protection & ACCESS_PAGE_R) != 0,
				Writable:   (flags0.protection & ACCESS_PAGE_W) != 0,
				Executable: (flags0.protection & ACCESS_PAGE_X) != 0,
				Access:     flags0.protection,

				Private: flags0.isPrivateMemory,
				Guarded: (flags0.protection & PAGE_GUARD) != 0,
			}

			// Add current region to result
			result = append(result, region)
			addr = vadStop
		}

		// Fill final gap
		if addr < stop {

			region := &Region{
				Start: addr,
				Stop:  stop,
				Size:  stop - addr,
			}

			// Add invalid region to result
			result = append(result, region)
		}
	}

	//----------------------------------------------------------------------------//

	/*
		// Anything past `GetMaxAddress` cannot be retrieved using VADs. Those regions
		// belong to kernel space, and thus must be retrieved using PTEs. Below is some
		// code which retrieves PTEs but does nothing with them. Complete this section
		// if support for kernel region mapping is required.

		var mapPtePtr uintptr
		// Attempt to retrieve the PTE information
		success, _, err = vmmDll.mapGetPteW.Call(
			m.leech.handle,
			uintptr(m.proc.pid),
			0,
			uintptr(unsafe.Pointer(&mapPtePtr)),
		)
		if err.(windows.Errno) != 0 {
			return nil, errors.New(
				"failed to get pte info",
				errors.Error("error", err),
			)
		}
		if success == 0 {
			return nil, errors.New(
				"failed to get pte info",
			)
		}

		defer func() {
			// Attempt to free the allocated memory
			_, _, _ = vmmDll.memFree.Call(mapPtePtr)
		}()

		//----------------------------------------------------------------------------//

		{
			// Convert returned pointer into PTE structure
			pteMap := (*vmmMapPte)(unsafe.Pointer(mapPtePtr))

			// Ensure version value matches constant
			if pteMap.version != vmmMapPteVersion {
				return nil, errors.New(
					"failed to verify pte version value",
				)
			}

			base := uintptr(unsafe.Pointer(&pteMap.ptes))
			size := unsafe.Sizeof(vmmMapPteEntry{})

			// Iterate through number of returned PTEs
			for i := uint32(0); i < pteMap.count; i++ {

				// Convert the current pointer to PTE structure
				ptePtr := unsafe.Pointer(base + uintptr(i)*size)
				pte := (*vmmMapPteEntry)(ptePtr)

				// NYI:
			}
		}
	*/

	//----------------------------------------------------------------------------//

	return result, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) ClearCache() {

	// TODO: Caching needs a mutex
	// NYI:
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) GetPtrSize() uintptr {

	// If 64-bit process
	if m.proc.is64Bit {
		return 8
	} else {
		return 4
	}
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) GetMinAddress() uintptr {

	return 0x10000
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) GetMaxAddress() uintptr {

	// We assume that this code is always running on a 64-bit Windows host.
	// On such systems, 32-bit processes (WOW64) typically have a 2 GB virtual
	// address space limit (up to 0x7FFF0000). If the process is linked with
	// the `/LARGEADDRESSAWARE` flag, it may instead use up to 4 GB (0xFFFF0000).
	//
	// Native 64-bit processes have access to a much larger address space,
	// theoretically up to 128 TB (0x00007FFFFFFFFFFF). In practice, usable
	// user-mode memory tops out below that, and we conservatively round to
	// 0x800000000000 as a working limit.
	//
	// To simplify our memory model and avoid having to dynamically read
	// `EPROCESS->HighestUserAddress` for every process, we use constants.
	//
	// These values are **intentionally rounded up** to skip over reserved
	// memory regions at both ends of the user-mode address space. Windows
	// reserves:
	//   - The first 64 KB (0x00000000 to 0x0000FFFF) for null pointer and
	//     stack overflow detection.
	//   - The last few pages before the kernel-mode boundary for guard pages
	//     and debugging traps.
	//
	// Therefore, we treat the "max user address" not only as the upper bound
	// of valid user memory, but also as the point where kernel pages begin.
	// This allows us to easily check for user vs kernel space addresses by
	// comparing against this boundary.

	// If 64-bit process
	if m.proc.is64Bit {
		return 0x800000000000
	} else {
		return 0x80000000
	}
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) GetPageSize() uintptr {

	return 0x1000
}

////////////////////////////////////////////////////////////////////////////////

/*func (m *Memory) ReadData(address uintptr, result void*, length uintptr) uintptr {

	// TODO: Caching needs a mutex

	// NYI
	return 0
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) ReadType(address uintptr, result void*, length uintptr) uintptr {

	// NYI
	return 0
}*/

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) ReadInt8(address uintptr, count uint32, stride uint32) int8 {

	// NYI:
	return 0
	//return m.readType(address, native.Memory._TYPE_INT8, 1, count, stride)
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) ReadInt16(address uintptr, count uint32, stride uint32) int16 {

	// NYI:
	return 0
	//return m.readType(address, native.Memory._TYPE_INT16, 2, count, stride)
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) ReadInt32(address uintptr, count uint32, stride uint32) int32 {

	// NYI:
	return 0
	//return m.readType(address, native.Memory._TYPE_INT32, 4, count, stride)
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) ReadInt64(address uintptr, count uint32, stride uint32) int64 {

	// NYI:
	return 0
	//return m.readType(address, native.Memory._TYPE_INT64, 8, count, stride)
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) ReadFloat32(address uintptr, count uint32, stride uint32) float32 {

	// NYI:
	return 0
	//return m.readType(address, native.Memory._TYPE_FLOAT32, 4, count, stride)
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) ReadFloat64(address uintptr, count uint32, stride uint32) float64 {

	// NYI:
	return 0
	//return m.readType(address, native.Memory._TYPE_FLOAT64, 8, count, stride)
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) ReadPtr(address uintptr, count uint32, stride uint32) uintptr {

	// NYI:
	return 0

	// If 64-bit process
	/*if m.proc.is64Bit {
		return m.readType(address, native.Memory._TYPE_INT64, 8, count, stride)
	} else {
		return m.readType(address, native.Memory._TYPE_INT32, 4, count, stride)
	}*/
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) ReadBool(address uintptr, count uint32, stride uint32) bool {

	// NYI:
	return false
	//return m.readType(address, native.Memory._TYPE_BOOL, 1, count, stride)
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) ReadString(address uintptr, length uint32, count uint32, stride uint32) string {

	// NYI:
	return ""
	//return m.readType(address, native.Memory._TYPE_STRING, length, count, stride)
}
