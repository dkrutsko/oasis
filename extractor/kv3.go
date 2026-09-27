package main

// Binary KeyValues3 decoder. This is a port of the reading side of
// ValveResourceFormat's `BinaryKV3.cs` (MIT License, Copyright (c) 2015
// ValveResourceFormat Contributors). See NOTICE for the license text.
//
// Only versions 4 and 5 with zstd compression are supported, which is
// what the physics of every CS2 map use. Other versions and compression
// methods are rejected with an error.

import (
	"encoding/binary"
	"math"

	"github.com/klauspost/compress/zstd"
	"golang.org/x/sync/errgroup"

	"github.com/dkrutsko/oasis/errors"
)

//----------------------------------------------------------------------------//
// Constants                                                                  //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

const (
	// The magic is "\x003VK" with the version in the low byte
	kv3MagicPrefix = 0x4B563300
	kv3MagicMask   = 0xFFFFFF00

	// Value that ends every KV3 block
	kv3Trailer = 0xFFEEDD00

	// The only compression method used by CS2 map physics
	kv3CompressionZstd = 2

	// Header sizes of the supported versions, counting the magic
	// and the format GUID that start the header
	kv3HeaderSizeV4 = 72
	kv3HeaderSizeV5 = 120

	// Limits that stop a corrupt block from exhausting memory or
	// the stack. Real map physics stay far below them.
	kv3MaxDecodedSize = 1 << 30
	kv3MaxDepth       = 64
	kv3MaxPreallocate = 1 << 16
)

////////////////////////////////////////////////////////////////////////////////

// Value types as stored in the types buffer, matching
// `KV3BinaryNodeType` in VRF
const (
	kv3TypeNull            = 1
	kv3TypeBoolean         = 2
	kv3TypeInt64           = 3
	kv3TypeUint64          = 4
	kv3TypeDouble          = 5
	kv3TypeString          = 6
	kv3TypeBlob            = 7
	kv3TypeArray           = 8
	kv3TypeObject          = 9
	kv3TypeArrayTyped      = 10
	kv3TypeInt32           = 11
	kv3TypeUint32          = 12
	kv3TypeBooleanTrue     = 13
	kv3TypeBooleanFalse    = 14
	kv3TypeInt64Zero       = 15
	kv3TypeInt64One        = 16
	kv3TypeDoubleZero      = 17
	kv3TypeDoubleOne       = 18
	kv3TypeFloat           = 19
	kv3TypeInt16           = 20
	kv3TypeUint16          = 21
	kv3TypeInt32AsByte     = 23
	kv3TypeArrayByteLength = 24
	kv3TypeArrayAuxiliary  = 25

	// A set high bit means a flag byte follows the type. The
	// low six bits hold the type itself.
	kv3TypeFlagged = 0x80
	kv3TypeMask    = 0x3F
)

//----------------------------------------------------------------------------//
// Types                                                                      //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// Kv3Kind is the type of a decoded KV3 value.
type Kv3Kind uint8

const (
	Kv3KindNull Kv3Kind = iota
	Kv3KindBool
	Kv3KindInt
	Kv3KindUint
	Kv3KindFloat
	Kv3KindString
	Kv3KindBlob
	Kv3KindArray
	Kv3KindObject
)

////////////////////////////////////////////////////////////////////////////////

// Kv3Value is a decoded KV3 value. Arrays keep their elements
// and objects their fields, both in order. Blobs point into the
// decompressed data rather than being copied.
type Kv3Value struct {
	Kind Kv3Kind

	number   uint64
	text     string
	blob     []byte
	elements []Kv3Value
	fields   []Kv3Field
}

////////////////////////////////////////////////////////////////////////////////

// Kv3Field is a named value of an object.
type Kv3Field struct {
	Name  string
	Value Kv3Value
}

////////////////////////////////////////////////////////////////////////////////

// kv3Header holds the header fields that locate and split the
// buffers of a KV3 block.
type kv3Header struct {
	version     int
	compression uint32

	dictionaryId uint16
	frameSize    uint16

	countTypes  int
	countBlocks int
	sizeBlobs   int

	sizeUncompressedTotal    int
	sizeCompressedTotal      int
	sizeBlockCompressedSizes int

	// Version 4 has a single buffer. Version 5 adds a second one
	// that holds the object lengths, the main values and types.
	buffer1 kv3BufferHeader
	buffer2 kv3BufferHeader

	countObjectsBuffer2 int
}

