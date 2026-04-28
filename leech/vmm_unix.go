//go:build darwin || linux

package leech

import (
	"runtime"
	"unsafe"

	"github.com/ebitengine/purego"

	"github.com/dkrutsko/oasis/errors"
)

////////////////////////////////////////////////////////////////////////////////

type vmmProc = uintptr

////////////////////////////////////////////////////////////////////////////////

type vmmLib struct {
	lib uintptr

	initializeEx vmmProc
	close        vmmProc
	memSize      vmmProc
	memFree      vmmProc

	processGetInformation       vmmProc
	processGetInformationAll    vmmProc
	processGetInformationString vmmProc

	mapGetModule vmmProc
	mapGetVad    vmmProc
	mapGetPte    vmmProc
	memReadEx    vmmProc
	memWrite     vmmProc

	configGet vmmProc
	configSet vmmProc
}

////////////////////////////////////////////////////////////////////////////////

func loadVmmDll() error {

	//----------------------------------------------------------------------------//

	// If already loaded
	// Without using lock
	if vmmDll != nil {
		return nil
	}

	// Lock for load
	vmmDllLock.Lock()
	defer vmmDllLock.Unlock()

	// If already loaded
	if vmmDll != nil {
		return nil
	}

	result := &vmmLib{}

	//----------------------------------------------------------------------------//

	// Select library name based on platform
	libName := "vmm.so"
	if runtime.GOOS == "darwin" {
		libName = "vmm.dylib"
	}

	var err error
	// Attempt to load the VMM shared library
	result.lib, err = purego.Dlopen(libName, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return errors.New(
			"failed to load vmm library",
			errors.Error("error", err),
		)
	}

	//----------------------------------------------------------------------------//

	// Helper to find a symbol and wrap the error. If any
	// lookup fails, the library is closed before returning.
	find := func(name string) (uintptr, error) {
		sym, err := purego.Dlsym(result.lib, name)
		if err != nil {
			purego.Dlclose(result.lib)
			return 0, errors.New(
				"failed to find proc in vmm library",
				errors.String("proc", name),
				errors.Error("error", err),
			)
		}
		return sym, nil
	}

	//----------------------------------------------------------------------------//

	result.initializeEx, err = find("VMMDLL_InitializeEx")
	if err != nil {
		return err
	}

	result.close, err = find("VMMDLL_Close")
	if err != nil {
		return err
	}

	result.memSize, err = find("VMMDLL_MemSize")
	if err != nil {
		return err
	}

	result.memFree, err = find("VMMDLL_MemFree")
	if err != nil {
		return err
	}

	//----------------------------------------------------------------------------//

	result.processGetInformation, err = find("VMMDLL_ProcessGetInformation")
	if err != nil {
		return err
	}

	result.processGetInformationAll, err = find("VMMDLL_ProcessGetInformationAll")
	if err != nil {
		return err
	}

	result.processGetInformationString, err = find("VMMDLL_ProcessGetInformationString")
	if err != nil {
		return err
	}

	//----------------------------------------------------------------------------//

	result.mapGetModule, err = find("VMMDLL_Map_GetModuleU")
	if err != nil {
		return err
	}

	result.mapGetVad, err = find("VMMDLL_Map_GetVadU")
	if err != nil {
		return err
	}

	result.mapGetPte, err = find("VMMDLL_Map_GetPteU")
	if err != nil {
		return err
	}

	result.memReadEx, err = find("VMMDLL_MemReadEx")
	if err != nil {
		return err
	}

	result.memWrite, err = find("VMMDLL_MemWrite")
	if err != nil {
		return err
	}

	//----------------------------------------------------------------------------//

	result.configGet, err = find("VMMDLL_ConfigGet")
	if err != nil {
		return err
	}

	result.configSet, err = find("VMMDLL_ConfigSet")
	if err != nil {
		return err
	}

	//----------------------------------------------------------------------------//

	vmmDll = result
	return nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

func vmmCall(proc vmmProc, args ...uintptr) (uintptr, error) {

	r1, _, _ := purego.SyscallN(proc, args...)
	return r1, nil
}

////////////////////////////////////////////////////////////////////////////////

func ptrToString(ptr uintptr) string {

	if ptr == 0 {
		return ""
	}

	// Read null-terminated UTF-8 string from pointer
	n := 0
	for *(*byte)(unsafe.Pointer(ptr + uintptr(n))) != 0 {
		n++
	}

	if n == 0 {
		return ""
	}

	return string(unsafe.Slice((*byte)(unsafe.Pointer(ptr)), n))
}
