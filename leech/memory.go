package leech

import (
	"context"
	"encoding/binary"
	"math"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"github.com/dkrutsko/oasis/errors"
)

////////////////////////////////////////////////////////////////////////////////

// Stats tracks memory operation counters for a Memory
// instance. Useful for measuring and optimizing the
// performance of a particular code path. All counters are
// cumulative since the last call to `ResetStats`.
type Stats struct {

	// SystemReads is the number of page reads sent to VMMDLL.
	SystemReads uint32

	// CachedReads is the number of reads served from the local
	// cache.
	CachedReads uint32

	// SystemWrites is the number of writes sent to VMMDLL.
	SystemWrites uint32

	// ReadErrors is the number of failed read operations.
	ReadErrors uint32

	// WriteErrors is the number of failed write operations.
	WriteErrors uint32
}

////////////////////////////////////////////////////////////////////////////////

// Memory provides typed read and write access to the virtual
// address space of a process via DMA. Use `CreateCache` to
// enable block-aligned caching for improved read performance.
// Use `ClearCache` to invalidate stale data when fresh reads
// are needed. To create a Memory instance, use
// `Process.GetMemory`. Memory is safe for concurrent use.
type Memory struct {
	leech *Leech
	proc  *Process

	cache      []byte
	cachePages map[uintptr]uintptr
	cacheNext  uintptr

	blockLength uintptr
	blockBuffer uintptr
	cacheSize   uintptr
	enlargeSize uintptr
	maximumSize uintptr

	lock  sync.Mutex
	stats Stats
}

////////////////////////////////////////////////////////////////////////////////

// MemType identifies a primitive data type for use with
// ReadTypes.
type MemType int

const (
	MemTypeBuffer  MemType = 0x0 // Raw byte buffer
	MemTypeInt8    MemType = 0x1 // Signed 8-bit integer
	MemTypeInt16   MemType = 0x2 // Signed 16-bit integer
	MemTypeInt32   MemType = 0x3 // Signed 32-bit integer
	MemTypeInt64   MemType = 0x4 // Signed 64-bit integer
	MemTypeFloat32 MemType = 0x5 // 32-bit IEEE float
	MemTypeFloat64 MemType = 0x6 // 64-bit IEEE float
	MemTypeBool    MemType = 0x7 // Boolean (1 byte)
	MemTypeString  MemType = 0x8 // Null-terminated C string
)

////////////////////////////////////////////////////////////////////////////////

// IsValid returns true if the underlying process is still
// valid on the target system.
func (m *Memory) IsValid() bool {

	// If process is valid
	return m.proc.IsValid()
}

////////////////////////////////////////////////////////////////////////////////

// GetProcess returns the process this Memory is attached to.
func (m *Memory) GetProcess() *Process {
	return m.proc
}

////////////////////////////////////////////////////////////////////////////////

// GetStats returns a snapshot of the current operation
// counters. Useful for measuring and optimizing the
// performance of a particular code path.
func (m *Memory) GetStats() Stats {

	m.lock.Lock()
	defer m.lock.Unlock()

	return m.stats
}

////////////////////////////////////////////////////////////////////////////////

// ResetStats zeroes all operation counters.
func (m *Memory) ResetStats() {

	m.lock.Lock()
	defer m.lock.Unlock()

	m.stats = Stats{}
}

////////////////////////////////////////////////////////////////////////////////

