package leech

import (
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/dkrutsko/oasis/errors"
)

////////////////////////////////////////////////////////////////////////////////

type vmmProc = *windows.Proc

////////////////////////////////////////////////////////////////////////////////

type vmmLib struct {
	dll *windows.DLL

	initializeEx *windows.Proc
	close        *windows.Proc
	memSize      *windows.Proc
	memFree      *windows.Proc

	processGetInformation       *windows.Proc
	processGetInformationAll    *windows.Proc
	processGetInformationString *windows.Proc

	mapGetModule *windows.Proc
	mapGetVad    *windows.Proc
	mapGetPte    *windows.Proc
	memReadEx    *windows.Proc
	memWrite     *windows.Proc

	configGet *windows.Proc
	configSet *windows.Proc
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

	var err error
	// Attempt to load the VMM dynamic library
	result.dll, err = windows.LoadDLL("vmm.dll")
	if err != nil {
		return errors.New(
			"failed to load vmm library",
			errors.Error("error", err),
		)
	}

	//----------------------------------------------------------------------------//

	// Helper to find a proc and wrap the error. If any
	// lookup fails, the DLL is released before returning.
	find := func(name string) (*windows.Proc, error) {
		proc, err := result.dll.FindProc(name)
		if err != nil {
			result.dll.Release()
			return nil, errors.New(
				"failed to find proc in vmm library",
				errors.String("proc", name),
				errors.Error("error", err),
			)
		}
		return proc, nil
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

	result.mapGetModule, err = find("VMMDLL_Map_GetModuleW")
	if err != nil {
		return err
	}

	result.mapGetVad, err = find("VMMDLL_Map_GetVadW")
	if err != nil {
		return err
	}

	result.mapGetPte, err = find("VMMDLL_Map_GetPteW")
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

	r1, _, err := proc.Call(args...)
	if err.(windows.Errno) != 0 {
		return r1, err
	}

	return r1, nil
}

////////////////////////////////////////////////////////////////////////////////

func ptrToString(ptr uintptr) string {

	if ptr == 0 {
		return ""
	}

	// Convert wide (UTF-16) string to Go string
	return windows.UTF16PtrToString(
		(*uint16)(unsafe.Pointer(ptr)),
	)
}
