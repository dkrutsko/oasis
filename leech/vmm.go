package leech

import (
	"strconv"
	"strings"
	"sync"
)

//----------------------------------------------------------------------------//
// Library                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// VmmRelease is the VMMDLL version this package was built
// against.
const VmmRelease = "5.14.5"

////////////////////////////////////////////////////////////////////////////////

var (
	vmmDll     *vmmLib
	vmmDllLock sync.Mutex
)

//----------------------------------------------------------------------------//
// Helpers                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

func cStrToString(b []byte) string {

	// Find null-terminator
	n := len(b)
	for i, c := range b {
		if c == 0 {
			n = i
			break
		}
	}

	// Convert string up to null-terminator
	return strings.TrimSpace(string(b[:n]))
}

//----------------------------------------------------------------------------//
// Process Information                                                        //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

const (
	vmmProcessInformationMagic   uint64 = 0xC0FFEE663DF9301E
	vmmProcessInformationVersion uint16 = 7
)

////////////////////////////////////////////////////////////////////////////////

type vmmMemoryModel uint32

const (
	vmmMemoryModelNA     vmmMemoryModel = 0
	vmmMemoryModelX86    vmmMemoryModel = 1
	vmmMemoryModelX86PAE vmmMemoryModel = 2
	vmmMemoryModelX64    vmmMemoryModel = 3
	vmmMemoryModelARM64  vmmMemoryModel = 4
)

////////////////////////////////////////////////////////////////////////////////

func (m vmmMemoryModel) String() string {

	switch m {
	case vmmMemoryModelNA:
		return "NA"

	case vmmMemoryModelX86:
		return "x86"

	case vmmMemoryModelX86PAE:
		return "x86PAE"

	case vmmMemoryModelX64:
		return "x64"

	case vmmMemoryModelARM64:
		return "arm64"
	}

	return strconv.FormatUint(uint64(m), 10)
}

////////////////////////////////////////////////////////////////////////////////

type vmmSystemType uint32

const (
	vmmSystemTypeUnknownPhysical vmmSystemType = 0
	vmmSystemTypeUnknown64       vmmSystemType = 1
	vmmSystemTypeWindows64       vmmSystemType = 2
	vmmSystemTypeUnknown32       vmmSystemType = 3
	vmmSystemTypeWindows32       vmmSystemType = 4
)

////////////////////////////////////////////////////////////////////////////////

func (s vmmSystemType) String() string {

	switch s {
	case vmmSystemTypeUnknownPhysical:
		return "UnknownPhysical"

	case vmmSystemTypeUnknown64:
		return "Unknown64"

	case vmmSystemTypeWindows64:
		return "Windows64"

	case vmmSystemTypeUnknown32:
		return "Unknown32"

	case vmmSystemTypeWindows32:
		return "Windows32"
	}

	return strconv.FormatUint(uint64(s), 10)
}

////////////////////////////////////////////////////////////////////////////////

type vmmProcessIntegrityLevel uint32

const (
	vmmProcessIntegrityLevelUnknown    vmmProcessIntegrityLevel = 0
	vmmProcessIntegrityLevelUntrusted  vmmProcessIntegrityLevel = 1
	vmmProcessIntegrityLevelLow        vmmProcessIntegrityLevel = 2
	vmmProcessIntegrityLevelMedium     vmmProcessIntegrityLevel = 3
	vmmProcessIntegrityLevelMediumPlus vmmProcessIntegrityLevel = 4
	vmmProcessIntegrityLevelHigh       vmmProcessIntegrityLevel = 5
	vmmProcessIntegrityLevelSystem     vmmProcessIntegrityLevel = 6
	vmmProcessIntegrityLevelProtected  vmmProcessIntegrityLevel = 7
)

////////////////////////////////////////////////////////////////////////////////

func (l vmmProcessIntegrityLevel) String() string {

	switch l {
	case vmmProcessIntegrityLevelUnknown:
		return "Unknown"

	case vmmProcessIntegrityLevelUntrusted:
		return "Untrusted"

	case vmmProcessIntegrityLevelLow:
		return "Low"

	case vmmProcessIntegrityLevelMedium:
		return "Medium"

	case vmmProcessIntegrityLevelMediumPlus:
		return "MediumPlus"

	case vmmProcessIntegrityLevelHigh:
		return "High"

	case vmmProcessIntegrityLevelSystem:
		return "System"

	case vmmProcessIntegrityLevelProtected:
		return "Protected"
	}

	return strconv.FormatUint(uint64(l), 10)
}

