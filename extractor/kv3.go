package main

// Binary KeyValues3 decoder. This is a port of the reading side of
// ValveResourceFormat's `BinaryKV3.cs` (MIT License, Copyright (c) 2015
// ValveResourceFormat Contributors). See NOTICE for the license text.
//
// Versions 1 to 5 are supported. The legacy VKV3 format (version 0) is not.
// CS2 map physics currently use version 5 with zstd compression.

import (
	"bytes"
	"encoding/binary"
	"math"
	"strconv"
	"sync"

	"github.com/klauspost/compress/zstd"

	"github.com/dkrutsko/oasis/errors"
)

////////////////////////////////////////////////////////////////////////////////

const (
	kv3MagicLegacy = 0x03564B56 // "VKV\x03"
	kv3MagicPrefix = 0x4B563300 // "\x003VK" with the version in the low byte
	kv3Trailer     = 0xFFEEDD00

	kv3CompressionNone = 0
	kv3CompressionLz4  = 1
	kv3CompressionZstd = 2

	// LZ4 frame size VRF requires for version 2 and later
	kv3Lz4FrameSize = 16384
)

// Binary node types, matching `KV3BinaryNodeType` in VRF
const (
	kv3Null            = 1
	kv3Boolean         = 2
	kv3Int64           = 3
	kv3UInt64          = 4
	kv3Double          = 5
	kv3String          = 6
	kv3BinaryBlob      = 7
	kv3Array           = 8
	kv3Object          = 9
	kv3ArrayTyped      = 10
	kv3Int32           = 11
	kv3UInt32          = 12
	kv3BooleanTrue     = 13
	kv3BooleanFalse    = 14
	kv3Int64Zero       = 15
	kv3Int64One        = 16
	kv3DoubleZero      = 17
	kv3DoubleOne       = 18
	kv3Float           = 19
	kv3Int16           = 20
	kv3UInt16          = 21
	kv3Int32AsByte     = 23
	kv3ArrayByteLength = 24
	kv3ArrayAuxiliary  = 25
)

var errKv3Truncated = errors.New("kv3 data is truncated")

////////////////////////////////////////////////////////////////////////////////

// Kind is the type of a decoded KV3 value.
type Kind uint8

const (
	KindNull Kind = iota
	KindBool
	KindInt
	KindUInt
	KindFloat
	KindString
	KindBlob
	KindArray
	KindObject
)

// Value is a decoded KV3 value. Arrays and objects keep their children in
// order. Object children are named, array children are not. Blobs point
// into the decompressed buffer rather than being copied.
type Value struct {
	Kind Kind

	bits     uint64
	str      string
	blob     []byte
	children []Field
}

// Field is a named object child or an unnamed array element.
type Field struct {
	Name  string
	Value Value
}

////////////////////////////////////////////////////////////////////////////////

