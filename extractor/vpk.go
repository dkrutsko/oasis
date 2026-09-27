package main

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"io"

	"github.com/dkrutsko/oasis/errors"
)

////////////////////////////////////////////////////////////////////////////////

const (
	vpkMagic = 0x55AA1234

	// Archive index meaning the entry data follows the directory tree in
	// the same file rather than in a numbered archive file.
	vpkSameArchive = 0x7FFF

	vpkEntryTerminator = 0xFFFF
)

////////////////////////////////////////////////////////////////////////////////

// VpkEntry is a single file in a VPK directory tree.
type VpkEntry struct {
	Extension string
	Directory string
	Name      string

	crc32        uint32
	preload      []byte
	archiveIndex uint16
	offset       uint32
	length       uint32
}

// VpkExtension groups the entries sharing one extension in tree order.
// This mirrors the per-extension lists ValvePak exposes as `Entries`.
type VpkExtension struct {
	Name    string
	Entries []*VpkEntry
}

// Vpk is a parsed VPK directory tree. Entry data is read on demand from the
// underlying reader.
type Vpk struct {
	Extensions []VpkExtension

	reader     io.ReaderAt
	dataOffset int64
}

////////////////////////////////////////////////////////////////////////////////

// ReadVpk parses the header and directory tree of a version 1 or 2 VPK.
func ReadVpk(r io.ReaderAt) (*Vpk, error) {

	//----------------------------------------------------------------------------//

	header := make([]byte, 28)
	if _, err := r.ReadAt(header[:12], 0); err != nil {
		return nil, errors.New(
			"failed to read vpk header",
			errors.Error("error", err),
		)
	}

	if binary.LittleEndian.Uint32(header[0:]) != vpkMagic {
		return nil, errors.New("file is not a vpk")
	}

	version := binary.LittleEndian.Uint32(header[4:])
	treeSize := binary.LittleEndian.Uint32(header[8:])

	var headerSize int64
	switch version {
	case 1:
		headerSize = 12
	case 2:
		headerSize = 28
	default:
		return nil, errors.New(
			"unsupported vpk version",
			errors.Uint32("version", version),
		)
	}

	//----------------------------------------------------------------------------//

	tree := make([]byte, treeSize)
	if _, err := r.ReadAt(tree, headerSize); err != nil {
		return nil, errors.New(
			"failed to read vpk tree",
			errors.Uint32("tree_size", treeSize),
			errors.Error("error", err),
		)
	}

	extensions, consumed, err := parseVpkTree(tree)
	if err != nil {
		return nil, err
	}

	// ValvePak locates entry data using the size of the tree it actually
	// parsed, so do the same here.
	return &Vpk{
		Extensions: extensions,
		reader:     r,
		dataOffset: headerSize + int64(consumed),
	}, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// GetExtension returns the entries for an exact (case sensitive) extension,
// or nil when the package has none.
func (v *Vpk) GetExtension(name string) *VpkExtension {

	for i := range v.Extensions {
		if v.Extensions[i].Name == name {
			return &v.Extensions[i]
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

// ReadEntry returns the full contents of an entry and verifies its CRC32,
// matching ValvePak's default `ReadEntry` behavior.
func (v *Vpk) ReadEntry(entry *VpkEntry) ([]byte, error) {

	//----------------------------------------------------------------------------//

	data := make([]byte, len(entry.preload)+int(entry.length))
	copy(data, entry.preload)

	if entry.length > 0 {
		if entry.archiveIndex != vpkSameArchive {
			return nil, errors.New(
				"multi-file vpks are not supported",
				errors.String("directory", entry.Directory),
				errors.String("name", entry.Name),
				errors.Uint16("archive_index", entry.archiveIndex),
			)
		}

		offset := v.dataOffset + int64(entry.offset)
		if _, err := v.reader.ReadAt(data[len(entry.preload):], offset); err != nil {
			return nil, errors.New(
				"failed to read vpk entry",
				errors.String("directory", entry.Directory),
				errors.String("name", entry.Name),
				errors.Error("error", err),
			)
		}
	}

	//----------------------------------------------------------------------------//

	if actual := crc32.ChecksumIEEE(data); actual != entry.crc32 {
		return nil, errors.New(
			"vpk entry crc32 mismatch",
			errors.String("directory", entry.Directory),
			errors.String("name", entry.Name),
			errors.Uint32("expected", entry.crc32),
			errors.Uint32("actual", actual),
		)
	}

	return data, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// parseVpkTree walks the extension, directory and file levels of the tree.
// Each level ends with an empty string. It returns the bytes consumed.
func parseVpkTree(tree []byte) ([]VpkExtension, int, error) {

	pos := 0
	readString := func() (string, error) {
		end := bytes.IndexByte(tree[pos:], 0)
		if end < 0 {
			return "", errors.New("vpk tree string is not terminated")
		}
		s := string(tree[pos : pos+end])
		pos += end + 1
		return s, nil
	}

	var extensions []VpkExtension

	for {
		extension, err := readString()
		if err != nil {
			return nil, 0, err
		}
		if extension == "" {
			break
		}

		group := VpkExtension{Name: extension}

		for {
			directory, err := readString()
			if err != nil {
				return nil, 0, err
			}
			if directory == "" {
				break
			}

			for {
				name, err := readString()
				if err != nil {
					return nil, 0, err
				}
				if name == "" {
					break
				}

				if len(tree)-pos < 18 {
					return nil, 0, errors.New("vpk tree entry is truncated")
				}

				entry := &VpkEntry{
					Extension:    extension,
					Directory:    directory,
					Name:         name,
					crc32:        binary.LittleEndian.Uint32(tree[pos:]),
					archiveIndex: binary.LittleEndian.Uint16(tree[pos+6:]),
					offset:       binary.LittleEndian.Uint32(tree[pos+8:]),
					length:       binary.LittleEndian.Uint32(tree[pos+12:]),
				}
				preloadSize := int(binary.LittleEndian.Uint16(tree[pos+4:]))
				terminator := binary.LittleEndian.Uint16(tree[pos+16:])
				pos += 18

				if terminator != vpkEntryTerminator {
					return nil, 0, errors.New(
						"invalid vpk entry terminator",
						errors.Uint16("terminator", terminator),
					)
				}

				if preloadSize > 0 {
					if len(tree)-pos < preloadSize {
						return nil, 0, errors.New("vpk preload data is truncated")
					}
					entry.preload = tree[pos : pos+preloadSize]
					pos += preloadSize
				}

				group.Entries = append(group.Entries, entry)
			}
		}

		extensions = append(extensions, group)
	}

	return extensions, pos, nil
}
