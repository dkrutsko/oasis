package main

import (
	"encoding/binary"
	"hash/crc32"
	"io"

	"github.com/dkrutsko/oasis/errors"
)

////////////////////////////////////////////////////////////////////////////////

const (
	// Signature at the start of every VPK file
	vpkSignature = 0x55AA1234

	// Header sizes of version 1 and version 2 VPK files
	vpkHeaderSizeV1 = 12
	vpkHeaderSizeV2 = 28

	// Size of a directory tree entry, not counting its name and
	// the preload data that follows it
	vpkEntrySize = 18

	// Archive index meaning the entry data follows the directory
	// tree in the same file rather than in a numbered archive
	vpkSameArchive = 0x7FFF

	// Value that ends every directory tree entry
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

////////////////////////////////////////////////////////////////////////////////

// Vpk is a parsed VPK directory tree. Entry data is read on
// demand from the underlying reader with `ReadEntry`.
type Vpk struct {
	Entries []VpkEntry

	reader     io.ReaderAt
	size       int64
	dataOffset int64
}

////////////////////////////////////////////////////////////////////////////////

// ReadVpk parses the header and directory tree of a version 1
// or 2 VPK. The size is that of the whole file and is used to
// reject trees and entries that point past its end.
func ReadVpk(r io.ReaderAt, size int64) (*Vpk, error) {

	//----------------------------------------------------------------------------//

	// Read the fields both versions share
	header := make([]byte, vpkHeaderSizeV1)

	_, err := r.ReadAt(header, 0)
	if err != nil {
		return nil, errors.New(
			"failed to read vpk header",
			errors.Error("error", err),
		)
	}

	signature := binary.LittleEndian.Uint32(header[0:4])
	version := binary.LittleEndian.Uint32(header[4:8])
	treeSize := int64(binary.LittleEndian.Uint32(header[8:12]))

	if signature != vpkSignature {
		return nil, errors.New(
			"file is not a vpk",
			errors.Uint32("signature", signature),
		)
	}

	var headerSize int64
	switch version {
	case 1:
		headerSize = vpkHeaderSizeV1
	case 2:
		headerSize = vpkHeaderSizeV2
	default:
		return nil, errors.New(
			"unsupported vpk version",
			errors.Uint32("version", version),
		)
	}

	//----------------------------------------------------------------------------//

	// Check the tree fits in the file before allocating it
	if headerSize+treeSize > size {
		return nil, errors.New(
			"vpk tree is out of bounds",
			errors.Int64("tree_size", treeSize),
			errors.Int64("file_size", size),
		)
	}

	tree := make([]byte, treeSize)

	_, err = r.ReadAt(tree, headerSize)
	if err != nil {
		return nil, errors.New(
			"failed to read vpk tree",
			errors.Error("error", err),
		)
	}

	entries, treeLength, err := parseVpkTree(tree)
	if err != nil {
		return nil, errors.New(
			"failed to parse vpk tree",
			errors.Error("error", err),
		)
	}

	//----------------------------------------------------------------------------//

	// Entry data starts after the tree that was actually parsed
	// rather than the tree size in the header, as in ValvePak
	return &Vpk{
		Entries:    entries,
		reader:     r,
		size:       size,
		dataOffset: headerSize + int64(treeLength),
	}, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// ReadEntry returns the full contents of an entry and verifies
// its CRC32. Only entries stored in the VPK itself are supported,
// not ones in numbered archive files.
func (v *Vpk) ReadEntry(entry *VpkEntry) ([]byte, error) {

	//----------------------------------------------------------------------------//

	offset := v.dataOffset + int64(entry.offset)
	length := int64(entry.length)

	// Check the data is in this file before allocating it
	if length > 0 {
		if entry.archiveIndex != vpkSameArchive {
			return nil, errors.New(
				"vpk entry is in a separate archive file",
				errors.String("directory", entry.Directory),
				errors.String("name", entry.Name),
				errors.Uint16("archive_index", entry.archiveIndex),
			)
		}

		if offset+length > v.size {
			return nil, errors.New(
				"vpk entry is out of bounds",
				errors.String("directory", entry.Directory),
				errors.String("name", entry.Name),
			)
		}
	}

	//----------------------------------------------------------------------------//

	// Preload data is stored in the tree and comes first
	data := make([]byte, int64(len(entry.preload))+length)
	copy(data, entry.preload)

	if length > 0 {
		_, err := v.reader.ReadAt(data[len(entry.preload):], offset)
		if err != nil {
			return nil, errors.New(
				"failed to read vpk entry",
				errors.String("directory", entry.Directory),
				errors.String("name", entry.Name),
				errors.Error("error", err),
			)
		}
	}

	//----------------------------------------------------------------------------//

	actual := crc32.ChecksumIEEE(data)
	if actual != entry.crc32 {
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

// parseVpkTree reads every entry in the directory tree and
// returns the number of bytes the tree used. The tree groups
// entries by extension, then by directory, and each level
// ends with an empty string.
func parseVpkTree(tree []byte) ([]VpkEntry, int, error) {

	reader := &binaryReader{data: tree}
	var entries []VpkEntry

	for {
		extension, err := reader.ReadString()
		if err != nil {
			return nil, 0, err
		}
		if extension == "" {
			break
		}

		for {
			directory, err := reader.ReadString()
			if err != nil {
				return nil, 0, err
			}
			if directory == "" {
				break
			}

			for {
				name, err := reader.ReadString()
				if err != nil {
					return nil, 0, err
				}
				if name == "" {
					break
				}

				entry, err := readVpkEntry(reader)
				if err != nil {
					return nil, 0, err
				}

				entry.Extension = extension
				entry.Directory = directory
				entry.Name = name

				entries = append(entries, entry)
			}
		}
	}

	return entries, reader.GetPosition(), nil
}

////////////////////////////////////////////////////////////////////////////////

// readVpkEntry reads the fixed part of a tree entry and the
// preload data that follows it.
func readVpkEntry(reader *binaryReader) (VpkEntry, error) {

	data, err := reader.ReadBytes(vpkEntrySize)
	if err != nil {
		return VpkEntry{}, err
	}

	entry := VpkEntry{
		crc32:        binary.LittleEndian.Uint32(data[0:4]),
		archiveIndex: binary.LittleEndian.Uint16(data[6:8]),
		offset:       binary.LittleEndian.Uint32(data[8:12]),
		length:       binary.LittleEndian.Uint32(data[12:16]),
	}

	preloadSize := int(binary.LittleEndian.Uint16(data[4:6]))
	terminator := binary.LittleEndian.Uint16(data[16:18])

	if terminator != vpkEntryTerminator {
		return VpkEntry{}, errors.New(
			"invalid vpk entry terminator",
			errors.Uint16("terminator", terminator),
		)
	}

	entry.preload, err = reader.ReadBytes(preloadSize)
	if err != nil {
		return VpkEntry{}, err
	}

	return entry, nil
}
