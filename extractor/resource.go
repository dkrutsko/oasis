package main

import (
	"encoding/binary"

	"github.com/dkrutsko/oasis/errors"
)

////////////////////////////////////////////////////////////////////////////////

const (
	// Header version of compiled resources (files ending in `_c`)
	resourceHeaderVersion = 12

	// Size of the resource header and of each block table entry
	resourceHeaderSize     = 16
	resourceBlockEntrySize = 12
)

////////////////////////////////////////////////////////////////////////////////

// ReadResourceBlock returns the data of the first block of the
// given type, such as "PHYS" or "DATA", in a compiled resource.
// Blocks with a size of zero are skipped, as VRF does.
func ReadResourceBlock(data []byte, blockType string) ([]byte, error) {

	//----------------------------------------------------------------------------//

	if len(data) < resourceHeaderSize {
		return nil, errors.New(
			"resource is too small",
			errors.Int("size", len(data)),
		)
	}

	version := binary.LittleEndian.Uint16(data[4:6])
	if version != resourceHeaderVersion {
		return nil, errors.New(
			"unexpected resource header version",
			errors.Uint16("version", version),
		)
	}

	// The table offset is relative to the field that holds it
	tableStart := 8 + int(binary.LittleEndian.Uint32(data[8:12]))
	blockCount := int(binary.LittleEndian.Uint32(data[12:16]))

	if tableStart+blockCount*resourceBlockEntrySize > len(data) {
		return nil, errors.New(
			"resource block table is out of bounds",
			errors.Int("block_count", blockCount),
		)
	}

	//----------------------------------------------------------------------------//

	for i := 0; i < blockCount; i++ {
		entry := tableStart + i*resourceBlockEntrySize

		if string(data[entry:entry+4]) != blockType {
			continue
		}

		// Each block offset is relative to the field that holds it
		offset := entry + 4 + int(binary.LittleEndian.Uint32(data[entry+4:entry+8]))
		size := int(binary.LittleEndian.Uint32(data[entry+8 : entry+12]))

		if size == 0 {
			continue
		}

		if offset+size > len(data) {
			return nil, errors.New(
				"resource block is out of bounds",
				errors.String("type", blockType),
			)
		}

		return data[offset : offset+size], nil
	}

	return nil, errors.New(
		"resource block not found",
		errors.String("type", blockType),
	)

	//----------------------------------------------------------------------------//
}