////////////////////////////////////////////////////////////////////////////////

// kv3BufferHeader holds the sizes of one value buffer and the
// number of 1, 2, 4 and 8 byte values in it.
type kv3BufferHeader struct {
	sizeCompressed   int
	sizeUncompressed int

	count1 int
	count2 int
	count4 int
	count8 int
}

////////////////////////////////////////////////////////////////////////////////

// kv3Buffers holds the 1, 2, 4 and 8 byte value arrays of one
// value buffer.
type kv3Buffers struct {
	bytes1 binaryReader
	bytes2 binaryReader
	bytes4 binaryReader
	bytes8 binaryReader
}

////////////////////////////////////////////////////////////////////////////////

// kv3Frame is one zstd frame of a KV3 block.
type kv3Frame struct {
	sizeCompressed   int
	sizeUncompressed int

	source []byte
	data   []byte
}

////////////////////////////////////////////////////////////////////////////////

// kv3Decoder holds the state of a single `DecodeKv3` call.
type kv3Decoder struct {
	version int
	keep    map[string]bool
	depth   int

	strings       []string
	types         binaryReader
	objectLengths binaryReader
	blobLengths   binaryReader
	blobs         binaryReader

	// Values are read from `buffer`. Auxiliary arrays read from
	// the other buffer, so the two are swapped while reading them.
	buffer    *kv3Buffers
	auxiliary *kv3Buffers
}