////////////////////////////////////////////////////////////////////////////////

type vmmProcessInformation struct {
	magic       uint64
	version     uint16
	size        uint16
	memoryModel vmmMemoryModel
	system      vmmSystemType
	userOnly    uint32
	pid         uint32
	ppid        uint32
	state       uint32
	name        [16]byte
	nameLong    [64]byte
	dtb         uint64
	dtbUserOpt  uint64

	win struct {
		eprocess       uint64
		peb            uint64
		reserved       uint64
		wow64          uint32
		peb32          uint32
		sessionId      uint32
		luid           uint64
		sid            [260]byte
		integrityLevel vmmProcessIntegrityLevel
	}
}

//----------------------------------------------------------------------------//
// Map Module                                                                 //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

const (
	vmmMapModuleVersion uint32 = 6
)

////////////////////////////////////////////////////////////////////////////////

type vmmModuleType uint32

const (
	vmmModuleTypeNormal    vmmModuleType = 0
	vmmModuleTypeData      vmmModuleType = 1
	vmmModuleTypeNotLinked vmmModuleType = 2
	vmmModuleTypeInjected  vmmModuleType = 3
)

////////////////////////////////////////////////////////////////////////////////

func (m vmmModuleType) String() string {

	switch m {
	case vmmModuleTypeNormal:
		return "Normal"

	case vmmModuleTypeData:
		return "Data"

	case vmmModuleTypeNotLinked:
		return "NotLinked"

	case vmmModuleTypeInjected:
		return "Injected"
	}

	return strconv.FormatUint(uint64(m), 10)
}

////////////////////////////////////////////////////////////////////////////////

type vmmMapModuleEntry struct {
	base         uint64
	entry        uint64
	size         uint32
	wow64        uint32
	name         uintptr
	reserved1    uint32
	reserved2    uint32
	path         uintptr
	moduleType   vmmModuleType
	fileSize     uint32
	sectionCount uint32
	eat          uint32
	iat          uint32
	reserved3    uint32
	reserved4    [3]uint64
	debugInfo    uintptr
	versionInfo  uintptr
}

////////////////////////////////////////////////////////////////////////////////

type vmmMapModule struct {
	version       uint32
	reserved      [5]uint32
	multiText     uintptr
	multiTextSize uint32
	count         uint32
	modules       uintptr
}

//----------------------------------------------------------------------------//
// Map VAD                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

const (
	vmmMapVadVersion uint32 = 6
)

////////////////////////////////////////////////////////////////////////////////

// Windows page protection flags matching the values in
// `Region.Access`. The AccessPage* constants are composite
// masks for testing readable, writable, and executable
// permissions.
const (
	PageNoAccess         = 0x01
	PageReadOnly         = 0x02
	PageReadWrite        = 0x04
	PageWriteCopy        = 0x08
	PageExecute          = 0x10
	PageExecuteRead      = 0x20
	PageExecuteReadWrite = 0x40
	PageExecuteWriteCopy = 0x80
	PageGuard            = 0x100
	PageNoCache          = 0x200
	PageWriteCombine     = 0x400

	AccessPageR = PageReadOnly | PageReadWrite | PageWriteCopy | PageExecute | PageExecuteRead | PageExecuteReadWrite | PageExecuteWriteCopy
	AccessPageW = PageReadWrite | PageWriteCopy | PageExecuteReadWrite | PageExecuteWriteCopy
	AccessPageX = PageExecute | PageExecuteRead | PageExecuteReadWrite | PageExecuteWriteCopy
)

////////////////////////////////////////////////////////////////////////////////

type vmmMapVadEntryFlags0Info struct {
	vadType         uint32 // Pos 0
	protection      uint32 // Pos 3
	isImage         bool   // Pos 8
	isFile          bool   // Pos 9
	isPageFile      bool   // Pos 10
	isPrivateMemory bool   // Pos 11
	isTeb           bool   // Pos 12
	isStack         bool   // Pos 13
	spare           uint32 // Pos 14
	heapNum         uint32 // Pos 16
	isHeap          bool   // Pos 23
	descriptionSize uint32 // Pos 24
}