// GetChild returns the first child with the given name. It returns nil when
// the value is nil, is not an object, or has no such child.
func (v *Value) GetChild(name string) *Value {

	if v == nil || v.Kind != KindObject {
		return nil
	}

	for i := range v.children {
		if v.children[i].Name == name {
			return &v.children[i].Value
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

// GetElement returns the array element at index, or nil when the value is
// nil, is not an array, or the index is out of range.
func (v *Value) GetElement(index int) *Value {

	if v == nil || v.Kind != KindArray || index < 0 || index >= len(v.children) {
		return nil
	}
	return &v.children[index].Value
}

////////////////////////////////////////////////////////////////////////////////

// GetBlob returns the bytes of a binary blob, or nil for any other value.
func (v *Value) GetBlob() []byte {

	if v == nil || v.Kind != KindBlob {
		return nil
	}
	return v.blob
}

////////////////////////////////////////////////////////////////////////////////

// GetText returns the value as the original extractor's text parser saw it.
// Missing values are empty, strings are returned as is, numbers as decimal
// text, and other types as a non-empty placeholder that never parses as a
// number.
func (v *Value) GetText() string {

	if v == nil {
		return ""
	}

	switch v.Kind {
	case KindString:
		return v.str
	case KindInt:
		return strconv.FormatInt(int64(v.bits), 10)
	case KindUInt:
		return strconv.FormatUint(v.bits, 10)
	case KindFloat:
		return strconv.FormatFloat(math.Float64frombits(v.bits), 'g', -1, 64)
	case KindBool:
		if v.bits != 0 {
			return "1"
		}
		return "0"
	case KindNull:
		return "null"
	default:
		return "[" + strconv.Itoa(int(v.Kind)) + "]"
	}
}

////////////////////////////////////////////////////////////////////////////////

type kv3Buffers struct {
	bytes1 []byte
	bytes2 []byte
	bytes4 []byte
	bytes8 []byte
}

type kv3Context struct {
	version       int
	types         []byte
	objectLengths []byte
	blobs         []byte
	blobLengths   []byte
	strings       []string

	// Object fields to keep, or nil to keep everything. Other fields are
	// still read, since the format is sequential, but are not stored.
	keep map[string]bool

	// Version 5 splits values into two buffers. Auxiliary arrays read from
	// the other one, so the two are swapped while reading them.
	buffer    *kv3Buffers
	auxiliary *kv3Buffers
}

type kv3Header struct {
	version     int
	compression uint32

	dictionaryId uint16
	frameSize    uint16

	countBytes1 int
	countBytes2 int
	countBytes4 int
	countBytes8 int
	countTypes  int
	countBlocks int

	sizeUncompressedTotal    int
	sizeCompressedTotal      int
	sizeBlobs                int
	sizeBlockCompressedSizes int

	sizeUncompressedBuffer1 int
	sizeCompressedBuffer1   int
	sizeUncompressedBuffer2 int
	sizeCompressedBuffer2   int

	countBytes1Buffer2  int
	countBytes2Buffer2  int
	countBytes4Buffer2  int
	countBytes8Buffer2  int
	countObjectsBuffer2 int
}

////////////////////////////////////////////////////////////////////////////////

// DecodeKv3 decodes a binary KV3 block. When keep is not nil, only object
// fields with those names are stored, which saves memory when the caller
// needs a small part of a large block. Array elements are always stored.
// The zstd decoder must allow at least three concurrent `DecodeAll` calls.
func DecodeKv3(data []byte, keep map[string]bool, zd *zstd.Decoder) (*Value, error) {

	//----------------------------------------------------------------------------//

	if len(data) < 4 {
		return nil, errKv3Truncated
	}

	magic := binary.LittleEndian.Uint32(data)
	if magic == kv3MagicLegacy {
		return nil, errors.New("legacy vkv3 format is not supported")
	}

	version := int(magic & 0xFF)
	if magic&0xFFFFFF00 != kv3MagicPrefix || version < 1 || version > 5 {
		return nil, errors.New(
			"unsupported kv3 signature",
			errors.Uint32("magic", magic),
		)
	}

	h, pos, err := readKv3Header(data, version)
	if err != nil {
		return nil, errors.New(
			"failed to read kv3 header",
			errors.Error("error", err),
		)
	}

	//----------------------------------------------------------------------------//

	// Decompress the value buffers. In version 5 with zstd the buffers and
	// the blobs are three independent frames with known sizes, so decode
	// them concurrently.
	var raw1, raw2, rawBlobs []byte

	if version >= 5 && h.compression == kv3CompressionZstd {
		raw1, raw2, rawBlobs, pos, err = decodeKv3ZstdFrames(data, pos, h, zd)
		if err != nil {
			return nil, errors.New(
				"failed to read kv3 buffers",
				errors.Error("error", err),
			)
		}
	} else {
		size1 := h.sizeUncompressedBuffer1
		if version < 5 && h.compression == kv3CompressionZstd {
			// Before version 5 the blobs share the first zstd frame
			size1 += h.sizeBlobs
		}

		raw1, pos, err = readKv3Section(data, pos, h.compression, h.sizeCompressedBuffer1, size1, zd)
		if err != nil {
			return nil, errors.New(
				"failed to read kv3 buffer",
				errors.Int("buffer", 1),
				errors.Error("error", err),
			)
		}

		if version >= 5 {
			raw2, pos, err = readKv3Section(data, pos, h.compression, h.sizeCompressedBuffer2, h.sizeUncompressedBuffer2, zd)
			if err != nil {
				return nil, errors.New(
					"failed to read kv3 buffer",
					errors.Int("buffer", 2),
					errors.Error("error", err),
				)
			}
		}
	}

	//----------------------------------------------------------------------------//

	ctx := &kv3Context{version: version, keep: keep}

	blobSizes, err := ctx.readBuffer1(raw1[:h.sizeUncompressedBuffer1], h)
	if err != nil {
		return nil, errors.New(
			"failed to parse kv3 buffer",
			errors.Int("buffer", 1),
			errors.Error("error", err),
		)
	}

	if version >= 5 {
		blobSizes, err = ctx.readBuffer2(raw2, h)
		if err != nil {
			return nil, errors.New(
				"failed to parse kv3 buffer",
				errors.Int("buffer", 2),
				errors.Error("error", err),
			)
		}
	}

	//----------------------------------------------------------------------------//

	if h.countBlocks > 0 {
		pos, err = ctx.readBlobs(data, pos, h, blobSizes, raw1, rawBlobs)
		if err != nil {
			return nil, errors.New(
				"failed to read kv3 blobs",
				errors.Error("error", err),
			)
		}
	}

	//----------------------------------------------------------------------------//

	rootType, err := ctx.readType()
	if err != nil {
		return nil, errors.New(
			"failed to read kv3 values",
			errors.Error("error", err),
		)
	}

	root, err := ctx.readValue(rootType, false)
	if err != nil {
		return nil, errors.New(
			"failed to read kv3 values",
			errors.Error("error", err),
		)
	}

	return &root, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// readKv3Header reads the fields between the magic and the first buffer. It
// returns the position of the first buffer.
func readKv3Header(data []byte, version int) (kv3Header, int, error) {

	//----------------------------------------------------------------------------//

	h := kv3Header{version: version}
	r := &kv3HeaderReader{data: data, pos: 4}

	r.skip(16) // format GUID, not needed

	h.compression = r.uint32()

	if version == 1 {
		// Version 1 has no extra compression fields
		h.countBytes1 = r.int32()
		h.countBytes4 = r.int32()
		h.countBytes8 = r.int32()
		h.sizeUncompressedTotal = r.int32()
		h.sizeCompressedTotal = len(data) - r.pos
	} else {
		h.dictionaryId = r.uint16()
		h.frameSize = r.uint16()
		h.countBytes1 = r.int32()
		h.countBytes4 = r.int32()
		h.countBytes8 = r.int32()
		h.countTypes = r.int32()
		r.uint16() // object count
		r.uint16() // array count
		h.sizeUncompressedTotal = r.int32()
		h.sizeCompressedTotal = r.int32()
		h.countBlocks = r.int32()
		h.sizeBlobs = r.int32()
	}

	if version >= 4 {
		h.countBytes2 = r.int32()
		h.sizeBlockCompressedSizes = r.int32()
	}

	if version >= 5 {
		h.sizeUncompressedBuffer1 = r.int32()
		h.sizeCompressedBuffer1 = r.int32()
		h.sizeUncompressedBuffer2 = r.int32()
		h.sizeCompressedBuffer2 = r.int32()
		h.countBytes1Buffer2 = r.int32()
		h.countBytes2Buffer2 = r.int32()
		h.countBytes4Buffer2 = r.int32()
		h.countBytes8Buffer2 = r.int32()
		r.int32() // unknown
		h.countObjectsBuffer2 = r.int32()
		r.int32() // array count in buffer 2
		r.int32() // unknown
	} else {
		h.sizeCompressedBuffer1 = h.sizeCompressedTotal
		h.sizeUncompressedBuffer1 = h.sizeUncompressedTotal
	}

	if r.err != nil {
		return h, 0, r.err
	}

	//----------------------------------------------------------------------------//

	// Sizes and counts are signed in the format, reject negative ones
	for _, n := range []int{
		h.countBytes1, h.countBytes2, h.countBytes4, h.countBytes8, h.countTypes,
		h.countBlocks, h.sizeUncompressedTotal, h.sizeCompressedTotal, h.sizeBlobs,
		h.sizeUncompressedBuffer1, h.sizeCompressedBuffer1,
		h.sizeUncompressedBuffer2, h.sizeCompressedBuffer2,
		h.countBytes1Buffer2, h.countBytes2Buffer2, h.countBytes4Buffer2,
		h.countBytes8Buffer2, h.countObjectsBuffer2,
	} {
		if n < 0 {
			return h, 0, errors.New("kv3 header has a negative size")
		}
	}

	//----------------------------------------------------------------------------//

	// Same restrictions as VRF on the compression parameters
	switch h.compression {
	case kv3CompressionNone:
		if h.dictionaryId != 0 || h.frameSize != 0 {
			return h, 0, errors.New("unexpected kv3 compression parameters")
		}
	case kv3CompressionLz4:
		if h.dictionaryId != 0 || (version >= 2 && h.frameSize != kv3Lz4FrameSize) {
			return h, 0, errors.New("unexpected kv3 lz4 parameters")
		}
	case kv3CompressionZstd:
		if h.dictionaryId != 0 || h.frameSize != 0 {
			return h, 0, errors.New("unexpected kv3 zstd parameters")
		}
	default:
		return h, 0, errors.New(
			"unknown kv3 compression method",
			errors.Uint32("method", h.compression),
		)
	}

	return h, r.pos, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// decodeKv3ZstdFrames decodes both value buffers and the blobs of a version 5
// zstd block concurrently. It returns the position after the blob frame.
func decodeKv3ZstdFrames(data []byte, pos int, h kv3Header, zd *zstd.Decoder) ([]byte, []byte, []byte, int, error) {

	//----------------------------------------------------------------------------//

	type frame struct {
		src  []byte
		size int
		out  []byte
		err  error
	}

	frames := []*frame{
		{size: h.sizeUncompressedBuffer1},
		{size: h.sizeUncompressedBuffer2},
	}
	compressed := []int{h.sizeCompressedBuffer1, h.sizeCompressedBuffer2}

	if h.countBlocks > 0 {
		if h.sizeBlockCompressedSizes != 0 {
			return nil, nil, nil, 0, errors.New("unexpected kv3 block compressed sizes")
		}
		frames = append(frames, &frame{size: h.sizeBlobs})
		compressed = append(compressed, h.sizeCompressedTotal-h.sizeCompressedBuffer1-h.sizeCompressedBuffer2)
	}

	for i, f := range frames {
		src, err := sliceAt(data, pos, compressed[i])
		if err != nil {
			return nil, nil, nil, 0, err
		}
		f.src = src
		pos += compressed[i]
	}

	//----------------------------------------------------------------------------//

	var wg sync.WaitGroup
	for _, f := range frames {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f.out, f.err = zstdDecode(zd, f.src, f.size)
		}()
	}
	wg.Wait()

	for _, f := range frames {
		if f.err != nil {
			return nil, nil, nil, 0, errors.New(
				"failed to decompress kv3 zstd frame",
				errors.Error("error", f.err),
			)
		}
	}

	//----------------------------------------------------------------------------//

	var blobs []byte
	if len(frames) > 2 {
		blobs = frames[2].out
	}

	return frames[0].out, frames[1].out, blobs, pos, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// readKv3Section reads one buffer that is stored uncompressed or compressed
// with a single LZ4 block or zstd frame. It returns the position after it.
func readKv3Section(data []byte, pos int, method uint32, compressed, size int, zd *zstd.Decoder) ([]byte, int, error) {

	switch method {
	case kv3CompressionNone:
		src, err := sliceAt(data, pos, size)
		return src, pos + size, err

	case kv3CompressionLz4:
		src, err := sliceAt(data, pos, compressed)
		if err != nil {
			return nil, 0, err
		}
		out := make([]byte, size)
		n, err := lz4DecodeBlock(out, 0, src)
		if err != nil {
			return nil, 0, err
		}
		if n != size {
			return nil, 0, errors.New(
				"unexpected lz4 decoded size",
				errors.Int("decoded", n),
				errors.Int("expected", size),
			)
		}
		return out, pos + compressed, nil

	default:
		src, err := sliceAt(data, pos, compressed)
		if err != nil {
			return nil, 0, err
		}
		out, err := zstdDecode(zd, src, size)
		return out, pos + compressed, err
	}
}

////////////////////////////////////////////////////////////////////////////////

// readBuffer1 splits the first buffer into its value arrays and reads the
// string table. Before version 5 it also holds the types. It returns the
// remaining bytes that list the blob sizes, if any.
func (c *kv3Context) readBuffer1(buf []byte, h kv3Header) ([]byte, error) {

	//----------------------------------------------------------------------------//

	b := &kv3Buffers{}
	offset := 0

	if err := splitKv3Buffers(buf, &offset, b, h.countBytes1, h.countBytes2, h.countBytes4, h.countBytes8); err != nil {
		return nil, err
	}

	// Version 5 does not align when there are no 8 byte values, earlier
	// versions do
	if h.countBytes8 == 0 && c.version < 5 {
		offset = align(offset, 8)
	}

	countStrings, err := takeInt32(&b.bytes4)
	if err != nil {
		return nil, err
	}
	if countStrings < 0 {
		return nil, errors.New("kv3 string count is negative")
	}
	c.strings = make([]string, countStrings)

	//----------------------------------------------------------------------------//

	if c.version >= 5 {
		// Strings come first in the byte array of the first buffer, which
		// then serves as the auxiliary buffer
		c.auxiliary = b
		for i := range c.strings {
			if c.strings[i], err = takeCString(&b.bytes1); err != nil {
				return nil, err
			}
		}
		return nil, nil
	}

	//----------------------------------------------------------------------------//

	c.buffer = b

	if offset > len(buf) {
		return nil, errKv3Truncated
	}
	stringsBuf := buf[offset:]
	stringsStart := offset

	for i := range c.strings {
		before := len(stringsBuf)
		if c.strings[i], err = takeCString(&stringsBuf); err != nil {
			return nil, err
		}
		offset += before - len(stringsBuf)
	}

	var typesLength int
	if c.version == 1 {
		typesLength = h.sizeUncompressedTotal - offset - 4
	} else {
		typesLength = h.countTypes - offset + stringsStart
	}

	if c.types, err = sliceAt(buf, offset, typesLength); err != nil {
		return nil, err
	}
	offset += typesLength

	return readKv3TypesEnd(buf, offset, h.countBlocks)

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// readBuffer2 splits the version 5 second buffer, which holds the object
// lengths, the main value arrays and the types.
func (c *kv3Context) readBuffer2(buf []byte, h kv3Header) ([]byte, error) {

	b := &kv3Buffers{}
	c.buffer = b

	offset := h.countObjectsBuffer2 * 4
	lengths, err := sliceAt(buf, 0, offset)
	if err != nil {
		return nil, err
	}
	c.objectLengths = lengths

	if err := splitKv3Buffers(buf, &offset, b, h.countBytes1Buffer2, h.countBytes2Buffer2, h.countBytes4Buffer2, h.countBytes8Buffer2); err != nil {
		return nil, err
	}

	if c.types, err = sliceAt(buf, offset, h.countTypes); err != nil {
		return nil, err
	}
	offset += h.countTypes

	return readKv3TypesEnd(buf, offset, h.countBlocks)
}

////////////////////////////////////////////////////////////////////////////////

// readBlobs reads the blob lengths and the blob data that follows the value
// buffers, then checks the trailer. It returns the position after it.
func (c *kv3Context) readBlobs(data []byte, pos int, h kv3Header, blobSizes, raw1, rawBlobs []byte) (int, error) {

	//----------------------------------------------------------------------------//

	if c.version < 2 {
		return 0, errors.New("kv3 version 1 cannot have separate blobs")
	}

	lengths, err := sliceAt(blobSizes, 0, h.countBlocks*4)
	if err != nil {
		return 0, err
	}
	c.blobLengths = lengths
	blobSizes = blobSizes[h.countBlocks*4:]

	if err := expectTrailer(blobSizes); err != nil {
		return 0, err
	}
	blobSizes = blobSizes[4:]

	//----------------------------------------------------------------------------//

	switch h.compression {
	case kv3CompressionNone:
		if c.blobs, err = sliceAt(data, pos, h.sizeBlobs); err != nil {
			return 0, err
		}
		pos += h.sizeBlobs

	case kv3CompressionLz4:
		// Blobs are a chain of LZ4 blocks, each decoding to one frame and
		// allowed to reference the frames before it
		c.blobs = make([]byte, h.sizeBlobs)
		decoded := 0

		for len(blobSizes) >= 2 {
			compressed := int(binary.LittleEndian.Uint16(blobSizes))
			blobSizes = blobSizes[2:]

			frame := min(int(h.frameSize), h.sizeBlobs-decoded)
			src, err := sliceAt(data, pos, compressed)
			if err != nil {
				return 0, err
			}
			pos += compressed

			n, err := lz4DecodeBlock(c.blobs[:decoded+frame], decoded, src)
			if err != nil {
				return 0, err
			}
			if n < 1 {
				return 0, errors.New("lz4 blob frame decoded to nothing")
			}
			decoded += n
		}

	default:
		if c.version >= 5 {
			c.blobs = rawBlobs
		} else {
			// Before version 5 the blobs were decompressed with buffer 1
			if c.blobs, err = sliceAt(raw1, h.sizeUncompressedBuffer1, h.sizeBlobs); err != nil {
				return 0, err
			}
		}
	}

	//----------------------------------------------------------------------------//

	if pos > len(data) {
		return 0, errKv3Truncated
	}
	if err := expectTrailer(data[pos:]); err != nil {
		return 0, err
	}

	return pos + 4, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

func (c *kv3Context) readType() (byte, error) {

	if len(c.types) == 0 {
		return 0, errKv3Truncated
	}

	t := c.types[0]
	c.types = c.types[1:]

	// A set high bit means a flag byte follows. Flags only annotate values
	// (resource names and such), so they are skipped.
	if t&0x80 != 0 {
		if c.version >= 3 {
			t &= 0x3F
		} else {
			t &= 0x7F
		}

		if len(c.types) == 0 {
			return 0, errKv3Truncated
		}
		c.types = c.types[1:]
	}

	return t, nil
}

////////////////////////////////////////////////////////////////////////////////

// readChild reads the next child of an array or object. Object children
// are preceded by a string table index for their name. When discard is set
// the child is read but not stored.
func (c *kv3Context) readChild(parent *Value, discard bool) error {

	t, err := c.readType()
	if err != nil {
		return err
	}

	name := ""
	if parent.Kind == KindObject {
		id, err := takeInt32(&c.buffer.bytes4)
		if err != nil {
			return err
		}
		if name, err = c.getString(id); err != nil {
			return err
		}

		if c.keep != nil && !c.keep[name] {
			discard = true
		}
	}

	value, err := c.readValue(t, discard)
	if err != nil {
		return err
	}

	if !discard {
		parent.children = append(parent.children, Field{Name: name, Value: value})
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

// readValue reads one value of type t. When discard is set, containers are
// walked without storing their children, since the result is thrown away.
func (c *kv3Context) readValue(t byte, discard bool) (Value, error) {

	b := c.buffer

	switch t {

	// Values encoded in the type alone
	case kv3Null:
		return Value{Kind: KindNull}, nil
	case kv3BooleanTrue:
		return Value{Kind: KindBool, bits: 1}, nil
	case kv3BooleanFalse:
		return Value{Kind: KindBool}, nil
	case kv3Int64Zero:
		return Value{Kind: KindInt}, nil
	case kv3Int64One:
		return Value{Kind: KindInt, bits: 1}, nil
	case kv3DoubleZero:
		return Value{Kind: KindFloat, bits: math.Float64bits(0)}, nil
	case kv3DoubleOne:
		return Value{Kind: KindFloat, bits: math.Float64bits(1)}, nil

	// 1 byte values
	case kv3Boolean:
		v, err := take(&b.bytes1, 1)
		if err != nil {
			return Value{}, err
		}
		if v[0] == 1 {
			return Value{Kind: KindBool, bits: 1}, nil
		}
		return Value{Kind: KindBool}, nil
	case kv3Int32AsByte:
		v, err := take(&b.bytes1, 1)
		return Value{Kind: KindInt, bits: uint64(le8(v))}, err

	// 2 byte values
	case kv3Int16:
		v, err := take(&b.bytes2, 2)
		return Value{Kind: KindInt, bits: uint64(int64(int16(le16(v))))}, err
	case kv3UInt16:
		v, err := take(&b.bytes2, 2)
		return Value{Kind: KindUInt, bits: uint64(le16(v))}, err

	// 4 byte values
	case kv3Int32:
		v, err := take(&b.bytes4, 4)
		return Value{Kind: KindInt, bits: uint64(int64(int32(le32(v))))}, err
	case kv3UInt32:
		v, err := take(&b.bytes4, 4)
		return Value{Kind: KindUInt, bits: uint64(le32(v))}, err
	case kv3Float:
		v, err := take(&b.bytes4, 4)
		f := float64(math.Float32frombits(le32(v)))
		return Value{Kind: KindFloat, bits: math.Float64bits(f)}, err

	// 8 byte values
	case kv3Int64:
		v, err := take(&b.bytes8, 8)
		return Value{Kind: KindInt, bits: le64(v)}, err
	case kv3UInt64:
		v, err := take(&b.bytes8, 8)
		return Value{Kind: KindUInt, bits: le64(v)}, err
	case kv3Double:
		v, err := take(&b.bytes8, 8)
		return Value{Kind: KindFloat, bits: le64(v)}, err

	case kv3String:
		id, err := takeInt32(&b.bytes4)
		if err != nil {
			return Value{}, err
		}
		s, err := c.getString(id)
		return Value{Kind: KindString, str: s}, err

	case kv3BinaryBlob:
		return c.readBlob()

	case kv3Array:
		n, err := takeInt32(&b.bytes4)
		if err != nil {
			return Value{}, err
		}
		return c.readContainer(KindArray, n, discard)

	case kv3ArrayTyped, kv3ArrayByteLength:
		var n int
		var err error
		if t == kv3ArrayByteLength {
			v, err := take(&b.bytes1, 1)
			if err != nil {
				return Value{}, err
			}
			n = int(v[0])
		} else {
			if n, err = takeInt32(&b.bytes4); err != nil {
				return Value{}, err
			}
		}
		return c.readTypedArray(n, false, discard)

	case kv3ArrayAuxiliary:
		v, err := take(&b.bytes1, 1)
		if err != nil {
			return Value{}, err
		}
		return c.readTypedArray(int(v[0]), true, discard)

	case kv3Object:
		var n int
		var err error
		if c.version >= 5 {
			n, err = takeInt32(&c.objectLengths)
		} else {
			n, err = takeInt32(&b.bytes4)
		}
		if err != nil {
			return Value{}, err
		}
		return c.readContainer(KindObject, n, discard)

	default:
		return Value{}, errors.New(
			"unknown kv3 type",
			errors.Uint8("type", t),
		)
	}
}

////////////////////////////////////////////////////////////////////////////////

func (c *kv3Context) readBlob() (Value, error) {

	var length int
	var err error

	if c.version < 2 {
		length, err = takeInt32(&c.buffer.bytes4)
	} else {
		length, err = takeInt32(&c.blobLengths)
	}
	if err != nil {
		return Value{}, err
	}

	// Non-positive lengths are empty blobs, as in VRF
	if length <= 0 {
		return Value{Kind: KindBlob, blob: []byte{}}, nil
	}

	var data []byte
	if c.version < 2 {
		data, err = take(&c.buffer.bytes1, length)
	} else {
		data, err = take(&c.blobs, length)
	}
	return Value{Kind: KindBlob, blob: data}, err
}

////////////////////////////////////////////////////////////////////////////////

// readContainer reads the children of an array or object. Each child has
// its own type.
func (c *kv3Context) readContainer(kind Kind, n int, discard bool) (Value, error) {

	if n < 0 {
		return Value{}, errors.New("kv3 container length is negative")
	}

	value := Value{Kind: kind}

	// Every child needs at least one type byte, which bounds the allocation
	if !discard {
		value.children = make([]Field, 0, min(n, len(c.types)))
	}

	for i := 0; i < n; i++ {
		if err := c.readChild(&value, discard); err != nil {
			return Value{}, err
		}
	}
	return value, nil
}

////////////////////////////////////////////////////////////////////////////////

// readTypedArray reads an array whose elements all share one type. The
// auxiliary variant reads its elements from the other value buffer.
func (c *kv3Context) readTypedArray(n int, auxiliary, discard bool) (Value, error) {

	if n < 0 {
		return Value{}, errors.New("kv3 array length is negative")
	}

	t, err := c.readType()
	if err != nil {
		return Value{}, err
	}

	if auxiliary {
		if c.auxiliary == nil {
			return Value{}, errors.New("kv3 auxiliary array without an auxiliary buffer")
		}
		c.buffer, c.auxiliary = c.auxiliary, c.buffer
		defer func() { c.buffer, c.auxiliary = c.auxiliary, c.buffer }()
	}

	value := Value{Kind: KindArray}

	// Cap the preallocation so a corrupt length cannot exhaust memory
	if !discard {
		value.children = make([]Field, 0, min(n, 1<<20))
	}

	for i := 0; i < n; i++ {
		element, err := c.readValue(t, discard)
		if err != nil {
			return Value{}, err
		}
		if !discard {
			value.children = append(value.children, Field{Value: element})
		}
	}
	return value, nil
}

////////////////////////////////////////////////////////////////////////////////

func (c *kv3Context) getString(id int) (string, error) {

	if id == -1 {
		return "", nil
	}
	if id < 0 || id >= len(c.strings) {
		return "", errors.New(
			"kv3 string index is out of range",
			errors.Int("index", id),
		)
	}
	return c.strings[id], nil
}

////////////////////////////////////////////////////////////////////////////////
// Helpers
////////////////////////////////////////////////////////////////////////////////

// kv3HeaderReader reads little endian header fields and remembers the first
// error, so the header can be read as a flat list of fields.
type kv3HeaderReader struct {
	data []byte
	pos  int
	err  error
}

func (r *kv3HeaderReader) next(n int) []byte {

	if r.err != nil {
		return nil
	}
	if len(r.data)-r.pos < n {
		r.err = errKv3Truncated
		return nil
	}
	b := r.data[r.pos : r.pos+n]
	r.pos += n
	return b
}

func (r *kv3HeaderReader) skip(n int) { r.next(n) }

func (r *kv3HeaderReader) uint16() uint16 {

	if b := r.next(2); b != nil {
		return binary.LittleEndian.Uint16(b)
	}
	return 0
}

func (r *kv3HeaderReader) uint32() uint32 {

	if b := r.next(4); b != nil {
		return binary.LittleEndian.Uint32(b)
	}
	return 0
}

func (r *kv3HeaderReader) int32() int { return int(int32(r.uint32())) }

////////////////////////////////////////////////////////////////////////////////

// splitKv3Buffers slices the 1, 2, 4 and 8 byte value arrays out of buf,
// aligning each to its element size.
func splitKv3Buffers(buf []byte, offset *int, b *kv3Buffers, count1, count2, count4, count8 int) error {

	var err error

	if count1 > 0 {
		if b.bytes1, err = sliceAt(buf, *offset, count1); err != nil {
			return err
		}
		*offset += count1
	}

	if count2 > 0 {
		*offset = align(*offset, 2)
		if b.bytes2, err = sliceAt(buf, *offset, count2*2); err != nil {
			return err
		}
		*offset += count2 * 2
	}

	if count4 > 0 {
		*offset = align(*offset, 4)
		if b.bytes4, err = sliceAt(buf, *offset, count4*4); err != nil {
			return err
		}
		*offset += count4 * 4
	}

	if count8 > 0 {
		*offset = align(*offset, 8)
		if b.bytes8, err = sliceAt(buf, *offset, count8*8); err != nil {
			return err
		}
		*offset += count8 * 8
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

// readKv3TypesEnd handles what follows the types: a trailer when there are
// no blobs, otherwise the blob size lists, which are returned.
func readKv3TypesEnd(buf []byte, offset, countBlocks int) ([]byte, error) {

	if offset > len(buf) {
		return nil, errKv3Truncated
	}
	if countBlocks == 0 {
		return nil, expectTrailer(buf[offset:])
	}
	return buf[offset:], nil
}

////////////////////////////////////////////////////////////////////////////////

func zstdDecode(zd *zstd.Decoder, src []byte, size int) ([]byte, error) {

	out, err := zd.DecodeAll(src, make([]byte, 0, size))
	if err != nil {
		return nil, errors.New(
			"failed to decompress zstd frame",
			errors.Error("error", err),
		)
	}
	if len(out) != size {
		return nil, errors.New(
			"unexpected zstd decoded size",
			errors.Int("decoded", len(out)),
			errors.Int("expected", size),
		)
	}
	return out, nil
}

////////////////////////////////////////////////////////////////////////////////

func expectTrailer(b []byte) error {

	if len(b) < 4 {
		return errKv3Truncated
	}
	if trailer := binary.LittleEndian.Uint32(b); trailer != kv3Trailer {
		return errors.New(
			"unexpected kv3 trailer",
			errors.Uint32("trailer", trailer),
		)
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func sliceAt(b []byte, offset, n int) ([]byte, error) {

	if offset < 0 || n < 0 || offset > len(b) || n > len(b)-offset {
		return nil, errKv3Truncated
	}
	return b[offset : offset+n], nil
}

////////////////////////////////////////////////////////////////////////////////

// take removes and returns the first n bytes of *b.
func take(b *[]byte, n int) ([]byte, error) {

	if n < 0 || len(*b) < n {
		return nil, errKv3Truncated
	}
	out := (*b)[:n]
	*b = (*b)[n:]
	return out, nil
}

////////////////////////////////////////////////////////////////////////////////

func takeInt32(b *[]byte) (int, error) {

	v, err := take(b, 4)
	if err != nil {
		return 0, err
	}
	return int(int32(le32(v))), nil
}

////////////////////////////////////////////////////////////////////////////////

// takeCString removes a null terminated string from the front of *b.
func takeCString(b *[]byte) (string, error) {

	end := bytes.IndexByte(*b, 0)
	if end < 0 {
		return "", errors.New("kv3 string is not terminated")
	}
	s := string((*b)[:end])
	*b = (*b)[end+1:]
	return s, nil
}

////////////////////////////////////////////////////////////////////////////////

func align(offset, alignment int) int {
	return (offset + alignment - 1) &^ (alignment - 1)
}

// Little endian readers that tolerate the short slices returned alongside
// an error, so callers can build a value and return the error together
func le8(b []byte) byte {
	if len(b) < 1 {
		return 0
	}
	return b[0]
}

func le16(b []byte) uint16 {
	if len(b) < 2 {
		return 0
	}
	return binary.LittleEndian.Uint16(b)
}

func le32(b []byte) uint32 {
	if len(b) < 4 {
		return 0
	}
	return binary.LittleEndian.Uint32(b)
}

func le64(b []byte) uint64 {
	if len(b) < 8 {
		return 0
	}
	return binary.LittleEndian.Uint64(b)
}