//----------------------------------------------------------------------------//
// Methods                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// GetField returns the value of the named field, or nil when the
// value is nil, is not an object, or has no such field.
func (v *Kv3Value) GetField(name string) *Kv3Value {

	if v == nil {
		return nil
	}

	for i := range v.fields {
		if v.fields[i].Name == name {
			return &v.fields[i].Value
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

// GetElements returns the elements of an array, or nil when the
// value is nil or is not an array.
func (v *Kv3Value) GetElements() []Kv3Value {

	if v == nil {
		return nil
	}

	return v.elements
}

////////////////////////////////////////////////////////////////////////////////

// GetBlob returns the bytes of a blob, or nil for any other value.
func (v *Kv3Value) GetBlob() []byte {

	if v == nil || v.Kind != Kv3KindBlob {
		return nil
	}

	return v.blob
}

////////////////////////////////////////////////////////////////////////////////

// GetString returns the text of a string, or an empty string for
// any other value.
func (v *Kv3Value) GetString() string {

	if v == nil || v.Kind != Kv3KindString {
		return ""
	}

	return v.text
}

////////////////////////////////////////////////////////////////////////////////

// GetInt returns the value of a signed or unsigned integer. The
// second result is false for any other value.
func (v *Kv3Value) GetInt() (int64, bool) {

	if v == nil || (v.Kind != Kv3KindInt && v.Kind != Kv3KindUint) {
		return 0, false
	}

	return int64(v.number), true
}

////////////////////////////////////////////////////////////////////////////////

// GetFloat returns the value of a floating point number. The
// second result is false for any other value.
func (v *Kv3Value) GetFloat() (float64, bool) {

	if v == nil || v.Kind != Kv3KindFloat {
		return 0, false
	}

	return math.Float64frombits(v.number), true
}

//----------------------------------------------------------------------------//
// Decoding                                                                   //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// DecodeKv3 decodes a binary KV3 block. When `keep` is not nil,
// only object fields with those names are stored, which saves
// memory when the caller needs a small part of a large block.
// Array elements are always stored. The zstd decoder must allow
// at least three concurrent `DecodeAll` calls.
func DecodeKv3(data []byte, keep map[string]bool, zd *zstd.Decoder) (*Kv3Value, error) {

	//----------------------------------------------------------------------------//

	if len(data) < 4 {
		return nil, errors.New(
			"kv3 block is too small",
			errors.Int("size", len(data)),
		)
	}

	magic := binary.LittleEndian.Uint32(data[0:4])
	version := int(magic &^ kv3MagicMask)

	if magic&kv3MagicMask != kv3MagicPrefix || (version != 4 && version != 5) {
		return nil, errors.New(
			"unsupported kv3 version",
			errors.Uint32("magic", magic),
		)
	}

	header, err := readKv3Header(data, version)
	if err != nil {
		return nil, errors.New(
			"failed to read kv3 header",
			errors.Error("error", err),
		)
	}

	//----------------------------------------------------------------------------//

	frames, err := decodeKv3Frames(data, header, zd)
	if err != nil {
		return nil, errors.New(
			"failed to decompress kv3 block",
			errors.Error("error", err),
		)
	}

	decoder := &kv3Decoder{
		version: version,
		keep:    keep,
	}

	if version == 4 {
		err = decoder.readVersion4(frames, header)
	} else {
		err = decoder.readVersion5(frames, header)
	}

	if err != nil {
		return nil, errors.New(
			"failed to read kv3 buffers",
			errors.Error("error", err),
		)
	}

	//----------------------------------------------------------------------------//

	// The root value has a type of its own like any other value
	rootType, err := decoder.readType()
	if err != nil {
		return nil, errors.New(
			"failed to read kv3 values",
			errors.Error("error", err),
		)
	}

	root, err := decoder.readValue(rootType, false)
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

// readKv3Header reads the header fields that follow the magic
// and the 16 byte format GUID.
func readKv3Header(data []byte, version int) (kv3Header, error) {

	//----------------------------------------------------------------------------//

	size := kv3HeaderSizeV4
	if version == 5 {
		size = kv3HeaderSizeV5
	}

	if len(data) < size {
		return kv3Header{}, errors.New(
			"kv3 header is truncated",
			errors.Int("size", len(data)),
		)
	}

	header := kv3Header{
		version:     version,
		compression: binary.LittleEndian.Uint32(data[20:24]),

		dictionaryId: binary.LittleEndian.Uint16(data[24:26]),
		frameSize:    binary.LittleEndian.Uint16(data[26:28]),

		countTypes:  readKv3Int32(data, 40),
		countBlocks: readKv3Int32(data, 56),
		sizeBlobs:   readKv3Int32(data, 60),

		sizeUncompressedTotal:    readKv3Int32(data, 48),
		sizeCompressedTotal:      readKv3Int32(data, 52),
		sizeBlockCompressedSizes: readKv3Int32(data, 68),

		buffer1: kv3BufferHeader{
			count1: readKv3Int32(data, 28),
			count2: readKv3Int32(data, 64),
			count4: readKv3Int32(data, 32),
			count8: readKv3Int32(data, 36),
		},
	}

	//----------------------------------------------------------------------------//

	if version == 4 {
		// The single buffer uses the sizes of the whole block
		header.buffer1.sizeCompressed = header.sizeCompressedTotal
		header.buffer1.sizeUncompressed = header.sizeUncompressedTotal

	} else {
		header.buffer1.sizeUncompressed = readKv3Int32(data, 72)
		header.buffer1.sizeCompressed = readKv3Int32(data, 76)

		header.buffer2 = kv3BufferHeader{
			sizeUncompressed: readKv3Int32(data, 80),
			sizeCompressed:   readKv3Int32(data, 84),

			count1: readKv3Int32(data, 88),
			count2: readKv3Int32(data, 92),
			count4: readKv3Int32(data, 96),
			count8: readKv3Int32(data, 100),
		}

		header.countObjectsBuffer2 = readKv3Int32(data, 108)
	}

	//----------------------------------------------------------------------------//

	// Sizes and counts are signed in the format
	sizes := []int{
		header.countTypes,
		header.countBlocks,
		header.sizeBlobs,
		header.sizeUncompressedTotal,
		header.sizeCompressedTotal,
		header.buffer1.sizeCompressed,
		header.buffer1.sizeUncompressed,
		header.buffer1.count1,
		header.buffer1.count2,
		header.buffer1.count4,
		header.buffer1.count8,
		header.buffer2.sizeCompressed,
		header.buffer2.sizeUncompressed,
		header.buffer2.count1,
		header.buffer2.count2,
		header.buffer2.count4,
		header.buffer2.count8,
		header.countObjectsBuffer2,
	}

	for _, value := range sizes {
		if value < 0 {
			return kv3Header{}, errors.New(
				"kv3 header has a negative size",
				errors.Int("value", value),
			)
		}
	}

	//----------------------------------------------------------------------------//

	// VRF rejects dictionaries and frame sizes with zstd as well
	if header.compression != kv3CompressionZstd {
		return kv3Header{}, errors.New(
			"unsupported kv3 compression method",
			errors.Uint32("method", header.compression),
		)
	}

	if header.dictionaryId != 0 || header.frameSize != 0 {
		return kv3Header{}, errors.New(
			"unexpected kv3 zstd parameters",
			errors.Uint16("dictionary_id", header.dictionaryId),
			errors.Uint16("frame_size", header.frameSize),
		)
	}

	return header, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// decodeKv3Frames decompresses the zstd frames that follow the
// header. Version 4 stores its buffer and the blobs in a single
// frame. Version 5 stores both buffers and the blobs in frames
// of their own, which are decoded at the same time.
func decodeKv3Frames(data []byte, header kv3Header, zd *zstd.Decoder) ([]*kv3Frame, error) {

	//----------------------------------------------------------------------------//

	var frames []*kv3Frame

	if header.version == 4 {
		frames = append(frames, &kv3Frame{
			sizeCompressed:   header.buffer1.sizeCompressed,
			sizeUncompressed: header.buffer1.sizeUncompressed + header.sizeBlobs,
		})

	} else {
		frames = append(frames, &kv3Frame{
			sizeCompressed:   header.buffer1.sizeCompressed,
			sizeUncompressed: header.buffer1.sizeUncompressed,
		})

		frames = append(frames, &kv3Frame{
			sizeCompressed:   header.buffer2.sizeCompressed,
			sizeUncompressed: header.buffer2.sizeUncompressed,
		})

		if header.countBlocks > 0 {
			if header.sizeBlockCompressedSizes != 0 {
				return nil, errors.New("unexpected kv3 block compressed sizes")
			}

			// The blob frame takes the rest of the compressed data
			compressed := header.sizeCompressedTotal
			compressed -= header.buffer1.sizeCompressed
			compressed -= header.buffer2.sizeCompressed

			frames = append(frames, &kv3Frame{
				sizeCompressed:   compressed,
				sizeUncompressed: header.sizeBlobs,
			})
		}
	}

	//----------------------------------------------------------------------------//

	size := kv3HeaderSizeV4
	if header.version == 5 {
		size = kv3HeaderSizeV5
	}

	// The frames follow the header back to back
	reader := &binaryReader{data: data, pos: size}

	for _, frame := range frames {
		if frame.sizeUncompressed > kv3MaxDecodedSize {
			return nil, errors.New(
				"kv3 frame is too large",
				errors.Int("size", frame.sizeUncompressed),
			)
		}

		source, err := reader.ReadBytes(frame.sizeCompressed)
		if err != nil {
			return nil, err
		}

		frame.source = source
	}

	// Blocks with blobs end with a trailer after the frames
	if header.countBlocks > 0 {
		err := readKv3Trailer(reader)
		if err != nil {
			return nil, err
		}
	}

	//----------------------------------------------------------------------------//

	var group errgroup.Group

	for _, frame := range frames {
		group.Go(func() error {

			output, err := decodeZstdFrame(zd, frame.source, frame.sizeUncompressed)
			if err != nil {
				return err
			}

			frame.data = output
			return nil
		})
	}

	err := group.Wait()
	if err != nil {
		return nil, err
	}

	return frames, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// readVersion4 splits the single version 4 buffer. It holds the
// value arrays, the strings, the types and the blob lengths. The
// blobs follow it in the same frame.
func (d *kv3Decoder) readVersion4(frames []*kv3Frame, header kv3Header) error {

	//----------------------------------------------------------------------------//

	size := header.buffer1.sizeUncompressed
	reader := &binaryReader{data: frames[0].data[:size]}

	buffer, err := readKv3Arrays(reader, header.buffer1)
	if err != nil {
		return err
	}

	d.buffer = buffer

	//----------------------------------------------------------------------------//

	// The strings start at an 8 byte boundary. The type count
	// covers the strings and the types together.
	reader.Align(8)
	stringsStart := reader.GetPosition()

	d.strings, err = readKv3Strings(&buffer.bytes4, reader)
	if err != nil {
		return err
	}

	typesSize := header.countTypes - (reader.GetPosition() - stringsStart)

	types, err := reader.ReadBytes(typesSize)
	if err != nil {
		return err
	}

	d.types = binaryReader{data: types}

	//----------------------------------------------------------------------------//

	err = d.readBlobLengths(reader, header.countBlocks)
	if err != nil {
		return err
	}

	d.blobs = binaryReader{data: frames[0].data[size:]}
	return nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// readVersion5 splits the two version 5 buffers. The first holds
// the strings and the values of auxiliary arrays. The second
// holds the object lengths, the main values, the types and the
// blob lengths. The blobs have a frame of their own.
func (d *kv3Decoder) readVersion5(frames []*kv3Frame, header kv3Header) error {

	//----------------------------------------------------------------------------//

	reader1 := &binaryReader{data: frames[0].data}

	auxiliary, err := readKv3Arrays(reader1, header.buffer1)
	if err != nil {
		return err
	}

	// The strings come first in the 1 byte values
	d.strings, err = readKv3Strings(&auxiliary.bytes4, &auxiliary.bytes1)
	if err != nil {
		return err
	}

	d.auxiliary = auxiliary

	//----------------------------------------------------------------------------//

	reader2 := &binaryReader{data: frames[1].data}

	objectLengths, err := reader2.ReadBytes(header.countObjectsBuffer2 * 4)
	if err != nil {
		return err
	}

	d.objectLengths = binaryReader{data: objectLengths}

	d.buffer, err = readKv3Arrays(reader2, header.buffer2)
	if err != nil {
		return err
	}

	types, err := reader2.ReadBytes(header.countTypes)
	if err != nil {
		return err
	}

	d.types = binaryReader{data: types}

	//----------------------------------------------------------------------------//

	err = d.readBlobLengths(reader2, header.countBlocks)
	if err != nil {
		return err
	}

	if len(frames) > 2 {
		d.blobs = binaryReader{data: frames[2].data}
	}

	return nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// readBlobLengths reads what follows the types, which is the
// blob lengths, if there are blobs, and then the trailer.
func (d *kv3Decoder) readBlobLengths(reader *binaryReader, countBlocks int) error {

	if countBlocks > 0 {
		lengths, err := reader.ReadBytes(countBlocks * 4)
		if err != nil {
			return err
		}

		d.blobLengths = binaryReader{data: lengths}
	}

	return readKv3Trailer(reader)
}

//----------------------------------------------------------------------------//
// Values                                                                     //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// readType returns the type of the next value. Flag bytes only
// annotate values, marking resource names and such, so they are
// skipped.
func (d *kv3Decoder) readType() (byte, error) {

	nodeType, err := d.types.ReadUint8()
	if err != nil {
		return 0, err
	}

	if nodeType&kv3TypeFlagged != 0 {
		nodeType &= kv3TypeMask

		_, err = d.types.ReadUint8()
		if err != nil {
			return 0, err
		}
	}

	return nodeType, nil
}

////////////////////////////////////////////////////////////////////////////////

// readValue reads one value of the given type. When `discard` is
// set, containers are walked without storing their contents,
// since the caller throws the value away.
func (d *kv3Decoder) readValue(nodeType byte, discard bool) (Kv3Value, error) {

	switch nodeType {

	case kv3TypeString:
		index, err := d.buffer.bytes4.ReadInt32()
		if err != nil {
			return Kv3Value{}, err
		}

		text, err := d.getString(index)
		if err != nil {
			return Kv3Value{}, err
		}

		return Kv3Value{Kind: Kv3KindString, text: text}, nil

	case kv3TypeBlob:
		return d.readBlob()

	case kv3TypeArray:
		length, err := d.buffer.bytes4.ReadInt32()
		if err != nil {
			return Kv3Value{}, err
		}

		return d.readArray(int(length), discard)

	case kv3TypeArrayTyped:
		length, err := d.buffer.bytes4.ReadInt32()
		if err != nil {
			return Kv3Value{}, err
		}

		return d.readTypedArray(int(length), false, discard)

	case kv3TypeArrayByteLength:
		length, err := d.buffer.bytes1.ReadUint8()
		if err != nil {
			return Kv3Value{}, err
		}

		return d.readTypedArray(int(length), false, discard)

	case kv3TypeArrayAuxiliary:
		length, err := d.buffer.bytes1.ReadUint8()
		if err != nil {
			return Kv3Value{}, err
		}

		return d.readTypedArray(int(length), true, discard)

	case kv3TypeObject:
		// Version 5 keeps object lengths in a list of their own
		var length int32
		var err error

		if d.version == 5 {
			length, err = d.objectLengths.ReadInt32()
		} else {
			length, err = d.buffer.bytes4.ReadInt32()
		}

		if err != nil {
			return Kv3Value{}, err
		}

		return d.readObject(int(length), discard)

	default:
		return d.readScalar(nodeType)
	}
}

////////////////////////////////////////////////////////////////////////////////

// readScalar reads a null, a boolean or a number. Some values
// are stored in the type alone, the rest in the value array of
// their size.
func (d *kv3Decoder) readScalar(nodeType byte) (Kv3Value, error) {

	switch nodeType {

	case kv3TypeNull:
		return Kv3Value{Kind: Kv3KindNull}, nil

	case kv3TypeBooleanTrue:
		return Kv3Value{Kind: Kv3KindBool, number: 1}, nil

	case kv3TypeBooleanFalse:
		return Kv3Value{Kind: Kv3KindBool}, nil

	case kv3TypeInt64Zero:
		return Kv3Value{Kind: Kv3KindInt}, nil

	case kv3TypeInt64One:
		return Kv3Value{Kind: Kv3KindInt, number: 1}, nil

	case kv3TypeDoubleZero:
		return Kv3Value{Kind: Kv3KindFloat, number: math.Float64bits(0)}, nil

	case kv3TypeDoubleOne:
		return Kv3Value{Kind: Kv3KindFloat, number: math.Float64bits(1)}, nil

	case kv3TypeBoolean:
		value, err := d.buffer.bytes1.ReadUint8()
		if err != nil {
			return Kv3Value{}, err
		}

		result := Kv3Value{Kind: Kv3KindBool}
		if value == 1 {
			result.number = 1
		}

		return result, nil

	case kv3TypeInt32AsByte:
		value, err := d.buffer.bytes1.ReadUint8()
		if err != nil {
			return Kv3Value{}, err
		}

		return Kv3Value{Kind: Kv3KindInt, number: uint64(value)}, nil

	case kv3TypeInt16:
		value, err := d.buffer.bytes2.ReadUint16()
		if err != nil {
			return Kv3Value{}, err
		}

		return Kv3Value{Kind: Kv3KindInt, number: uint64(int64(int16(value)))}, nil

	case kv3TypeUint16:
		value, err := d.buffer.bytes2.ReadUint16()
		if err != nil {
			return Kv3Value{}, err
		}

		return Kv3Value{Kind: Kv3KindUint, number: uint64(value)}, nil

	case kv3TypeInt32:
		value, err := d.buffer.bytes4.ReadUint32()
		if err != nil {
			return Kv3Value{}, err
		}

		return Kv3Value{Kind: Kv3KindInt, number: uint64(int64(int32(value)))}, nil

	case kv3TypeUint32:
		value, err := d.buffer.bytes4.ReadUint32()
		if err != nil {
			return Kv3Value{}, err
		}

		return Kv3Value{Kind: Kv3KindUint, number: uint64(value)}, nil

	case kv3TypeFloat:
		value, err := d.buffer.bytes4.ReadUint32()
		if err != nil {
			return Kv3Value{}, err
		}

		number := float64(math.Float32frombits(value))
		return Kv3Value{Kind: Kv3KindFloat, number: math.Float64bits(number)}, nil

	case kv3TypeInt64:
		value, err := d.buffer.bytes8.ReadUint64()
		if err != nil {
			return Kv3Value{}, err
		}

		return Kv3Value{Kind: Kv3KindInt, number: value}, nil

	case kv3TypeUint64:
		value, err := d.buffer.bytes8.ReadUint64()
		if err != nil {
			return Kv3Value{}, err
		}

		return Kv3Value{Kind: Kv3KindUint, number: value}, nil

	case kv3TypeDouble:
		value, err := d.buffer.bytes8.ReadUint64()
		if err != nil {
			return Kv3Value{}, err
		}

		return Kv3Value{Kind: Kv3KindFloat, number: value}, nil

	default:
		return Kv3Value{}, errors.New(
			"unknown kv3 type",
			errors.Uint8("type", nodeType),
		)
	}
}

////////////////////////////////////////////////////////////////////////////////

// readBlob reads a blob. Blob lengths are in a list of their own
// and the bytes are in the blob data.
func (d *kv3Decoder) readBlob() (Kv3Value, error) {

	length, err := d.blobLengths.ReadInt32()
	if err != nil {
		return Kv3Value{}, err
	}

	// Lengths of zero or less are empty blobs, as in VRF
	if length <= 0 {
		return Kv3Value{Kind: Kv3KindBlob, blob: []byte{}}, nil
	}

	data, err := d.blobs.ReadBytes(int(length))
	if err != nil {
		return Kv3Value{}, err
	}

	return Kv3Value{Kind: Kv3KindBlob, blob: data}, nil
}

////////////////////////////////////////////////////////////////////////////////

// readArray reads an array whose elements each have their own
// type.
func (d *kv3Decoder) readArray(length int, discard bool) (Kv3Value, error) {

	err := d.enterContainer(length)
	if err != nil {
		return Kv3Value{}, err
	}
	defer d.leaveContainer()

	result := Kv3Value{Kind: Kv3KindArray}

	// Every element has a type byte, which bounds the allocation
	if !discard {
		result.elements = make([]Kv3Value, 0, min(length, d.types.GetRemaining()))
	}

	for i := 0; i < length; i++ {
		nodeType, err := d.readType()
		if err != nil {
			return Kv3Value{}, err
		}

		element, err := d.readValue(nodeType, discard)
		if err != nil {
			return Kv3Value{}, err
		}

		if !discard {
			result.elements = append(result.elements, element)
		}
	}

	return result, nil
}

////////////////////////////////////////////////////////////////////////////////

// readTypedArray reads an array whose elements share one type.
// Auxiliary arrays read their elements from the other buffer.
func (d *kv3Decoder) readTypedArray(length int, auxiliary, discard bool) (Kv3Value, error) {

	//----------------------------------------------------------------------------//

	err := d.enterContainer(length)
	if err != nil {
		return Kv3Value{}, err
	}
	defer d.leaveContainer()

	nodeType, err := d.readType()
	if err != nil {
		return Kv3Value{}, err
	}

	if auxiliary {
		if d.auxiliary == nil {
			return Kv3Value{}, errors.New("kv3 auxiliary array without an auxiliary buffer")
		}

		d.buffer, d.auxiliary = d.auxiliary, d.buffer
		defer func() { d.buffer, d.auxiliary = d.auxiliary, d.buffer }()
	}

	//----------------------------------------------------------------------------//

	result := Kv3Value{Kind: Kv3KindArray}

	// Elements need not use any bytes, so the length alone
	// cannot be trusted for the allocation
	if !discard {
		result.elements = make([]Kv3Value, 0, min(length, kv3MaxPreallocate))
	}

	for i := 0; i < length; i++ {
		element, err := d.readValue(nodeType, discard)
		if err != nil {
			return Kv3Value{}, err
		}

		if !discard {
			result.elements = append(result.elements, element)
		}
	}

	return result, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// readObject reads an object. Each field has its own type and a
// string table index for its name. Fields that are not in `keep`
// are read but not stored.
func (d *kv3Decoder) readObject(length int, discard bool) (Kv3Value, error) {

	//----------------------------------------------------------------------------//

	err := d.enterContainer(length)
	if err != nil {
		return Kv3Value{}, err
	}
	defer d.leaveContainer()

	result := Kv3Value{Kind: Kv3KindObject}

	// Every field has a type byte, which bounds the allocation
	if !discard {
		result.fields = make([]Kv3Field, 0, min(length, d.types.GetRemaining()))
	}

	//----------------------------------------------------------------------------//

	for i := 0; i < length; i++ {
		nodeType, err := d.readType()
		if err != nil {
			return Kv3Value{}, err
		}

		index, err := d.buffer.bytes4.ReadInt32()
		if err != nil {
			return Kv3Value{}, err
		}

		name, err := d.getString(index)
		if err != nil {
			return Kv3Value{}, err
		}

		skip := discard || (d.keep != nil && !d.keep[name])

		value, err := d.readValue(nodeType, skip)
		if err != nil {
			return Kv3Value{}, err
		}

		if !skip {
			result.fields = append(result.fields, Kv3Field{Name: name, Value: value})
		}
	}

	return result, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// enterContainer checks the length of an array or object and
// how deeply it is nested. Each call must be paired with a call
// to `leaveContainer`.
func (d *kv3Decoder) enterContainer(length int) error {

	if length < 0 {
		return errors.New(
			"kv3 container length is negative",
			errors.Int("length", length),
		)
	}

	if d.depth >= kv3MaxDepth {
		return errors.New(
			"kv3 values are nested too deeply",
			errors.Int("depth", d.depth),
		)
	}

	d.depth++
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func (d *kv3Decoder) leaveContainer() {
	d.depth--
}

////////////////////////////////////////////////////////////////////////////////

// getString returns an entry of the string table. An index of
// -1 is an empty string.
func (d *kv3Decoder) getString(index int32) (string, error) {

	if index == -1 {
		return "", nil
	}

	if index < 0 || int(index) >= len(d.strings) {
		return "", errors.New(
			"kv3 string index is out of range",
			errors.Int32("index", index),
		)
	}

	return d.strings[index], nil
}

//----------------------------------------------------------------------------//
// Helpers                                                                    //
//----------------------------------------------------------------------------//

////////////////////////////////////////////////////////////////////////////////

// readKv3Int32 reads a signed header field. The caller checks
// the header is long enough.
func readKv3Int32(data []byte, offset int) int {
	return int(int32(binary.LittleEndian.Uint32(data[offset : offset+4])))
}

////////////////////////////////////////////////////////////////////////////////

// readKv3Arrays slices the 1, 2, 4 and 8 byte value arrays out of
// a buffer, in that order.
func readKv3Arrays(reader *binaryReader, header kv3BufferHeader) (*kv3Buffers, error) {

	var err error
	result := &kv3Buffers{}

	result.bytes1, err = readKv3Array(reader, header.count1, 1)
	if err != nil {
		return nil, err
	}

	result.bytes2, err = readKv3Array(reader, header.count2, 2)
	if err != nil {
		return nil, err
	}

	result.bytes4, err = readKv3Array(reader, header.count4, 4)
	if err != nil {
		return nil, err
	}

	result.bytes8, err = readKv3Array(reader, header.count8, 8)
	if err != nil {
		return nil, err
	}

	return result, nil
}

////////////////////////////////////////////////////////////////////////////////

// readKv3Array reads `count` values of `size` bytes each. The
// array starts at a multiple of the value size, but empty arrays
// are not aligned.
func readKv3Array(reader *binaryReader, count, size int) (binaryReader, error) {

	if count == 0 {
		return binaryReader{}, nil
	}

	reader.Align(size)

	data, err := reader.ReadBytes(count * size)
	if err != nil {
		return binaryReader{}, err
	}

	return binaryReader{data: data}, nil
}

////////////////////////////////////////////////////////////////////////////////

// readKv3Strings reads the string table. The number of strings
// is the first 4 byte value, and the strings themselves are null
// terminated.
func readKv3Strings(bytes4, source *binaryReader) ([]string, error) {

	count, err := bytes4.ReadInt32()
	if err != nil {
		return nil, err
	}

	// Every string takes at least its terminator, which bounds
	// the allocation
	if count < 0 || int(count) > source.GetRemaining() {
		return nil, errors.New(
			"invalid kv3 string count",
			errors.Int32("count", count),
		)
	}

	result := make([]string, count)

	for i := range result {
		result[i], err = source.ReadString()
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

////////////////////////////////////////////////////////////////////////////////

func readKv3Trailer(reader *binaryReader) error {

	trailer, err := reader.ReadUint32()
	if err != nil {
		return err
	}

	if trailer != kv3Trailer {
		return errors.New(
			"unexpected kv3 trailer",
			errors.Uint32("trailer", trailer),
		)
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

func decodeZstdFrame(zd *zstd.Decoder, source []byte, size int) ([]byte, error) {

	output, err := zd.DecodeAll(source, make([]byte, 0, size))
	if err != nil {
		return nil, errors.New(
			"failed to decompress zstd frame",
			errors.Error("error", err),
		)
	}

	if len(output) != size {
		return nil, errors.New(
			"unexpected zstd decoded size",
			errors.Int("decoded", len(output)),
			errors.Int("expected", size),
		)
	}

	return output, nil
}
