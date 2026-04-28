package shmem

import (
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/dkrutsko/oasis/errors"
)

////////////////////////////////////////////////////////////////////////////////

type windowsHandle struct {
	buf  []byte
	addr uintptr

	mapping windows.Handle
	file    windows.Handle

	path  string
	owner bool
}

////////////////////////////////////////////////////////////////////////////////

func (h *windowsHandle) data() []byte {

	return h.buf
}

////////////////////////////////////////////////////////////////////////////////

func (h *windowsHandle) size() int {

	return len(h.buf)
}

////////////////////////////////////////////////////////////////////////////////

func (h *windowsHandle) flush() error {

	// Only meaningful for file-backed segments
	if h.addr != 0 && h.path != "" {
		err := windows.FlushViewOfFile(h.addr, 0)
		if err != nil {
			return errors.New(
				"failed to flush shared memory",
				errors.Error("error", err),
			)
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

func (h *windowsHandle) close() error {

	if h.addr != 0 {
		windows.UnmapViewOfFile(h.addr)
		h.addr = 0
		h.buf = nil
	}

	if h.mapping != 0 && h.mapping != windows.InvalidHandle {
		windows.CloseHandle(h.mapping)
		h.mapping = windows.InvalidHandle
	}

	if h.file != 0 && h.file != windows.InvalidHandle {
		windows.CloseHandle(h.file)
		h.file = windows.InvalidHandle
	}

	if h.owner && h.path != "" {
		os.Remove(h.path)
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

func platformCreate(s *Segment) error {

	switch s.opts.Technique {

	case TechniqueAuto:
		err := createSharedMemory(s)
		if err == nil {
			return nil
		}
		return createFileBacked(s)

	case TechniqueSharedMemory:
		return createSharedMemory(s)

	case TechniqueFileBacked:
		return createFileBacked(s)

	default:
		return ErrUnsupportedTechnique
	}
}

////////////////////////////////////////////////////////////////////////////////

func platformOpen(s *Segment, readOnly bool) error {

	switch s.opts.Technique {

	case TechniqueAuto:
		err := openSharedMemory(s, readOnly)
		if err == nil {
			return nil
		}
		return openFileBacked(s, readOnly)

	case TechniqueSharedMemory:
		return openSharedMemory(s, readOnly)

	case TechniqueFileBacked:
		return openFileBacked(s, readOnly)

	default:
		return ErrUnsupportedTechnique
	}
}

////////////////////////////////////////////////////////////////////////////////

func createSharedMemory(s *Segment) error {

	//----------------------------------------------------------------------------//

	opts := s.opts

	namePtr, err := windows.UTF16PtrFromString(opts.Name)
	if err != nil {
		return errors.New(
			"failed to encode segment name",
			errors.Error("error", err),
		)
	}

	// Create a pagefile-backed mapping
	mapping, err := windows.CreateFileMapping(
		windows.InvalidHandle,
		nil,
		windows.PAGE_READWRITE,
		uint32(int64(opts.Size)>>32),
		uint32(opts.Size),
		namePtr,
	)

	if err != nil {
		return errors.New(
			"failed to create file mapping",
			errors.Error("error", err),
		)
	}

	// Map into process address space
	addr, err := windows.MapViewOfFile(
		mapping,
		windows.FILE_MAP_WRITE,
		0, 0, 0,
	)

	if err != nil {
		windows.CloseHandle(mapping)
		return errors.New(
			"failed to map view",
			errors.Error("error", err),
		)
	}

	s.h = &windowsHandle{
		buf:  sliceFromAddr(addr, opts.Size),
		addr: addr,

		mapping: mapping,
		file:    windows.InvalidHandle,

		owner: true,
	}

	s.technique = TechniqueSharedMemory
	return nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

func openSharedMemory(s *Segment, readOnly bool) error {

	//----------------------------------------------------------------------------//

	opts := s.opts

	namePtr, err := windows.UTF16PtrFromString(opts.Name)
	if err != nil {
		return errors.New(
			"failed to encode segment name",
			errors.Error("error", err),
		)
	}

	// Open an existing mapping. CreateFileMapping returns
	// ERROR_ALREADY_EXISTS when it was already created by another
	// process. If err is nil it did not exist and we created one
	// by accident so we clean it up and fail.
	mapping, err := windows.CreateFileMapping(
		windows.InvalidHandle,
		nil,
		windows.PAGE_READWRITE,
		0, 1,
		namePtr,
	)

	if err == nil {
		windows.CloseHandle(mapping)
		return ErrNotFound
	}

	if err != windows.ERROR_ALREADY_EXISTS {
		return errors.New(
			"failed to open file mapping",
			errors.Error("error", err),
		)
	}

	// Map into process address space
	access := uint32(windows.FILE_MAP_WRITE)
	if readOnly {
		access = windows.FILE_MAP_READ
	}

	addr, err := windows.MapViewOfFile(
		mapping, access,
		0, 0, 0,
	)

	if err != nil {
		windows.CloseHandle(mapping)
		return errors.New(
			"failed to map view",
			errors.Error("error", err),
		)
	}

	// Discover the mapped region size
	size, err := queryRegionSize(addr)
	if err != nil {
		windows.UnmapViewOfFile(addr)
		windows.CloseHandle(mapping)
		return err
	}

	s.h = &windowsHandle{
		buf:  sliceFromAddr(addr, size),
		addr: addr,

		mapping: mapping,
		file:    windows.InvalidHandle,

		owner: false,
	}

	s.technique = TechniqueSharedMemory
	return nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

func createFileBacked(s *Segment) error {

	//----------------------------------------------------------------------------//

	opts := s.opts
	name := strings.TrimPrefix(opts.Name, "/")
	path := filepath.Join(os.TempDir(), name)

	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return errors.New(
			"failed to encode file path",
			errors.Error("error", err),
		)
	}

	// Create the backing file
	file, err := windows.CreateFile(
		pathPtr,
		windows.GENERIC_READ|windows.GENERIC_WRITE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		nil,
		windows.CREATE_ALWAYS,
		windows.FILE_ATTRIBUTE_NORMAL,
		0,
	)

	if err != nil {
		return errors.New(
			"failed to create backing file",
			errors.Error("error", err),
			errors.String("path", path),
		)
	}

	// Create a file-backed mapping
	mapping, err := windows.CreateFileMapping(
		file,
		nil,
		windows.PAGE_READWRITE,
		uint32(int64(opts.Size)>>32),
		uint32(opts.Size),
		nil,
	)

	if err != nil {
		windows.CloseHandle(file)
		return errors.New(
			"failed to create file mapping",
			errors.Error("error", err),
		)
	}

	// Map into process address space
	addr, err := windows.MapViewOfFile(
		mapping,
		windows.FILE_MAP_WRITE,
		0, 0, 0,
	)

	if err != nil {
		windows.CloseHandle(mapping)
		windows.CloseHandle(file)
		return errors.New(
			"failed to map view",
			errors.Error("error", err),
		)
	}

	s.h = &windowsHandle{
		buf:  sliceFromAddr(addr, opts.Size),
		addr: addr,

		mapping: mapping,
		file:    file,

		path:  path,
		owner: true,
	}

	s.technique = TechniqueFileBacked
	return nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

func openFileBacked(s *Segment, readOnly bool) error {

	//----------------------------------------------------------------------------//

	opts := s.opts
	name := strings.TrimPrefix(opts.Name, "/")
	path := filepath.Join(os.TempDir(), name)

	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return errors.New(
			"failed to encode file path",
			errors.Error("error", err),
		)
	}

	// Open existing backing file
	fileAccess := uint32(windows.GENERIC_READ | windows.GENERIC_WRITE)
	fileProt := uint32(windows.PAGE_READWRITE)
	mapAccess := uint32(windows.FILE_MAP_WRITE)
	if readOnly {
		fileAccess = windows.GENERIC_READ
		fileProt = windows.PAGE_READONLY
		mapAccess = windows.FILE_MAP_READ
	}

	file, err := windows.CreateFile(
		pathPtr,
		fileAccess,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_ATTRIBUTE_NORMAL,
		0,
	)

	if err != nil {
		return errors.New(
			"failed to open backing file",
			errors.Error("error", err),
			errors.String("path", path),
		)
	}

	// Create mapping from the existing file
	mapping, err := windows.CreateFileMapping(
		file,
		nil,
		fileProt,
		0, 0,
		nil,
	)

	if err != nil {
		windows.CloseHandle(file)
		return errors.New(
			"failed to create file mapping",
			errors.Error("error", err),
		)
	}

	// Map into process address space
	addr, err := windows.MapViewOfFile(
		mapping, mapAccess,
		0, 0, 0,
	)

	if err != nil {
		windows.CloseHandle(mapping)
		windows.CloseHandle(file)
		return errors.New(
			"failed to map view",
			errors.Error("error", err),
		)
	}

	// Discover the mapped region size
	size, err := queryRegionSize(addr)
	if err != nil {
		windows.UnmapViewOfFile(addr)
		windows.CloseHandle(mapping)
		windows.CloseHandle(file)
		return err
	}

	s.h = &windowsHandle{
		buf:  sliceFromAddr(addr, size),
		addr: addr,

		mapping: mapping,
		file:    file,

		path:  path,
		owner: false,
	}

	s.technique = TechniqueFileBacked
	return nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

func queryRegionSize(addr uintptr) (int, error) {

	var info windows.MemoryBasicInformation
	err := windows.VirtualQuery(addr, &info, unsafe.Sizeof(info))
	if err != nil {
		return 0, errors.New(
			"failed to query region size",
			errors.Error("error", err),
		)
	}

	return int(info.RegionSize), nil
}

////////////////////////////////////////////////////////////////////////////////

// sliceFromAddr constructs a byte slice whose backing array starts at
// the given address without triggering the go vet false positive for
// uintptr to unsafe.Pointer conversion.
func sliceFromAddr(addr uintptr, size int) []byte {

	var b []byte
	hdr := (*struct {
		Data uintptr
		Len  int
		Cap  int
	})(unsafe.Pointer(&b))

	hdr.Data = addr
	hdr.Len = size
	hdr.Cap = size

	return b
}
