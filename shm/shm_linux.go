package shm

import "golang.org/x/sys/unix"

////////////////////////////////////////////////////////////////////////////////

func shmOpen(name string, oflag int, perm uint32) (int, error) {

	return unix.Open("/dev/shm"+name, oflag, perm)
}

////////////////////////////////////////////////////////////////////////////////

func shmUnlink(name string) error {

	return unix.Unlink("/dev/shm" + name)
}

////////////////////////////////////////////////////////////////////////////////

func (h *unixHandle) unlinked() bool {

	if h.fd <= 0 {
		return false
	}

	// On Linux, shm segments live at /dev/shm. After
	// shm_unlink the link count drops to zero while the
	// fd remains valid.
	var stat unix.Stat_t
	err := unix.Fstat(h.fd, &stat)
	if err != nil {
		return false
	}

	return stat.Nlink == 0
}
