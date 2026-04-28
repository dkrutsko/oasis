package shmem

import "golang.org/x/sys/unix"

////////////////////////////////////////////////////////////////////////////////

func shmOpen(name string, oflag int, perm uint32) (int, error) {

	return unix.Open("/dev/shm"+name, oflag, perm)
}

////////////////////////////////////////////////////////////////////////////////

func shmUnlink(name string) error {

	return unix.Unlink("/dev/shm" + name)
}
