//go:build darwin || linux

package shm

import (
	"os"

	"golang.org/x/sys/unix"

	"github.com/dkrutsko/oasis/errors"
)

////////////////////////////////////////////////////////////////////////////////

type unixHandle struct {
	buf []byte
	fd  int

	name  string
	path  string
	owner bool
}

////////////////////////////////////////////////////////////////////////////////

func (h *unixHandle) data() []byte {

	return h.buf
}

////////////////////////////////////////////////////////////////////////////////

func (h *unixHandle) size() int {

	return len(h.buf)
}

////////////////////////////////////////////////////////////////////////////////

func (h *unixHandle) flush() error {

	// Only meaningful for file-backed segments
	if h.buf != nil && h.path != "" {
		err := unix.Msync(h.buf, unix.MS_SYNC)
		if err != nil {
			return errors.New(
				"failed to sync shared memory",
				errors.Error("error", err),
			)
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

func (h *unixHandle) close() error {

	if h.buf != nil {
		unix.Munmap(h.buf)
		h.buf = nil
	}

	if h.fd > 0 {
		unix.Close(h.fd)
		h.fd = -1
	}

	if h.owner {
		if h.name != "" {
			shmUnlink(h.name)
		}
		if h.path != "" {
			os.Remove(h.path)
		}
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

	perm := uint32(opts.Perm)
	if perm == 0 {
		perm = 0600
	}

	// Remove any stale segment from a previous run. On macOS,
	// ftruncate on an already-allocated segment returns EINVAL
	// so we must unlink first to get a fresh one.
	shmUnlink(opts.Name)

	// Create the shared memory segment
	fd, err := shmOpen(opts.Name, unix.O_CREAT|unix.O_RDWR, perm)
	if err != nil {
		return errors.New(
			"failed to open shared memory",
			errors.Error("error", err),
		)
	}

	// Resize to the requested size
	err = unix.Ftruncate(fd, int64(opts.Size))
	if err != nil {
		unix.Close(fd)
		return errors.New(
			"failed to resize shared memory",
			errors.Error("error", err),
		)
	}

	// Map into process address space
	data, err := unix.Mmap(
		fd, 0, opts.Size,
		unix.PROT_READ|unix.PROT_WRITE,
		unix.MAP_SHARED,
	)

	if err != nil {
		unix.Close(fd)
		return errors.New(
			"failed to map shared memory",
			errors.Error("error", err),
		)
	}

	s.h = &unixHandle{
		buf: data,
		fd:  fd,

		name:  opts.Name,
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

	// Open existing shared memory segment
	flags := unix.O_RDWR
	prot := unix.PROT_READ | unix.PROT_WRITE
	if readOnly {
		flags = unix.O_RDONLY
		prot = unix.PROT_READ
	}

	fd, err := shmOpen(opts.Name, flags, 0)
	if err != nil {
		return errors.New(
			"failed to open shared memory",
			errors.Error("error", err),
		)
	}

	// Discover the segment size
	size, err := fdSize(fd)
	if err != nil {
		unix.Close(fd)
		return err
	}

	// Map into process address space
	data, err := unix.Mmap(fd, 0, size, prot, unix.MAP_SHARED)
	if err != nil {
		unix.Close(fd)
		return errors.New(
			"failed to map shared memory",
			errors.Error("error", err),
		)
	}

	s.h = &unixHandle{
		buf: data,
		fd:  fd,

		name:  opts.Name,
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
	path := "/tmp" + opts.Name

	perm := uint32(opts.Perm)
	if perm == 0 {
		perm = 0600
	}

	// Create the backing file with O_NOFOLLOW to prevent
	// symlink attacks in world-writable directories
	fd, err := unix.Open(
		path,
		unix.O_CREAT|unix.O_RDWR|unix.O_NOFOLLOW,
		perm,
	)

	if err != nil {
		return errors.New(
			"failed to open backing file",
			errors.Error("error", err),
			errors.String("path", path),
		)
	}

	// Resize to the requested size
	err = unix.Ftruncate(fd, int64(opts.Size))
	if err != nil {
		unix.Close(fd)
		return errors.New(
			"failed to resize backing file",
			errors.Error("error", err),
		)
	}

	// Map into process address space
	data, err := unix.Mmap(
		fd, 0, opts.Size,
		unix.PROT_READ|unix.PROT_WRITE,
		unix.MAP_SHARED,
	)

	if err != nil {
		unix.Close(fd)
		return errors.New(
			"failed to map backing file",
			errors.Error("error", err),
		)
	}

	s.h = &unixHandle{
		buf: data,
		fd:  fd,

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
	path := "/tmp" + opts.Name

	// Open existing backing file
	flags := unix.O_RDWR | unix.O_NOFOLLOW
	prot := unix.PROT_READ | unix.PROT_WRITE
	if readOnly {
		flags = unix.O_RDONLY | unix.O_NOFOLLOW
		prot = unix.PROT_READ
	}

	fd, err := unix.Open(path, flags, 0)
	if err != nil {
		return errors.New(
			"failed to open backing file",
			errors.Error("error", err),
			errors.String("path", path),
		)
	}

	// Discover the file size
	size, err := fdSize(fd)
	if err != nil {
		unix.Close(fd)
		return err
	}

	// Map into process address space
	data, err := unix.Mmap(fd, 0, size, prot, unix.MAP_SHARED)
	if err != nil {
		unix.Close(fd)
		return errors.New(
			"failed to map backing file",
			errors.Error("error", err),
		)
	}

	s.h = &unixHandle{
		buf: data,
		fd:  fd,

		path:  path,
		owner: false,
	}

	s.technique = TechniqueFileBacked
	return nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

func fdSize(fd int) (int, error) {

	var stat unix.Stat_t
	err := unix.Fstat(fd, &stat)
	if err != nil {
		return 0, errors.New(
			"failed to stat segment",
			errors.Error("error", err),
		)
	}

	size := int(stat.Size)
	if size <= 0 {
		return 0, ErrEmpty
	}

	return size, nil
}
