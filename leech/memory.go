package leech

import (
	"encoding/binary"
	"math"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/dkrutsko/oasis/errors"
)

////////////////////////////////////////////////////////////////////////////////

type Memory struct {
	leech *Leech
	proc  *Process

	cache map[uintptr][]byte
	lock  sync.Mutex
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) IsValid() bool {

	// If process is valid
	return m.proc.IsValid()
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) GetProcess() *Process {
	return m.proc
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) GetRegions(start uintptr, stop uintptr) ([]*Region, error) {

	//----------------------------------------------------------------------------//

	var result []*Region

	// Lock for retrieval
	m.leech.lock.RLock()
	defer m.leech.lock.RUnlock()

	// Make sure that handle and PID are valid
	if m.leech.handle == 0 || m.proc.pid == 0 {
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
			errors.Uint32("pid", m.proc.pid),
			errors.Error("error", err),
		)
	}
	if success == 0 {
		return nil, errors.New(
			"failed to get vad info",
			errors.Uint32("pid", m.proc.pid),
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
				errors.Uint32("pid", m.proc.pid),
				errors.Error("error", err),
			)
		}
		if success == 0 {
			return nil, errors.New(
				"failed to get pte info",
				errors.Uint32("pid", m.proc.pid),
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

	m.lock.Lock()
	defer m.lock.Unlock()

	// Clear all data in memory cache
	m.cache = make(map[uintptr][]byte)
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

func (m *Memory) readPage(address uintptr) ([]byte, error) {

	//----------------------------------------------------------------------------//

	// Retrieve size of a page
	pageSize := m.GetPageSize()

	// Check if address aligned
	if address%pageSize != 0 {
		return nil, errors.New("address is unaligned")
	}

	// Lock for retrieval
	m.leech.lock.RLock()
	defer m.leech.lock.RUnlock()

	// Make sure that handle and PID are valid
	if m.leech.handle == 0 || m.proc.pid == 0 {
		return nil, errors.New("process is not valid")
	}

	//----------------------------------------------------------------------------//

	var bytesRead uint32
	// Make buffer to hold the data
	result := make([]byte, pageSize)

	// Attempt to read the memory of the page
	success, _, err := vmmDll.memReadEx.Call(
		m.leech.handle,
		uintptr(m.proc.pid),
		address,
		uintptr(unsafe.Pointer(&result[0])),
		pageSize,
		uintptr(unsafe.Pointer(&bytesRead)),
		uintptr(0x1), // VMMDLL_FLAG_NOCACHE
	)
	if err.(windows.Errno) != 0 {
		return nil, errors.New(
			"failed to read page",
			errors.Uint32("pid", m.proc.pid),
			errors.Uintptr("address", address),
			errors.Error("error", err),
		)
	}
	if success == 0 {
		return nil, errors.New(
			"failed to read page",
			errors.Uint32("pid", m.proc.pid),
			errors.Uintptr("address", address),
		)
	}

	if uintptr(bytesRead) != pageSize {
		return result, errors.New(
			"not enough bytes have been read",
			errors.Uint32("pid", m.proc.pid),
			errors.Uintptr("address", address),
			errors.Uint32("read", bytesRead),
		)
	}

	//----------------------------------------------------------------------------//

	return result, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) ReadData(address uintptr, length uintptr) ([]byte, error) {

	// Check length
	if length == 0 {
		return nil, errors.New("length must be greater than zero")
	}

	// Address is within min bounds
	if address < m.GetMinAddress() {
		return nil, errors.New("address is below minimum address")
	}

	// Address is within max bounds
	if address+length > m.GetMaxAddress() {
		return nil, errors.New("address is above maximum address")
	}

	var offset uintptr = 0
	// Make buffer to hold result
	result := make([]byte, length)

	// Retrieve size of a page
	pageSize := m.GetPageSize()

	m.lock.Lock()
	defer m.lock.Unlock()

	// Perform page reads
	for offset < length {

		// Calculate alignment to current start of page
		aligned := (address + offset) &^ (pageSize - 1)

		// Check if page is in cache
		cache, ok := m.cache[aligned]
		if !ok {
			// Try and read the entire page
			data, err := m.readPage(aligned)
			if err != nil {
				return nil, err
			}

			// Cache read page data
			m.cache[aligned] = data
			cache = data
		}

		offsetInPage := (address + offset) - aligned

		bytesToCopy := pageSize - offsetInPage
		// Calculate how much to copy from this page
		if rem := length - offset; bytesToCopy > rem {
			bytesToCopy = rem
		}

		// Copy portion from cached page into result
		copy(result[offset:], cache[offsetInPage:offsetInPage+bytesToCopy])
		offset += bytesToCopy
	}

	return result, nil
}

////////////////////////////////////////////////////////////////////////////////

type MemType int

const (
	MemTypeBuffer  MemType = 0x0
	MemTypeInt8    MemType = 0x1
	MemTypeInt16   MemType = 0x2
	MemTypeInt32   MemType = 0x3
	MemTypeInt64   MemType = 0x4
	MemTypeFloat32 MemType = 0x5
	MemTypeFloat64 MemType = 0x6
	MemTypeBool    MemType = 0x7
	MemTypeString  MemType = 0x8
)

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) ReadTypes(
	address uintptr,
	memType MemType,
	length, count, stride uint32,
) ([]any, error) {

	// If there's anything to read
	if count == 0 || length == 0 {
		return nil, nil
	}

	// Default stride
	if stride == 0 {
		stride = length
	}

	// Check the stride
	if stride < length {
		return nil, errors.New(
			"stride is too small",
			errors.Uint32("stride", stride),
			errors.Uint32("length", length),
		)
	}

	size := count*stride + length - stride
	// Try and read all required memory in one call
	data, err := m.ReadData(address, uintptr(size))
	if err != nil {
		return nil, err
	}

	offset := uint32(0)
	// Create all the result data
	results := make([]any, count)

	// Read for required number of values
	for i := uint32(0); i < count; i++ {

		slice := data[offset : offset+length]

		switch memType {
		case MemTypeBuffer:
			results[i] = slice

		case MemTypeInt8:
			results[i] = int8(slice[0])

		case MemTypeInt16:
			results[i] = int16(binary.LittleEndian.Uint16(slice))

		case MemTypeInt32:
			results[i] = int32(binary.LittleEndian.Uint32(slice))

		case MemTypeInt64:
			results[i] = int64(binary.LittleEndian.Uint64(slice))

		case MemTypeFloat32:
			bits := binary.LittleEndian.Uint32(slice)
			results[i] = math.Float32frombits(bits)

		case MemTypeFloat64:
			bits := binary.LittleEndian.Uint64(slice)
			results[i] = math.Float64frombits(bits)

		case MemTypeBool:
			results[i] = slice[0] != 0

		case MemTypeString:
			results[i] = cStrToString(slice)

		default:
			return nil, errors.New("unsupported data type")
		}

		offset += stride
	}

	return results, nil
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) ReadInt8(address uintptr) (int8, error) {

	res, err := m.ReadTypes(address, MemTypeInt8, 1, 1, 0)
	if err != nil {
		return 0, err
	}

	if len(res) != 1 {
		return 0, errors.New("not enough results for read type")
	}

	val, ok := res[0].(int8)
	if !ok {
		return 0, errors.New("failed to cast to the final type")
	}

	return val, nil
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) ReadInt8s(address uintptr, count, stride uint32) ([]int8, error) {

	res, err := m.ReadTypes(address, MemTypeInt8, 1, count, stride)
	if err != nil {
		return nil, err
	}

	if len(res) != int(count) {
		return nil, errors.New("not enough results for read type")
	}

	out := make([]int8, count)
	for i, v := range res {
		val, ok := v.(int8)
		if !ok {
			return nil, errors.New("failed to cast to final type")
		}

		out[i] = val
	}

	return out, nil
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) ReadInt16(address uintptr) (int16, error) {

	res, err := m.ReadTypes(address, MemTypeInt16, 2, 1, 0)
	if err != nil {
		return 0, err
	}

	if len(res) != 1 {
		return 0, errors.New("not enough results for read type")
	}

	val, ok := res[0].(int16)
	if !ok {
		return 0, errors.New("failed to cast to the final type")
	}

	return val, nil
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) ReadInt16s(address uintptr, count, stride uint32) ([]int16, error) {

	res, err := m.ReadTypes(address, MemTypeInt16, 2, count, stride)
	if err != nil {
		return nil, err
	}

	if len(res) != int(count) {
		return nil, errors.New("not enough results for read type")
	}

	out := make([]int16, count)
	for i, v := range res {
		val, ok := v.(int16)
		if !ok {
			return nil, errors.New("failed to cast to final type")
		}

		out[i] = val
	}

	return out, nil
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) ReadInt32(address uintptr) (int32, error) {

	res, err := m.ReadTypes(address, MemTypeInt32, 4, 1, 0)
	if err != nil {
		return 0, err
	}

	if len(res) != 1 {
		return 0, errors.New("not enough results for read type")
	}

	val, ok := res[0].(int32)
	if !ok {
		return 0, errors.New("failed to cast to the final type")
	}

	return val, nil
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) ReadInt32s(address uintptr, count, stride uint32) ([]int32, error) {

	res, err := m.ReadTypes(address, MemTypeInt32, 4, count, stride)
	if err != nil {
		return nil, err
	}

	if len(res) != int(count) {
		return nil, errors.New("not enough results for read type")
	}

	out := make([]int32, count)
	for i, v := range res {
		val, ok := v.(int32)
		if !ok {
			return nil, errors.New("failed to cast to final type")
		}

		out[i] = val
	}

	return out, nil
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) ReadInt64(address uintptr) (int64, error) {

	res, err := m.ReadTypes(address, MemTypeInt64, 8, 1, 0)
	if err != nil {
		return 0, err
	}

	if len(res) != 1 {
		return 0, errors.New("not enough results for read type")
	}

	val, ok := res[0].(int64)
	if !ok {
		return 0, errors.New("failed to cast to the final type")
	}

	return val, nil
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) ReadInt64s(address uintptr, count, stride uint32) ([]int64, error) {

	res, err := m.ReadTypes(address, MemTypeInt64, 8, count, stride)
	if err != nil {
		return nil, err
	}

	if len(res) != int(count) {
		return nil, errors.New("not enough results for read type")
	}

	out := make([]int64, count)
	for i, v := range res {
		val, ok := v.(int64)
		if !ok {
			return nil, errors.New("failed to cast to final type")
		}

		out[i] = val
	}

	return out, nil
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) ReadFloat32(address uintptr) (float32, error) {

	res, err := m.ReadTypes(address, MemTypeFloat32, 4, 1, 0)
	if err != nil {
		return 0, err
	}

	if len(res) != 1 {
		return 0, errors.New("not enough results for read type")
	}

	val, ok := res[0].(float32)
	if !ok {
		return 0, errors.New("failed to cast to the final type")
	}

	return val, nil
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) ReadFloat32s(address uintptr, count, stride uint32) ([]float32, error) {

	res, err := m.ReadTypes(address, MemTypeFloat32, 4, count, stride)
	if err != nil {
		return nil, err
	}

	if len(res) != int(count) {
		return nil, errors.New("not enough results for read type")
	}

	out := make([]float32, count)
	for i, v := range res {
		val, ok := v.(float32)
		if !ok {
			return nil, errors.New("failed to cast to final type")
		}

		out[i] = val
	}

	return out, nil
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) ReadFloat64(address uintptr) (float64, error) {

	res, err := m.ReadTypes(address, MemTypeFloat64, 8, 1, 0)
	if err != nil {
		return 0, err
	}

	if len(res) != 1 {
		return 0, errors.New("not enough results for read type")
	}

	val, ok := res[0].(float64)
	if !ok {
		return 0, errors.New("failed to cast to the final type")
	}

	return val, nil
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) ReadFloat64s(address uintptr, count, stride uint32) ([]float64, error) {

	res, err := m.ReadTypes(address, MemTypeFloat64, 8, count, stride)
	if err != nil {
		return nil, err
	}

	if len(res) != int(count) {
		return nil, errors.New("not enough results for read type")
	}

	out := make([]float64, count)
	for i, v := range res {
		val, ok := v.(float64)
		if !ok {
			return nil, errors.New("failed to cast to final type")
		}

		out[i] = val
	}

	return out, nil
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) ReadPtr(address uintptr) (uintptr, error) {

	//----------------------------------------------------------------------------//

	// If 64-bit process
	if m.proc.is64Bit {

		res, err := m.ReadTypes(address, MemTypeInt64, 8, 1, 0)
		if err != nil {
			return 0, err
		}

		if len(res) != 1 {
			return 0, errors.New("not enough results for read type")
		}

		val, ok := res[0].(int64)
		if !ok {
			return 0, errors.New("failed to cast to the final type")
		}

		return uintptr(val), nil
	}

	//----------------------------------------------------------------------------//

	res, err := m.ReadTypes(address, MemTypeInt32, 4, 1, 0)
	if err != nil {
		return 0, err
	}

	if len(res) != 1 {
		return 0, errors.New("not enough results for read type")
	}

	val, ok := res[0].(int32)
	if !ok {
		return 0, errors.New("failed to cast to the final type")
	}

	return uintptr(val), nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) ReadPtrs(address uintptr, count, stride uint32) ([]uintptr, error) {

	//----------------------------------------------------------------------------//

	// If 64-bit process
	if m.proc.is64Bit {

		res, err := m.ReadTypes(address, MemTypeInt64, 8, count, stride)
		if err != nil {
			return nil, err
		}

		if len(res) != int(count) {
			return nil, errors.New("not enough results for read type")
		}

		out := make([]uintptr, count)
		for i, v := range res {
			val, ok := v.(int64)
			if !ok {
				return nil, errors.New("failed to cast to final type")
			}

			out[i] = uintptr(val)
		}

		return out, nil
	}

	//----------------------------------------------------------------------------//

	res, err := m.ReadTypes(address, MemTypeInt32, 4, count, stride)
	if err != nil {
		return nil, err
	}

	if len(res) != int(count) {
		return nil, errors.New("not enough results for read type")
	}

	out := make([]uintptr, count)
	for i, v := range res {
		val, ok := v.(int32)
		if !ok {
			return nil, errors.New("failed to cast to final type")
		}

		out[i] = uintptr(val)
	}

	return out, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) ReadBool(address uintptr) (bool, error) {

	res, err := m.ReadTypes(address, MemTypeBool, 1, 1, 0)
	if err != nil {
		return false, err
	}

	if len(res) != 1 {
		return false, errors.New("not enough results for read type")
	}

	val, ok := res[0].(bool)
	if !ok {
		return false, errors.New("failed to cast to the final type")
	}

	return val, nil
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) ReadBools(address uintptr, count, stride uint32) ([]bool, error) {

	res, err := m.ReadTypes(address, MemTypeBool, 1, count, stride)
	if err != nil {
		return nil, err
	}

	if len(res) != int(count) {
		return nil, errors.New("not enough results for read type")
	}

	out := make([]bool, count)
	for i, v := range res {
		val, ok := v.(bool)
		if !ok {
			return nil, errors.New("failed to cast to final type")
		}

		out[i] = val
	}

	return out, nil
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) ReadString(address uintptr, length uint32) (string, error) {

	res, err := m.ReadTypes(address, MemTypeString, length, 1, 0)
	if err != nil {
		return "", err
	}

	if len(res) != 1 {
		return "", errors.New("not enough results for read type")
	}

	val, ok := res[0].(string)
	if !ok {
		return "", errors.New("failed to cast to the final type")
	}

	return val, nil
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) ReadStrings(address uintptr, length, count, stride uint32) ([]string, error) {

	res, err := m.ReadTypes(address, MemTypeString, length, count, stride)
	if err != nil {
		return nil, err
	}

	if len(res) != int(count) {
		return nil, errors.New("not enough results for read type")
	}

	out := make([]string, count)
	for i, v := range res {
		val, ok := v.(string)
		if !ok {
			return nil, errors.New("failed to cast to final type")
		}

		out[i] = val
	}

	return out, nil
}
