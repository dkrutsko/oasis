package main

import "github.com/dkrutsko/oasis/errors"

////////////////////////////////////////////////////////////////////////////////

var errLz4Corrupt = errors.New("corrupt lz4 block")

////////////////////////////////////////////////////////////////////////////////

// lz4DecodeBlock decodes one raw LZ4 block from src into dst starting at
// dst[start:]. Matches may reference bytes already in dst[:start], which is
// how chained blocks (LZ4ChainDecoder in VRF) share history. It returns the
// number of bytes written.
func lz4DecodeBlock(dst []byte, start int, src []byte) (int, error) {

	di := start
	si := 0

	for si < len(src) {
		token := src[si]
		si++

		// Literal length, extended with 255-valued bytes
		literals := int(token >> 4)
		if literals == 15 {
			for {
				if si >= len(src) {
					return 0, errLz4Corrupt
				}
				b := src[si]
				si++
				literals += int(b)
				if b != 255 {
					break
				}
			}
		}

		if literals > len(src)-si || literals > len(dst)-di {
			return 0, errLz4Corrupt
		}
		copy(dst[di:], src[si:si+literals])
		si += literals
		di += literals

		// The last sequence of a block has literals only
		if si >= len(src) {
			break
		}

		if len(src)-si < 2 {
			return 0, errLz4Corrupt
		}
		offset := int(src[si]) | int(src[si+1])<<8
		si += 2

		if offset == 0 || offset > di {
			return 0, errLz4Corrupt
		}

		// Match length, extended the same way, with a minimum of 4
		match := int(token & 15)
		if match == 15 {
			for {
				if si >= len(src) {
					return 0, errLz4Corrupt
				}
				b := src[si]
				si++
				match += int(b)
				if b != 255 {
					break
				}
			}
		}
		match += 4

		if match > len(dst)-di {
			return 0, errLz4Corrupt
		}

		// Overlapping matches repeat recent bytes, so copy forward one byte
		// at a time when the source and destination overlap
		from := di - offset
		if offset >= match {
			copy(dst[di:di+match], dst[from:from+match])
		} else {
			for i := 0; i < match; i++ {
				dst[di+i] = dst[from+i]
			}
		}
		di += match
	}

	return di - start, nil
}
