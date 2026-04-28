package leech

// VMMDLL core configuration options. Use with `Leech.GetConfig`
// and `Leech.SetConfig`. Read-only options are marked (R),
// read-write options are marked (RW), and write-only options
// are marked (W).
const (
	ConfigCorePrintfEnable  uint64 = 0x4000000100000000 // RW
	ConfigCoreVerbose       uint64 = 0x4000000200000000 // RW
	ConfigCoreVerboseExtra  uint64 = 0x4000000300000000 // RW
	ConfigCoreMaxNativeAddr uint64 = 0x4000000800000000 // R
	ConfigCoreSystem        uint64 = 0x2000000100000000 // R
	ConfigCoreMemoryModel   uint64 = 0x2000000200000000 // R
)

// VMMDLL cache and timing options.
const (
	ConfigTickPeriod       uint64 = 0x2000000400000000 // RW
	ConfigReadCacheTicks   uint64 = 0x2000000500000000 // RW
	ConfigTlbCacheTicks    uint64 = 0x2000000600000000 // RW
	ConfigProcCachePartial uint64 = 0x2000000700000000 // RW
	ConfigProcCacheTotal   uint64 = 0x2000000800000000 // RW
	ConfigPagingEnabled    uint64 = 0x2000000D00000000 // RW
)

// VMMDLL version information options.
const (
	ConfigVmmVersionMajor    uint64 = 0x2000000900000000 // R
	ConfigVmmVersionMinor    uint64 = 0x2000000A00000000 // R
	ConfigVmmVersionRevision uint64 = 0x2000000B00000000 // R
	ConfigWinVersionMajor    uint64 = 0x2000010100000000 // R
	ConfigWinVersionMinor    uint64 = 0x2000010200000000 // R
	ConfigWinVersionBuild    uint64 = 0x2000010300000000 // R
)

// LeechCore FPGA device options. These are LC_OPT_FPGA_*
// constants forwarded to LeechCore via `Leech.SetConfig`.
const (
	ConfigFpgaMaxSizeRx   uint64 = 0x0300000300000000 // RW
	ConfigFpgaMaxSizeTx   uint64 = 0x0300000400000000 // RW
	ConfigFpgaDelayRead   uint64 = 0x0300000800000000 // RW
	ConfigFpgaDelayWrite  uint64 = 0x0300000700000000 // RW
	ConfigFpgaRetryOnError uint64 = 0x0300000900000000 // RW
	ConfigFpgaAlgoTiny    uint64 = 0x0300008400000000 // RW
	ConfigFpgaDeviceId    uint64 = 0x0300008000000000 // RW
	ConfigFpgaFpgaId      uint64 = 0x0300008100000000 // R
	ConfigFpgaVersionMajor uint64 = 0x0300008200000000 // R
	ConfigFpgaVersionMinor uint64 = 0x0300008300000000 // R
)

// VMMDLL refresh options. Writing any value to these options
// triggers the corresponding cache refresh.
const (
	ConfigRefreshAll        uint64 = 0x2001ffff00000000 // W
	ConfigRefreshFreqMem    uint64 = 0x2001100000000000 // W
	ConfigRefreshFreqTlb    uint64 = 0x2001080000000000 // W
	ConfigRefreshFreqFast   uint64 = 0x2001040000000000 // W
	ConfigRefreshFreqMedium uint64 = 0x2001000100000000 // W
	ConfigRefreshFreqSlow   uint64 = 0x2001001000000000 // W
)

