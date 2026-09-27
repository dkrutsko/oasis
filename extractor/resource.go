package main

import (
	"encoding/binary"

	"github.com/dkrutsko/oasis/errors"
)

////////////////////////////////////////////////////////////////////////////////

// Header version of compiled Source 2 resources (files ending in `_c`).
const resourceHeaderVersion = 12

////////////////////////////////////////////////////////////////////////////////

// ResourceBlock is one block of a compiled resource, such as PHYS or DATA.
type ResourceBlock struct {
	Type string
	Data []byte
}

////////////////////////////////////////////////////////////////////////////////

// ReadResourceBlocks parses the block table of a compiled resource. Blocks
// with a size of zero are skipped, as VRF does.
func ReadResourceBlocks(data []byte) ([]ResourceBlock, error) {

	//----------------------------------------------------------------------------//

	if len(data) < 16 {
		return nil, errors.New(
			"resource is too small",
			errors.Int("size", len(data)),
		)
	}

	headerVersion := binary.LittleEndian.Uint16(data[4:])
	if headerVersion != resourceHeaderVersion {
		return nil, errors.New(
			"unexpected resource header version",
			errors.Uint16("version", headerVersion),
		)
	}

	// The block offset is relative to its own position (byte 8)
	blockStart := 8 + int(binary.LittleEndian.Uint32(data[8:]))
	blockCount := int(binary.LittleEndian.Uint32(data[12:]))

	//----------------------------------------------------------------------------//

	blocks := make([]ResourceBlock, 0, blockCount)

	for i := 0; i < blockCount; i++ {
		entry := blockStart + i*12
		if entry < 0 || entry+12 > len(data) {
			return nil, errors.New("resource block table is truncated")
		}

		// Each block offset is relative to the position of the offset field
		offset := entry + 4 + int(binary.LittleEndian.Uint32(data[entry+4:]))
		size := int(binary.LittleEndian.Uint32(data[entry+8:]))

		if size == 0 {
			continue
		}

		if offset < 0 || offset+size > len(data) {
			return nil, errors.New(
				"resource block is out of bounds",
				errors.Int("block", i),
			)
		}

		blocks = append(blocks, ResourceBlock{
			Type: string(data[entry : entry+4]),
			Data: data[offset : offset+size],
		})
	}

	return blocks, nil

	//----------------------------------------------------------------------------//
}
