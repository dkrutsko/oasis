package main

import (
	"github.com/dkrutsko/oasis/errors"
)

////////////////////////////////////////////////////////////////////////////////

const (
	// Shortest match an LZ4 sequence can hold
	lz4MinMatch = 4

	// Lengths that reach this value continue in the bytes that
	// follow, each adding up to 255 more
	lz4LengthExtended = 15
)

////////////////////////////////////////////////////////////////////////////////

// decodeLz4Block decodes one raw LZ4 block from `source` into
// `target`, starting at `target[start:]`. Matches may reference
// the bytes already in `target[:start]`, which is how chained
// blocks share history. It returns the number of bytes written.
func decodeLz4Block(target []byte, start int, source []byte) (int, error) {

	reader := &binaryReader{data: source}
	written := start

	for reader.GetRemaining() > 0 {

		//----------------------------------------------------------------------------//

		// The token holds the literal length in its high bits and
		// the match length in its low bits
		token, err := reader.ReadUint8()
		if err != nil {
			return 0, err
		}

		literalLength, err := readLz4Length(reader, int(token>>4))
		if err != nil {
			return 0, err
		}

		literals, err := reader.ReadBytes(literalLength)
		if err != nil {
			return 0, err
		}

		if len(literals) > len(target)-written {
			return 0, errors.New(
				"lz4 literals are out of bounds",
				errors.Int("length", len(literals)),
			)
		}

		copy(target[written:], literals)
		written += len(literals)

		// The last sequence of a block has literals only
		if reader.GetRemaining() == 0 {
			break
		}

		//----------------------------------------------------------------------------//

		value, err := reader.ReadUint16()
		if err != nil {
			return 0, err
		}

		offset := int(value)
		if offset == 0 || offset > written {
			return 0, errors.New(
				"lz4 match offset is out of bounds",
				errors.Int("offset", offset),
			)
		}

		matchLength, err := readLz4Length(reader, int(token&0xF))
		if err != nil {
			return 0, err
		}

		matchLength += lz4MinMatch

		if matchLength > len(target)-written {
			return 0, errors.New(
				"lz4 match is out of bounds",
				errors.Int("length", matchLength),
			)
		}

		// A match that overlaps its own output repeats the most
		// recent bytes, so it is copied one byte at a time
		from := written - offset

		if offset >= matchLength {
			copy(target[written:written+matchLength], target[from:from+matchLength])
		} else {
			for i := 0; i < matchLength; i++ {
				target[written+i] = target[from+i]
			}
		}

		written += matchLength

		//----------------------------------------------------------------------------//
	}

	return written - start, nil
}

////////////////////////////////////////////////////////////////////////////////

// readLz4Length finishes reading a literal or match length that
// starts in a token.
func readLz4Length(reader *binaryReader, length int) (int, error) {

	if length != lz4LengthExtended {
		return length, nil
	}

	for {
		value, err := reader.ReadUint8()
		if err != nil {
			return 0, err
		}

		length += int(value)
		if value != 255 {
			return length, nil
		}
	}
}