// GetRegion returns the memory region containing the given
// address. Returns (nil, nil) if the address is unmapped.
func (m *Memory) GetRegion(ctx context.Context, address uintptr) (*Region, error) {

	//----------------------------------------------------------------------------//

	// Get regions covering a single page at the address
	regions, err := m.GetRegions(ctx, address, address+1)
	if err != nil {
		return nil, err
	}

	//----------------------------------------------------------------------------//

	// Find the region that contains the address
	for _, region := range regions {
		if region.Contains(address) {
			return region, nil
		}
	}

	return nil, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// GetRegions returns all memory regions within the address
// range [start, stop). User-space regions are retrieved from
// VADs and kernel-space regions (above `GetMaxUserAddress`)
// from PTEs. Gaps between mapped regions are filled with
// invalid placeholder regions. Both `start` and `stop` are
// aligned to page boundaries. The resulting list is sorted
// from start to stop and includes both bound and unbound
// regions. Returns an empty list if the range is empty.
func (m *Memory) GetRegions(ctx context.Context, start uintptr, stop uintptr) ([]*Region, error) {

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
	if start < m.GetMinUserAddress() {
		start = m.GetMinUserAddress()
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
	success, err := vmmCall(
		vmmDll.mapGetVad,
		m.leech.handle,
		uintptr(m.proc.pid),
		0,
		uintptr(unsafe.Pointer(&mapVadPtr)),
	)
	if err != nil {
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
		_, _ = vmmCall(vmmDll.memFree, mapVadPtr)
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

			if ctx.Err() != nil {
				return result, ctx.Err()
			}

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

			// In Windows, it's possible to allocate virtual memory that is
			// not initially backed by physical pages - for example, memory
			// that is only committed or paged in upon first access.
			//
			// Physical residency can be determined by walking the PFN
			// (Page Frame Number) database via DMA, but we use the VAD
			// committed flag as a practical approximation instead.
			//
			// Some applications use lazy allocation intentionally to
			// detect memory scanners or debuggers, altering behavior
			// if the memory is accessed prematurely.
			//
			// Since `VMMDLL_MemRead` does not fault in uncommitted memory
			// upon access, the `bound` flag is still valid for most
			// scanning use cases.
			//
			// Demo: https://gist.github.com/dkrutsko/d6118638b0ef711b30bfcfe5b083d067
			bound := flags1.isCommitted

			region := &Region{
				Valid: true,
				Bound: bound,

				Start: vadStart,
				Stop:  vadStop,
				Size:  vadStop - vadStart,

				Readable:   (flags0.protection & AccessPageR) != 0,
				Writable:   (flags0.protection & AccessPageW) != 0,
				Executable: (flags0.protection & AccessPageX) != 0,
				Access:     flags0.protection,

				Private: flags0.isPrivateMemory,
				Guarded: (flags0.protection & PageGuard) != 0,
			}

			// Add current region to result
			result = append(result, region)
			addr = vadStop
		}

		// Fill final gap up to the user/kernel boundary.
		// Anything past `GetMaxUserAddress` is handled
		// by the PTE section below.
		vadStop := stop
		if vadStop > m.GetMaxUserAddress() {
			vadStop = m.GetMaxUserAddress()
		}

		if addr < vadStop {

			region := &Region{
				Start: addr,
				Stop:  vadStop,
				Size:  vadStop - addr,
			}

			// Add invalid region to result
			result = append(result, region)
		}
	}

	//----------------------------------------------------------------------------//

	// Anything past `GetMaxUserAddress` cannot be retrieved using
	// VADs. Those regions belong to kernel space, and thus must be
	// retrieved using PTEs. Each PTE entry represents a contiguous
	// run of pages with the same flags, so they map directly to
	// regions.

	if stop > m.GetMaxUserAddress() {

		var mapPtePtr uintptr
		// Attempt to retrieve the PTE information
		success, err = vmmCall(
			vmmDll.mapGetPte,
			m.leech.handle,
			uintptr(m.proc.pid),
			0,
			uintptr(unsafe.Pointer(&mapPtePtr)),
		)
		if err != nil {
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
			_, _ = vmmCall(vmmDll.memFree, mapPtePtr)
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
			pageSize := m.GetPageSize()

			// Iterate through number of returned PTEs
			for i := uint32(0); i < pteMap.count; i++ {

				if ctx.Err() != nil {
					return result, ctx.Err()
				}

				// Convert the current pointer to PTE structure
				ptePtr := unsafe.Pointer(base + uintptr(i)*size)
				pte := (*vmmMapPteEntry)(ptePtr)

				pteStart := uintptr(pte.base)
				pteStop := pteStart + uintptr(pte.pageCount)*pageSize

				// Only include PTEs in the kernel region
				if pteStart < m.GetMaxUserAddress() {
					continue
				}

				// Skip PTEs outside the requested range
				if pteStop <= start {
					continue
				}

				if pteStart >= stop {
					break
				}

				// Decode PTE page flags. PTEs present in the map
				// are always readable. Writable and executable are
				// derived from the hardware page flags.
				flags := pte.pageFlags.Decode()

				region := &Region{
					Valid: true,
					Bound: true,

					Start: pteStart,
					Stop:  pteStop,
					Size:  pteStop - pteStart,

					Readable:   true,
					Writable:   flags.isWritable,
					Executable: !flags.isNoExecute,
				}

				// Add current region to result
				result = append(result, region)
			}
		}
	}

	//----------------------------------------------------------------------------//

	return result, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

func parsePattern(pattern string) ([]byte, []bool, error) {

	// Split pattern into hex tokens
	tokens := strings.Fields(pattern)
	if len(tokens) == 0 {
		return nil, nil, errors.New("pattern is empty")
	}

	bytes := make([]byte, len(tokens))
	mask := make([]bool, len(tokens))

	for i, token := range tokens {

		// Check for wildcard token
		if token == "?" || token == "??" {
			mask[i] = true
			continue
		}

		// Parse hex byte value
		val, err := strconv.ParseUint(token, 16, 8)
		if err != nil {
			return nil, nil, errors.New(
				"invalid pattern token",
				errors.String("token", token),
			)
		}

		bytes[i] = byte(val)
	}

	return bytes, mask, nil
}

////////////////////////////////////////////////////////////////////////////////

// Find searches for a byte pattern within the address range
// [start, stop) and returns the addresses of all matches.
// The pattern is a space-separated string of two-digit hex
// bytes where "?" or "??" denotes a wildcard that matches
// any byte. For example: "48 8B 45 ? 48 89 C7". Searching
// is performed region by region, reading in 4 MB chunks.
// Only valid, bound, and readable regions are scanned.
// Reads bypass the cache to avoid memory pressure during
// large scans. If `limit` is greater than zero, the search
// stops after that many matches. Returns an empty list if
// the pattern is invalid or no matches are found.
func (m *Memory) Find(
	ctx context.Context,
	pattern string,
	start, stop uintptr,
	limit int,
) ([]uintptr, error) {

	//----------------------------------------------------------------------------//

	// Parse the pattern into bytes and wildcard mask
	patBytes, patMask, err := parsePattern(pattern)
	if err != nil {
		return nil, err
	}

	patLen := uintptr(len(patBytes))

	//----------------------------------------------------------------------------//

	// Retrieve all regions in the requested range
	regions, err := m.GetRegions(ctx, start, stop)
	if err != nil {
		return nil, err
	}

	//----------------------------------------------------------------------------//

	var results []uintptr
	chunkSize := uintptr(4 * 1024 * 1024) // 4 MB
	overlap := patLen - 1

	// Iterate through each scannable region
	for _, region := range regions {

		// Skip regions that cannot be scanned
		if !region.Valid || !region.Bound || !region.Readable {
			continue
		}

		regionStart := region.Start
		regionStop := region.Stop

		// Clamp to requested range
		if regionStart < start {
			regionStart = start
		}
		if regionStop > stop {
			regionStop = stop
		}
		if regionStart >= regionStop {
			continue
		}

		// Read region in chunks with overlap
		addr := regionStart
		for addr < regionStop {

			if ctx.Err() != nil {
				return results, ctx.Err()
			}

			readEnd := addr + chunkSize
			if readEnd > regionStop {
				readEnd = regionStop
			}

			readLen := readEnd - addr

			// Read a chunk of memory directly without
			// polluting the cache with scan data
			data, err := m.readDirect(addr, readLen)
			if err != nil {
				// Skip unreadable chunks
				addr = readEnd
				continue
			}

			// Scan the chunk for pattern matches
			scanLen := uintptr(len(data))
			if scanLen < patLen {
				break
			}

			for i := uintptr(0); i <= scanLen-patLen; i++ {

				match := true
				for j := uintptr(0); j < patLen; j++ {
					if !patMask[j] && data[i+j] != patBytes[j] {
						match = false
						break
					}
				}

				if match {
					results = append(results, addr+i)

					// Check if limit has been reached
					if limit > 0 && len(results) >= limit {
						return results, nil
					}
				}
			}

			// Advance by chunk minus overlap to catch
			// patterns spanning chunk boundaries
			if readLen > overlap {
				addr += readLen - overlap
			} else {
				addr = readEnd
			}
		}
	}

	//----------------------------------------------------------------------------//

	return results, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// CreateCache engages the memory caching system with a pre-
// allocated buffer. `blockLength` is the aligned block size
// (must be a power of two). `blockBuffer` is the maximum
// amount of data that can be read in a single cached read.
// The sum of `blockLength` and `blockBuffer` defines the
// amount of data read per native read. `initialSize` is the
// starting buffer size. The cache grows by `enlargeSize`
// when full (0 = no growth, falls back to native reads).
// `maximumSize` caps the buffer (0 = unlimited). All values
// must be divisible by `GetPageSize`. Returns an error if
// any parameter is invalid.
func (m *Memory) CreateCache(
	blockLength, blockBuffer,
	initialSize, enlargeSize,
	maximumSize uintptr,
) error {

	//----------------------------------------------------------------------------//

	pageSize := m.GetPageSize()

	// Validate non-zero required parameters
	if blockLength == 0 || blockBuffer == 0 || initialSize == 0 {
		return errors.New("cache parameters must not be zero")
	}

	// Ensure alignment to page size
	if blockLength%pageSize != 0 ||
		blockBuffer%pageSize != 0 ||
		initialSize%pageSize != 0 ||
		enlargeSize%pageSize != 0 ||
		maximumSize%pageSize != 0 {
		return errors.New("cache parameters must be page-aligned")
	}

	// Block length must be a power of two
	if blockLength&(blockLength-1) != 0 {
		return errors.New("block length must be a power of two")
	}

	// Block length must be able to store the buffer
	if blockLength < blockBuffer {
		return errors.New("block length must be >= block buffer")
	}

	// Initial size must hold at least one block
	entrySize := blockLength + blockBuffer
	if initialSize < entrySize {
		return errors.New("initial size must be >= block length + block buffer")
	}

	// Validate enlarge size
	if enlargeSize != 0 && enlargeSize < entrySize {
		return errors.New("enlarge size must be >= block length + block buffer")
	}

	// Validate maximum size
	if maximumSize != 0 && maximumSize < initialSize {
		return errors.New("maximum size must be >= initial size")
	}

	//----------------------------------------------------------------------------//

	m.lock.Lock()
	defer m.lock.Unlock()

	m.blockLength = blockLength
	m.blockBuffer = blockBuffer
	m.cacheSize = initialSize
	m.enlargeSize = enlargeSize
	m.maximumSize = maximumSize
	m.cacheNext = 0

	// Pre-allocate the cache buffer
	m.cache = make([]byte, initialSize)
	m.cachePages = make(map[uintptr]uintptr)

	return nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// ClearCache removes all cached data without reallocating
// the buffer. Call this function when the latest version of
// the memory needs to be read, as the data in the cache may
// be out of date.
func (m *Memory) ClearCache() {

	m.lock.Lock()
	defer m.lock.Unlock()

	m.cacheNext = 0
	clear(m.cachePages)
}

////////////////////////////////////////////////////////////////////////////////

// DeleteCache disables caching and frees the buffer.
func (m *Memory) DeleteCache() {

	m.lock.Lock()
	defer m.lock.Unlock()

	m.blockLength = 0
	m.blockBuffer = 0
	m.cacheSize = 0
	m.enlargeSize = 0
	m.maximumSize = 0
	m.cacheNext = 0

	m.cache = nil
	m.cachePages = nil
}

////////////////////////////////////////////////////////////////////////////////

// IsCaching returns true if a cache has been created via
// `CreateCache`.
func (m *Memory) IsCaching() bool {

	m.lock.Lock()
	defer m.lock.Unlock()

	return m.cache != nil
}

////////////////////////////////////////////////////////////////////////////////

// GetCacheSize returns the current allocated size of the
// cache buffer, in bytes. The cache may grow depending on
// the cache parameters and how much data has been read.
func (m *Memory) GetCacheSize() uintptr {

	m.lock.Lock()
	defer m.lock.Unlock()

	return m.cacheSize
}

////////////////////////////////////////////////////////////////////////////////

// GetPtrSize returns the size of a single pointer in the
// target process, in bytes. Returns 8 for 64-bit and 4 for
// 32-bit processes.
func (m *Memory) GetPtrSize() uintptr {

	// If 64-bit process
	if m.proc.is64Bit {
		return 8
	} else {
		return 4
	}
}

////////////////////////////////////////////////////////////////////////////////

// GetMinUserAddress returns the minimum accessible user-mode
// address for the target process. The first 64 KB is reserved
// by Windows for null pointer detection.
func (m *Memory) GetMinUserAddress() uintptr {

	return 0x10000
}

////////////////////////////////////////////////////////////////////////////////

// GetMaxUserAddress returns the maximum accessible user-mode
// address for the target process. This also serves as the
// boundary between VAD and PTE region mapping in
// `GetRegions`.
func (m *Memory) GetMaxUserAddress() uintptr {

	// The DMA target is always a Windows system. On 64-bit Windows,
	// user-mode virtual addresses span 0 to 0x7FFFFFFFFFFF (128 TB).
	// This boundary is set by the OS regardless of whether the CPU
	// supports 5-level paging (LA57). 0x800000000000 (2^47) marks
	// the start of the non-canonical hole and the end of user space.
	//
	// For 32-bit processes (WOW64), the default user-mode limit is
	// 2 GB (0x80000000). Processes linked with `/LARGEADDRESSAWARE`
	// may use up to 4 GB, but we use the conservative 2 GB default
	// to avoid reading `EPROCESS->HighestUserAddress` per process.
	//
	// This value also serves as the boundary between VAD-based
	// region mapping (user space) and PTE-based region mapping
	// (kernel space) in `GetRegions`.

	// If 64-bit process
	if m.proc.is64Bit {
		return 0x800000000000
	} else {
		return 0x80000000
	}
}

////////////////////////////////////////////////////////////////////////////////

// GetPageSize returns the size of a single page of memory,
// in bytes. This value is typically 4096.
func (m *Memory) GetPageSize() uintptr {

	return 0x1000
}

////////////////////////////////////////////////////////////////////////////////

// ReadData reads `length` bytes from the target process
// starting at `address`. If a cache has been created via
// `CreateCache` and `length` fits within `blockBuffer`,
// reads are served from the cache. Otherwise, a direct
// native read is performed. Use `ClearCache` when fresh
// data is needed. Returns an error if `length` is zero,
// `address` is in the null pointer reserved region, or
// the read fails.
func (m *Memory) ReadData(address uintptr, length uintptr) ([]byte, error) {

	//----------------------------------------------------------------------------//

	// Check length
	if length == 0 {
		return nil, errors.New("length must be greater than zero")
	}

	// Reject reads in the null pointer reserved region
	if address < m.GetMinUserAddress() {
		return nil, errors.New("address is below minimum address")
	}

	//----------------------------------------------------------------------------//

	m.lock.Lock()
	defer m.lock.Unlock()

	// If caching is not enabled or the read is too large
	// for the cache, perform a direct native read
	if m.cache == nil || length > m.blockBuffer {
		return m.readNative(address, length)
	}

	//----------------------------------------------------------------------------//

	// Compute block-aligned address and entry size
	aligned := address &^ (m.blockLength - 1)
	entrySize := m.blockLength + m.blockBuffer

	// Check if the block has already been cached
	if _, ok := m.cachePages[aligned]; !ok {

		// Check if there is room in the cache buffer
		if m.cacheSize-m.cacheNext < entrySize {

			// Try to grow the cache buffer
			if m.enlargeSize == 0 {
				return m.readNative(address, length)
			}

			newSize := m.cacheSize + m.enlargeSize
			if m.maximumSize != 0 && newSize > m.maximumSize {
				return m.readNative(address, length)
			}

			// Allocate a larger buffer and copy existing data
			newCache := make([]byte, newSize)
			copy(newCache, m.cache[:m.cacheNext])
			m.cache = newCache
			m.cacheSize = newSize
		}

		// Read the full block from the target process
		slot := m.cache[m.cacheNext : m.cacheNext+entrySize]
		err := m.readInto(aligned, slot)
		if err != nil {
			// Fall back to a direct read on failure
			return m.readNative(address, length)
		}

		// Record the slot offset in the page map
		m.cachePages[aligned] = m.cacheNext
		m.cacheNext += entrySize
	}

	m.stats.CachedReads++

	// Return a subslice of the cache buffer. Callers must
	// finish processing the data before the next ClearCache
	// call, which reuses the same buffer for new reads.
	offset := m.cachePages[aligned] + (address - aligned)

	//----------------------------------------------------------------------------//

	return m.cache[offset : offset+length], nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// WriteData writes `data` to the target process starting at
// `address`. Affected cache pages are invalidated so that
// subsequent reads reflect the new data. Returns an error if
// `data` is empty, `address` is in the null pointer reserved
// region, or the write fails.
func (m *Memory) WriteData(address uintptr, data []byte) error {

	//----------------------------------------------------------------------------//

	length := uintptr(len(data))

	// Check length
	if length == 0 {
		return errors.New("length must be greater than zero")
	}

	// Reject writes in the null pointer reserved region
	if address < m.GetMinUserAddress() {
		return errors.New("address is below minimum address")
	}

	//----------------------------------------------------------------------------//

	// Retrieve size of a page
	pageSize := m.GetPageSize()

	m.lock.Lock()
	defer m.lock.Unlock()

	var offset uintptr = 0
	// Perform page writes
	for offset < length {

		// Calculate alignment to current start of page
		aligned := (address + offset) &^ (pageSize - 1)

		offsetInPage := (address + offset) - aligned

		bytesToWrite := pageSize - offsetInPage
		// Calculate how much to write to this page
		if rem := length - offset; bytesToWrite > rem {
			bytesToWrite = rem
		}

		// Write the chunk to the target address
		err := m.writePage(
			address+offset,
			data[offset:offset+bytesToWrite],
		)
		if err != nil {
			return err
		}

		// Invalidate cached block covering this address
		if m.cache != nil && m.blockLength > 0 {
			blockAligned := (address + offset) &^ (m.blockLength - 1)
			delete(m.cachePages, blockAligned)
		}

		offset += bytesToWrite
	}

	//----------------------------------------------------------------------------//

	return nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) readDirect(address uintptr, length uintptr) ([]byte, error) {

	//----------------------------------------------------------------------------//

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
	result := make([]byte, length)

	// Attempt to read the memory directly
	success, err := vmmCall(
		vmmDll.memReadEx,
		m.leech.handle,
		uintptr(m.proc.pid),
		address,
		uintptr(unsafe.Pointer(&result[0])),
		length,
		uintptr(unsafe.Pointer(&bytesRead)),
		uintptr(0x1), // VMMDLL_FLAG_NOCACHE
	)
	if err != nil {
		return nil, errors.New(
			"failed to read memory",
			errors.Uint32("pid", m.proc.pid),
			errors.Uintptr("address", address),
			errors.Error("error", err),
		)
	}
	if success == 0 {
		return nil, errors.New(
			"failed to read memory",
			errors.Uint32("pid", m.proc.pid),
			errors.Uintptr("address", address),
		)
	}

	//----------------------------------------------------------------------------//

	return result[:bytesRead], nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) readNative(address uintptr, length uintptr) ([]byte, error) {

	//----------------------------------------------------------------------------//

	result := make([]byte, length)

	// Single DMA call for the full read. VMMDLL handles
	// cross-page reads internally, so there is no need to
	// split into per-page requests.
	err := m.readInto(address, result)
	if err != nil {
		return nil, err
	}

	//----------------------------------------------------------------------------//

	return result, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) readInto(address uintptr, buf []byte) error {

	//----------------------------------------------------------------------------//

	// Lock for retrieval
	m.leech.lock.RLock()
	defer m.leech.lock.RUnlock()

	// Make sure that handle and PID are valid
	if m.leech.handle == 0 || m.proc.pid == 0 {
		return errors.New("process is not valid")
	}

	//----------------------------------------------------------------------------//

	var bytesRead uint32

	// Read directly into the provided buffer
	success, err := vmmCall(
		vmmDll.memReadEx,
		m.leech.handle,
		uintptr(m.proc.pid),
		address,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
		uintptr(unsafe.Pointer(&bytesRead)),
		uintptr(0x1|0x2), // VMMDLL_FLAG_NOCACHE | VMMDLL_FLAG_ZEROPAD_ON_FAIL
	)
	if err != nil {
		m.stats.ReadErrors++
		return errors.New(
			"failed to read memory",
			errors.Uint32("pid", m.proc.pid),
			errors.Uintptr("address", address),
			errors.Error("error", err),
		)
	}
	if success == 0 {
		m.stats.ReadErrors++
		return errors.New(
			"failed to read memory",
			errors.Uint32("pid", m.proc.pid),
			errors.Uintptr("address", address),
		)
	}

	//----------------------------------------------------------------------------//

	m.stats.SystemReads++
	return nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

func (m *Memory) writePage(address uintptr, data []byte) error {

	//----------------------------------------------------------------------------//

	// Lock for write
	m.leech.lock.RLock()
	defer m.leech.lock.RUnlock()

	// Make sure that handle and PID are valid
	if m.leech.handle == 0 || m.proc.pid == 0 {
		return errors.New("process is not valid")
	}

	//----------------------------------------------------------------------------//

	// Attempt to write the memory
	success, err := vmmCall(
		vmmDll.memWrite,
		m.leech.handle,
		uintptr(m.proc.pid),
		address,
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(len(data)),
	)
	if err != nil {
		m.stats.WriteErrors++
		return errors.New(
			"failed to write memory",
			errors.Uint32("pid", m.proc.pid),
			errors.Uintptr("address", address),
			errors.Error("error", err),
		)
	}
	if success == 0 {
		m.stats.WriteErrors++
		return errors.New(
			"failed to write memory",
			errors.Uint32("pid", m.proc.pid),
			errors.Uintptr("address", address),
		)
	}

	//----------------------------------------------------------------------------//

	m.stats.SystemWrites++
	return nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// ReadTypes reads `count` values of the given `memType` from
// the target `address`. Each value is `length` bytes and
// values are spaced by `stride` bytes. If `stride` is zero,
// it defaults to `length`. Returns a slice of `any` values
// that must be type-asserted by the caller. The typed methods
// (ReadInt32, ReadFloat64, etc.) are simpler alternatives for
// common types.
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

	// Check for overflow in total size calculation
	size := uint64(count)*uint64(stride) + uint64(length) - uint64(stride)
	if size > uint64(^uint32(0)) {
		return nil, errors.New(
			"read size overflow",
			errors.Uint32("count", count),
			errors.Uint32("stride", stride),
		)
	}
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

// ReadInt8 reads a signed 8-bit integer from the address.
func (m *Memory) ReadInt8(address uintptr) (int8, error) {

	data, err := m.ReadData(address, 1)
	if err != nil {
		return 0, err
	}

	return int8(data[0]), nil
}

////////////////////////////////////////////////////////////////////////////////

// ReadInt8s reads an array of signed 8-bit integers.
func (m *Memory) ReadInt8s(address uintptr, count, stride uint32) ([]int8, error) {

	res, err := m.ReadTypes(address, MemTypeInt8, 1, count, stride)
	if err != nil {
		return nil, err
	}

	out := make([]int8, len(res))
	for i, v := range res {
		out[i] = v.(int8)
	}

	return out, nil
}

////////////////////////////////////////////////////////////////////////////////

// ReadInt16 reads a signed 16-bit integer from the address.
func (m *Memory) ReadInt16(address uintptr) (int16, error) {

	data, err := m.ReadData(address, 2)
	if err != nil {
		return 0, err
	}

	return int16(binary.LittleEndian.Uint16(data)), nil
}

////////////////////////////////////////////////////////////////////////////////

// ReadInt16s reads an array of signed 16-bit integers.
func (m *Memory) ReadInt16s(address uintptr, count, stride uint32) ([]int16, error) {

	res, err := m.ReadTypes(address, MemTypeInt16, 2, count, stride)
	if err != nil {
		return nil, err
	}

	out := make([]int16, len(res))
	for i, v := range res {
		out[i] = v.(int16)
	}

	return out, nil
}

////////////////////////////////////////////////////////////////////////////////

// ReadInt32 reads a signed 32-bit integer from the address.
func (m *Memory) ReadInt32(address uintptr) (int32, error) {

	data, err := m.ReadData(address, 4)
	if err != nil {
		return 0, err
	}

	return int32(binary.LittleEndian.Uint32(data)), nil
}

////////////////////////////////////////////////////////////////////////////////

// ReadInt32s reads an array of signed 32-bit integers.
func (m *Memory) ReadInt32s(address uintptr, count, stride uint32) ([]int32, error) {

	res, err := m.ReadTypes(address, MemTypeInt32, 4, count, stride)
	if err != nil {
		return nil, err
	}

	out := make([]int32, len(res))
	for i, v := range res {
		out[i] = v.(int32)
	}

	return out, nil
}

////////////////////////////////////////////////////////////////////////////////

// ReadInt64 reads a signed 64-bit integer from the address.
func (m *Memory) ReadInt64(address uintptr) (int64, error) {

	data, err := m.ReadData(address, 8)
	if err != nil {
		return 0, err
	}

	return int64(binary.LittleEndian.Uint64(data)), nil
}

////////////////////////////////////////////////////////////////////////////////

// ReadInt64s reads an array of signed 64-bit integers.
func (m *Memory) ReadInt64s(address uintptr, count, stride uint32) ([]int64, error) {

	res, err := m.ReadTypes(address, MemTypeInt64, 8, count, stride)
	if err != nil {
		return nil, err
	}

	out := make([]int64, len(res))
	for i, v := range res {
		out[i] = v.(int64)
	}

	return out, nil
}

////////////////////////////////////////////////////////////////////////////////

// ReadFloat32 reads a 32-bit float from the address.
func (m *Memory) ReadFloat32(address uintptr) (float32, error) {

	data, err := m.ReadData(address, 4)
	if err != nil {
		return 0, err
	}

	return math.Float32frombits(binary.LittleEndian.Uint32(data)), nil
}

////////////////////////////////////////////////////////////////////////////////

// ReadFloat32s reads an array of 32-bit floats.
func (m *Memory) ReadFloat32s(address uintptr, count, stride uint32) ([]float32, error) {

	res, err := m.ReadTypes(address, MemTypeFloat32, 4, count, stride)
	if err != nil {
		return nil, err
	}

	out := make([]float32, len(res))
	for i, v := range res {
		out[i] = v.(float32)
	}

	return out, nil
}

////////////////////////////////////////////////////////////////////////////////

// ReadFloat64 reads a 64-bit float from the address.
func (m *Memory) ReadFloat64(address uintptr) (float64, error) {

	data, err := m.ReadData(address, 8)
	if err != nil {
		return 0, err
	}

	return math.Float64frombits(binary.LittleEndian.Uint64(data)), nil
}

////////////////////////////////////////////////////////////////////////////////

// ReadFloat64s reads an array of 64-bit floats.
func (m *Memory) ReadFloat64s(address uintptr, count, stride uint32) ([]float64, error) {

	res, err := m.ReadTypes(address, MemTypeFloat64, 8, count, stride)
	if err != nil {
		return nil, err
	}

	out := make([]float64, len(res))
	for i, v := range res {
		out[i] = v.(float64)
	}

	return out, nil
}

////////////////////////////////////////////////////////////////////////////////

// ReadPtr reads a pointer-sized value from the address. The
// size is determined by the target process bitness.
func (m *Memory) ReadPtr(address uintptr) (uintptr, error) {

	// If 64-bit process
	if m.proc.is64Bit {
		data, err := m.ReadData(address, 8)
		if err != nil {
			return 0, err
		}
		return uintptr(binary.LittleEndian.Uint64(data)), nil
	}

	data, err := m.ReadData(address, 4)
	if err != nil {
		return 0, err
	}

	return uintptr(binary.LittleEndian.Uint32(data)), nil
}

////////////////////////////////////////////////////////////////////////////////

// ReadPtrs reads an array of pointer-sized values.
func (m *Memory) ReadPtrs(address uintptr, count, stride uint32) ([]uintptr, error) {

	// If 64-bit process
	if m.proc.is64Bit {

		res, err := m.ReadTypes(address, MemTypeInt64, 8, count, stride)
		if err != nil {
			return nil, err
		}

		out := make([]uintptr, len(res))
		for i, v := range res {
			out[i] = uintptr(v.(int64))
		}

		return out, nil
	}

	res, err := m.ReadTypes(address, MemTypeInt32, 4, count, stride)
	if err != nil {
		return nil, err
	}

	out := make([]uintptr, len(res))
	for i, v := range res {
		out[i] = uintptr(v.(int32))
	}

	return out, nil
}

////////////////////////////////////////////////////////////////////////////////

// ReadBool reads a boolean value (1 byte) from the address.
func (m *Memory) ReadBool(address uintptr) (bool, error) {

	data, err := m.ReadData(address, 1)
	if err != nil {
		return false, err
	}

	return data[0] != 0, nil
}

////////////////////////////////////////////////////////////////////////////////

// ReadBools reads an array of boolean values.
func (m *Memory) ReadBools(address uintptr, count, stride uint32) ([]bool, error) {

	res, err := m.ReadTypes(address, MemTypeBool, 1, count, stride)
	if err != nil {
		return nil, err
	}

	out := make([]bool, len(res))
	for i, v := range res {
		out[i] = v.(bool)
	}

	return out, nil
}

////////////////////////////////////////////////////////////////////////////////

// ReadString reads a null-terminated C string of the given
// maximum length from the address.
func (m *Memory) ReadString(address uintptr, length uint32) (string, error) {

	data, err := m.ReadData(address, uintptr(length))
	if err != nil {
		return "", err
	}

	return cStrToString(data), nil
}

////////////////////////////////////////////////////////////////////////////////

// ReadStrings reads an array of null-terminated C strings.
func (m *Memory) ReadStrings(address uintptr, length, count, stride uint32) ([]string, error) {

	res, err := m.ReadTypes(address, MemTypeString, length, count, stride)
	if err != nil {
		return nil, err
	}

	out := make([]string, len(res))
	for i, v := range res {
		out[i] = v.(string)
	}

	return out, nil
}

////////////////////////////////////////////////////////////////////////////////

// WriteInt8 writes a signed 8-bit integer to the address.
func (m *Memory) WriteInt8(address uintptr, value int8) error {

	return m.WriteData(address, []byte{byte(value)})
}

////////////////////////////////////////////////////////////////////////////////

// WriteInt16 writes a signed 16-bit integer to the address.
func (m *Memory) WriteInt16(address uintptr, value int16) error {

	data := make([]byte, 2)
	binary.LittleEndian.PutUint16(data, uint16(value))
	return m.WriteData(address, data)
}

////////////////////////////////////////////////////////////////////////////////

// WriteInt32 writes a signed 32-bit integer to the address.
func (m *Memory) WriteInt32(address uintptr, value int32) error {

	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, uint32(value))
	return m.WriteData(address, data)
}

////////////////////////////////////////////////////////////////////////////////

// WriteInt64 writes a signed 64-bit integer to the address.
func (m *Memory) WriteInt64(address uintptr, value int64) error {

	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, uint64(value))
	return m.WriteData(address, data)
}

////////////////////////////////////////////////////////////////////////////////

// WriteFloat32 writes a 32-bit float to the address.
func (m *Memory) WriteFloat32(address uintptr, value float32) error {

	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, math.Float32bits(value))
	return m.WriteData(address, data)
}

////////////////////////////////////////////////////////////////////////////////

// WriteFloat64 writes a 64-bit float to the address.
func (m *Memory) WriteFloat64(address uintptr, value float64) error {

	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, math.Float64bits(value))
	return m.WriteData(address, data)
}

////////////////////////////////////////////////////////////////////////////////

// WritePtr writes a pointer-sized value to the address. The
// size is determined by the target process bitness.
func (m *Memory) WritePtr(address uintptr, value uintptr) error {

	// If 64-bit process
	if m.proc.is64Bit {
		return m.WriteInt64(address, int64(value))
	}

	return m.WriteInt32(address, int32(value))
}

////////////////////////////////////////////////////////////////////////////////

// WriteBool writes a boolean value (1 byte) to the address.
func (m *Memory) WriteBool(address uintptr, value bool) error {

	b := byte(0)
	if value {
		b = 1
	}

	return m.WriteData(address, []byte{b})
}
