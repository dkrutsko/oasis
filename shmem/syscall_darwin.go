package shmem

import (
	"unsafe"

	"golang.org/x/sys/unix"
)

////////////////////////////////////////////////////////////////////////////////

func shmOpen(name string, oflag int, perm uint32) (int, error) {

	namePtr, err := unix.BytePtrFromString(name)
	if err != nil {
		return -1, err
	}

	fd, _, errno := unix.Syscall(
		unix.SYS_SHM_OPEN,
		uintptr(unsafe.Pointer(namePtr)),
		uintptr(oflag),
		uintptr(perm),
	)

	if errno != 0 {
		return -1, errno
	}

	return int(fd), nil
}

////////////////////////////////////////////////////////////////////////////////

func shmUnlink(name string) error {

	namePtr, err := unix.BytePtrFromString(name)
	if err != nil {
		return err
	}

	_, _, errno := unix.Syscall(
		unix.SYS_SHM_UNLINK,
		uintptr(unsafe.Pointer(namePtr)),
		0, 0,
	)

	if errno != 0 {
		return errno
	}

	return nil
}