/*

TODO:

Maybe these should not be enums, but individual functions like `GetConfigPrintfEnable`, etc.
VMMDLL_ConfigGet
VMMDLL_ConfigSet




// Options used together with the functions: VMMDLL_ConfigGet & VMMDLL_ConfigSet
// Options are defined with either: VMMDLL_OPT_* in this header file or as
// LC_OPT_* in leechcore.h
// For more detailed information check the sources for individual device types.
#define VMMDLL_OPT_CORE_PRINTF_ENABLE                   0x4000000100000000  // RW
#define VMMDLL_OPT_CORE_VERBOSE                         0x4000000200000000  // RW
#define VMMDLL_OPT_CORE_VERBOSE_EXTRA                   0x4000000300000000  // RW
#define VMMDLL_OPT_CORE_VERBOSE_EXTRA_TLP               0x4000000400000000  // RW
#define VMMDLL_OPT_CORE_MAX_NATIVE_ADDRESS              0x4000000800000000  // R
#define VMMDLL_OPT_CORE_LEECHCORE_HANDLE                0x4000001000000000  // R - underlying leechcore handle (do not close).
#define VMMDLL_OPT_CORE_VMM_ID                          0x4000002000000000  // R - use with startup option '-create-from-vmmid' to create a thread-safe duplicate VMM instance.

#define VMMDLL_OPT_CORE_SYSTEM                          0x2000000100000000  // R
#define VMMDLL_OPT_CORE_MEMORYMODEL                     0x2000000200000000  // R

#define VMMDLL_OPT_CONFIG_IS_REFRESH_ENABLED            0x2000000300000000  // R - 1/0
#define VMMDLL_OPT_CONFIG_TICK_PERIOD                   0x2000000400000000  // RW - base tick period in ms
#define VMMDLL_OPT_CONFIG_READCACHE_TICKS               0x2000000500000000  // RW - memory cache validity period (in ticks)
#define VMMDLL_OPT_CONFIG_TLBCACHE_TICKS                0x2000000600000000  // RW - page table (tlb) cache validity period (in ticks)
#define VMMDLL_OPT_CONFIG_PROCCACHE_TICKS_PARTIAL       0x2000000700000000  // RW - process refresh (partial) period (in ticks)
#define VMMDLL_OPT_CONFIG_PROCCACHE_TICKS_TOTAL         0x2000000800000000  // RW - process refresh (full) period (in ticks)
#define VMMDLL_OPT_CONFIG_VMM_VERSION_MAJOR             0x2000000900000000  // R
#define VMMDLL_OPT_CONFIG_VMM_VERSION_MINOR             0x2000000A00000000  // R
#define VMMDLL_OPT_CONFIG_VMM_VERSION_REVISION          0x2000000B00000000  // R
#define VMMDLL_OPT_CONFIG_STATISTICS_FUNCTIONCALL       0x2000000C00000000  // RW - enable function call statistics (.status/statistics_fncall file)
#define VMMDLL_OPT_CONFIG_IS_PAGING_ENABLED             0x2000000D00000000  // RW - 1/0
#define VMMDLL_OPT_CONFIG_DEBUG                         0x2000000E00000000  // W
#define VMMDLL_OPT_CONFIG_YARA_RULES                    0x2000000F00000000  // R

#define VMMDLL_OPT_WIN_VERSION_MAJOR                    0x2000010100000000  // R
#define VMMDLL_OPT_WIN_VERSION_MINOR                    0x2000010200000000  // R
#define VMMDLL_OPT_WIN_VERSION_BUILD                    0x2000010300000000  // R
#define VMMDLL_OPT_WIN_SYSTEM_UNIQUE_ID                 0x2000010400000000  // R

#define VMMDLL_OPT_FORENSIC_MODE                        0x2000020100000000  // RW - enable/retrieve forensic mode type [0-4].

// REFRESH OPTIONS:
#define VMMDLL_OPT_REFRESH_ALL                          0x2001ffff00000000  // W - refresh all caches
#define VMMDLL_OPT_REFRESH_FREQ_MEM                     0x2001100000000000  // W - refresh memory cache (excl. TLB) [fully]
#define VMMDLL_OPT_REFRESH_FREQ_MEM_PARTIAL             0x2001000200000000  // W - refresh memory cache (excl. TLB) [partial 33%/call]
#define VMMDLL_OPT_REFRESH_FREQ_TLB                     0x2001080000000000  // W - refresh page table (TLB) cache [fully]
#define VMMDLL_OPT_REFRESH_FREQ_TLB_PARTIAL             0x2001000400000000  // W - refresh page table (TLB) cache [partial 33%/call]
#define VMMDLL_OPT_REFRESH_FREQ_FAST                    0x2001040000000000  // W - refresh fast frequency - incl. partial process refresh
#define VMMDLL_OPT_REFRESH_FREQ_MEDIUM                  0x2001000100000000  // W - refresh medium frequency - incl. full process refresh
#define VMMDLL_OPT_REFRESH_FREQ_SLOW                    0x2001001000000000  // W - refresh slow frequency.

// PROCESS OPTIONS: [LO-DWORD: Process PID]
#define VMMDLL_OPT_PROCESS_DTB                          0x2002000100000000  // W - force set process directory table base.
#define VMMDLL_OPT_PROCESS_DTB_FAST_LOWINTEGRITY        0x2002000200000000  // W - force set process directory table base (fast, low integrity mode, with less checks) - use at own risk!.

static LPCSTR VMMDLL_MEMORYMODEL_TOSTRING[5] = { "N/A", "X86", "X86PAE", "X64", "ARM64" };

typedef enum tdVMMDLL_MEMORYMODEL_TP {
    VMMDLL_MEMORYMODEL_NA       = 0,
    VMMDLL_MEMORYMODEL_X86      = 1,
    VMMDLL_MEMORYMODEL_X86PAE   = 2,
    VMMDLL_MEMORYMODEL_X64      = 3,
    VMMDLL_MEMORYMODEL_ARM64    = 4,
} VMMDLL_MEMORYMODEL_TP;

typedef enum tdVMMDLL_SYSTEM_TP {
    VMMDLL_SYSTEM_UNKNOWN_PHYSICAL = 0,
    VMMDLL_SYSTEM_UNKNOWN_64    = 1,
    VMMDLL_SYSTEM_WINDOWS_64    = 2,
    VMMDLL_SYSTEM_UNKNOWN_32    = 3,
    VMMDLL_SYSTEM_WINDOWS_32    = 4,
    VMMDLL_SYSTEM_UNKNOWN_X64   = 1,    // deprecated - do not use!
    VMMDLL_SYSTEM_WINDOWS_X64   = 2,    // deprecated - do not use!
    VMMDLL_SYSTEM_UNKNOWN_X86   = 3,    // deprecated - do not use!
    VMMDLL_SYSTEM_WINDOWS_X86   = 4     // deprecated - do not use!
} VMMDLL_SYSTEM_TP;

// Get a device specific option value. Please see defines VMMDLL_OPT_* for infor-
// mation about valid option values. Please note that option values may overlap
// between different device types with different meanings.
// -- hVMM
// -- fOption
// -- pqwValue = pointer to ULONG64 to receive option value.
// -- return = success/fail.
EXPORTED_FUNCTION _Success_(return)
BOOL VMMDLL_ConfigGet(_In_ VMM_HANDLE hVMM, _In_ ULONG64 fOption, _Out_ PULONG64 pqwValue);

// Set a device specific option value. Please see defines VMMDLL_OPT_* for infor-
// mation about valid option values. Please note that option values may overlap
// between different device types with different meanings.
// -- hVMM
// -- fOption
// -- qwValue
// -- return = success/fail.
EXPORTED_FUNCTION _Success_(return)
BOOL VMMDLL_ConfigSet(_In_ VMM_HANDLE hVMM, _In_ ULONG64 fOption, _In_ ULONG64 qwValue);




*/
