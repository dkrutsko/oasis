// Package shm provides cross-platform shared memory between
// processes. It supports POSIX shared memory on unix, pagefile-
// backed file mappings on Windows, and file-backed memory mapping
// on all platforms. The backing store is selected via the
// Technique field in Options.
package shm

import (
	"os"
	"strings"
	"sync"

	"github.com/dkrutsko/oasis/errors"
)

////////////////////////////////////////////////////////////////////////////////

// TechniqueType selects the backing store for the shared memory
// segment.
type TechniqueType int

const (
	// TechniqueAuto tries TechniqueSharedMemory first and falls
	// back to TechniqueFileBacked if the system does not support
	// it or the request exceeds a platform limit.
	TechniqueAuto TechniqueType = iota

	// TechniqueSharedMemory uses POSIX shared memory on unix
	// systems and pagefile-backed file mappings on Windows. The
	// segment lives entirely in RAM.
	TechniqueSharedMemory

	// TechniqueFileBacked uses a regular file in the system
	// temporary directory with a shared memory map. The operating
	// system keeps hot pages in the page cache.
	TechniqueFileBacked
)

////////////////////////////////////////////////////////////////////////////////

// String returns a human-readable name for the technique.
func (t TechniqueType) String() string {

	switch t {
	case TechniqueAuto:
		return "auto"
	case TechniqueSharedMemory:
		return "shared_memory"
	case TechniqueFileBacked:
		return "file_backed"
	default:
		return "unknown"
	}
}

////////////////////////////////////////////////////////////////////////////////

var (
	// ErrAlreadyCreated is returned when Create or Open is called
	// on a segment that already has an active mapping.
	ErrAlreadyCreated = errors.New("segment already created")

	// ErrInvalidName is returned when the segment name is empty.
	ErrInvalidName = errors.New("segment name must not be empty")

	// ErrInvalidSize is returned when the segment size is not
	// positive.
	ErrInvalidSize = errors.New("segment size must be greater than zero")

	// ErrNotFound is returned by Open when the named segment does
	// not exist.
	ErrNotFound = errors.New("segment does not exist")

	// ErrEmpty is returned by Open when the segment exists but has
	// zero size.
	ErrEmpty = errors.New("segment is empty")

	// ErrUnsafeName is returned when the segment name contains
	// path traversal components such as "..".
	ErrUnsafeName = errors.New("segment name contains unsafe path components")

	// ErrUnsupportedTechnique is returned when the Technique field
	// contains an unrecognized value.
	ErrUnsupportedTechnique = errors.New("unsupported technique")
)

////////////////////////////////////////////////////////////////////////////////

// Options configures a shared memory segment.
type Options struct {
	// Name identifies the segment. On unix systems it must start
	// with a forward slash (e.g. "/my_segment"). On Windows the
	// name is passed directly to CreateFileMapping and may contain
	// any characters except backslash.
	Name string

	// Size is the segment size in bytes. Only required when
	// calling Create.
	Size int

	// Perm sets the file mode bits for the segment. Only applies
	// when calling Create on unix systems. Ignored on Windows.
	// Defaults to 0600 if zero.
	Perm os.FileMode

	// Technique selects the backing store. Defaults to
	// TechniqueAuto.
	Technique TechniqueType
}

////////////////////////////////////////////////////////////////////////////////

type handle interface {
	data() []byte
	size() int
	flush() error
	close() error
	unlinked() bool
}

////////////////////////////////////////////////////////////////////////////////

// Segment represents a named shared memory segment backed by the
// operating system. A Segment is safe for concurrent use by
// multiple goroutines. Use New to construct, then Create or Open
// to initialize.
type Segment struct {
	opts      *Options
	h         handle
	technique TechniqueType
	mu        sync.RWMutex
}

////////////////////////////////////////////////////////////////////////////////

// New returns a new Segment configured with the given options. Call
// Create to allocate a new segment or Open to attach to an existing
// one.
func New(opts *Options) *Segment {

	return &Segment{opts: opts}
}

////////////////////////////////////////////////////////////////////////////////

// Create allocates a new shared memory segment. The caller is the
// owner and Close will clean up the segment.
func (s *Segment) Create() error {

	//----------------------------------------------------------------------------//

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.h != nil {
		return ErrAlreadyCreated
	}

	if s.opts.Name == "" {
		return ErrInvalidName
	}

	if strings.Contains(s.opts.Name, "..") {
		return ErrUnsafeName
	}

	if s.opts.Size <= 0 {
		return ErrInvalidSize
	}

	//----------------------------------------------------------------------------//

	return platformCreate(s)

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// Open attaches to an existing shared memory segment created by
// another process. The size is discovered automatically. When
// readOnly is true the mapping is read-only and writes will fault.
func (s *Segment) Open(readOnly bool) error {

	//----------------------------------------------------------------------------//

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.h != nil {
		return ErrAlreadyCreated
	}

	if s.opts.Name == "" {
		return ErrInvalidName
	}

	if strings.Contains(s.opts.Name, "..") {
		return ErrUnsafeName
	}

	//----------------------------------------------------------------------------//

	return platformOpen(s, readOnly)

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// IsCreated returns true if the segment has been successfully created
// or opened.
func (s *Segment) IsCreated() bool {

	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.h != nil
}

////////////////////////////////////////////////////////////////////////////////

// GetData returns the mapped memory as a byte slice. Writes to this
// slice are visible to all processes sharing the segment. The
// returned slice is backed by the mapped region and becomes invalid
// when Close is called. Callers must not use the slice after Close.
func (s *Segment) GetData() []byte {

	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.h == nil {
		return nil
	}

	return s.h.data()
}

////////////////////////////////////////////////////////////////////////////////

// GetSize returns the size of the mapped region in bytes.
func (s *Segment) GetSize() int {

	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.h == nil {
		return 0
	}

	return s.h.size()
}

////////////////////////////////////////////////////////////////////////////////

// GetTechnique returns the technique that was actually used to create
// the segment. This is useful when TechniqueAuto was requested.
func (s *Segment) GetTechnique() TechniqueType {

	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.technique
}

////////////////////////////////////////////////////////////////////////////////

// IsUnlinked returns true if the segment's backing name has been
// removed from the filesystem while the mapping is still open.
// On POSIX this checks that the file descriptor's link count has
// dropped to zero (e.g. after shm_unlink by another process).
func (s *Segment) IsUnlinked() bool {

	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.h == nil {
		return false
	}

	return s.h.unlinked()
}

////////////////////////////////////////////////////////////////////////////////

// Flush forces any pending writes to the backing store. For file-
// backed segments this syncs to disk. For shared memory segments
// this is a no-op since the data is already in RAM.
func (s *Segment) Flush() error {

	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.h == nil {
		return nil
	}

	return s.h.flush()
}

////////////////////////////////////////////////////////////////////////////////

// Close unmaps the shared memory and closes the underlying handle.
// If the segment was allocated with Create, it is also removed
// from the system. Any slices returned by GetData become invalid
// after Close returns.
func (s *Segment) Close() error {

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.h != nil {
		err := s.h.close()
		s.h = nil
		return err
	}

	return nil
}