////////////////////////////////////////////////////////////////////////////////

type vmmMapVadEntryFlags0 uint32

////////////////////////////////////////////////////////////////////////////////

func (f vmmMapVadEntryFlags0) Decode() vmmMapVadEntryFlags0Info {

	flags := uint32(f)
	// Convert flags into structure
	return vmmMapVadEntryFlags0Info{
		vadType:         ((flags >> 0) & 0x07),
		protection:      ((flags >> 3) & 0x1F),
		isImage:         ((flags >> 8) & 0x01) != 0,
		isFile:          ((flags >> 9) & 0x01) != 0,
		isPageFile:      ((flags >> 10) & 0x01) != 0,
		isPrivateMemory: ((flags >> 11) & 0x01) != 0,
		isTeb:           ((flags >> 12) & 0x01) != 0,
		isStack:         ((flags >> 13) & 0x01) != 0,
		spare:           ((flags >> 14) & 0x03),
		heapNum:         ((flags >> 16) & 0x7F),
		isHeap:          ((flags >> 23) & 0x01) != 0,
		descriptionSize: ((flags >> 24) & 0xFF),
	}
}

////////////////////////////////////////////////////////////////////////////////

type vmmMapVadEntryFlags1Info struct {
	commitCharge uint32 // Pos 0
	isCommitted  bool   // Pos 31
}

////////////////////////////////////////////////////////////////////////////////

type vmmMapVadEntryFlags1 uint32

////////////////////////////////////////////////////////////////////////////////

func (f vmmMapVadEntryFlags1) Decode() vmmMapVadEntryFlags1Info {

	flags := uint32(f)
	// Convert flags into structure
	return vmmMapVadEntryFlags1Info{
		commitCharge: ((flags >> 0) & 0x7FFFFFFF),
		isCommitted:  ((flags >> 31) & 0x00000001) != 0,
	}
}

////////////////////////////////////////////////////////////////////////////////

type vmmMapVadEntry struct {
	start            uint64
	end              uint64
	vad              uint64
	flags0           vmmMapVadEntryFlags0
	flags1           vmmMapVadEntryFlags1
	u2               uint32
	prototypePteSize uint32
	prototypePteAddr uint64
	subsection       uint64
	text             uintptr
	reserved1        uint32
	reserved2        uint32
	fileObject       uint64
	vadExPages       uint32
	vadExPagesBase   uint32
	reserved3        uint64
}

////////////////////////////////////////////////////////////////////////////////

type vmmMapVad struct {
	version       uint32
	reserved      [4]uint32
	pageCount     uint32
	multiText     uintptr
	multiTextSize uint32
	count         uint32
	vads          uintptr
}

//----------------------------------------------------------------------------//
// Map PTE                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

const (
	vmmMapPteVersion uint32 = 2
)

////////////////////////////////////////////////////////////////////////////////

type vmmMapPteEntryPageFlagsInfo struct {
	isWritable  bool // PAGE_W
	isNonSecure bool // PAGE_NS
	isNoExecute bool // PAGE_NX
}

////////////////////////////////////////////////////////////////////////////////

type vmmMapPteEntryPageFlags uint64

////////////////////////////////////////////////////////////////////////////////

func (f vmmMapPteEntryPageFlags) Decode() vmmMapPteEntryPageFlagsInfo {

	flags := uint64(f)
	// Convert flags into the structure
	return vmmMapPteEntryPageFlagsInfo{
		isWritable:  (flags & 0x0000000000000002) != 0,
		isNonSecure: (flags & 0x0000000000000004) != 0,
		isNoExecute: (flags & 0x8000000000000000) != 0,
	}
}

////////////////////////////////////////////////////////////////////////////////

type vmmMapPteEntry struct {
	base          uint64
	pageCount     uint64
	pageFlags     vmmMapPteEntryPageFlags
	isWoW64       uint32
	reserved1     uint32
	text          uintptr
	reserved2     uint32
	softwareCount uint32
}

////////////////////////////////////////////////////////////////////////////////

type vmmMapPte struct {
	version       uint32
	reserved      [5]uint32
	multiText     uintptr
	multiTextSize uint32
	count         uint32
	ptes          uintptr
}
