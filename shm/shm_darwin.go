package shm

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

////////////////////////////////////////////////////////////////////////////////

func (h *unixHandle) unlinked() bool {

	// macOS kernel-managed shm segments do not report a
	// meaningful Nlink via fstat. Instead, try to re-open
	// the segment by name. If the open fails, the segment
	// has been unlinked. If it succeeds but points to a
	// different kernel object (different inode), the
	// segment was replaced.
	if h.name == "" {
		return false
	}

	fd, err := shmOpen(h.name, unix.O_RDONLY, 0)
	if err != nil {
		return true
	}
	defer unix.Close(fd)

	// Compare inodes to detect replacement. If Moonlight
	// unlinks and recreates the segment, the name resolves
	// to a new kernel object while our mmap still points
	// to the old one.
	var oldStat, newStat unix.Stat_t
	if unix.Fstat(h.fd, &oldStat) != nil {
		return true
	}
	if unix.Fstat(fd, &newStat) != nil {
		return true
	}

	return oldStat.Ino != newStat.Ino
}
