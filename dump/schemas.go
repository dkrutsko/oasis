package dump

import (
	"context"
	"encoding/binary"
	"strings"
	"unicode"

	"github.com/dkrutsko/oasis/errors"
	"github.com/dkrutsko/oasis/leech"
)

////////////////////////////////////////////////////////////////////////////////

// Pattern for the schema system global in schemasystem.dll, from
// `src/analysis/schemas.rs` in cs2-dumper
const schemaSystemPattern = "4c8d35${'} 0f2845"

////////////////////////////////////////////////////////////////////////////////

// Layout of the schema system types, from `src/source2` in cs2-dumper
const (
	pointerSize = 0x08

	schemaSystemTypeScopes        = 0x190
	schemaSystemRegistrationCount = 0x280

	utlVectorCount = 0x00
	utlVectorData  = 0x08
	utlVectorSize  = 0x10

	typeScopeName          = 0x08
	typeScopeNameSize      = 256
	typeScopeClassBindings = 0x560
	typeScopeEnumBindings  = 0x1DD0
	typeScopeSize          = 0x3640

	tsHashBlocksAllocated        = 0x0C
	tsHashPeakAllocated          = 0x10
	tsHashFreeBlocks             = 0x20
	tsHashBuckets                = 0x60
	tsHashBucketCount            = 256
	tsHashBucketSize             = 0x18
	tsHashBucketFirstUncommitted = 0x10
	tsHashSize                   = 0x1870

	tsHashNodeNext = 0x08
	tsHashNodeData = 0x10
	tsHashNodeSize = 0x18

	tsHashBlobNext = 0x00
	tsHashBlobData = 0x10
	tsHashBlobSize = 0x30

	classInfoName          = 0x08
	classInfoFieldCount    = 0x24
	classInfoMetadataCount = 0x26
	classInfoFields        = 0x30
	classInfoBaseClasses   = 0x40
	classInfoMetadata      = 0x48
	classInfoSize          = 0x70

	baseClassInfoClass = 0x18
	baseClassName      = 0x10

	classFieldName   = 0x00
	classFieldType   = 0x08
	classFieldOffset = 0x10
	classFieldSize   = 0x20

	metadataEntryName  = 0x00
	metadataEntryValue = 0x08
	metadataEntrySize  = 0x10

	varNameName     = 0x00
	varNameTypeName = 0x08
	varNameSize     = 0x10

	enumInfoName            = 0x08
	enumInfoAlignment       = 0x19
	enumInfoEnumeratorCount = 0x1C
	enumInfoEnumerators     = 0x20
	enumInfoSize            = 0x48

	enumeratorName  = 0x00
	enumeratorValue = 0x08
	enumeratorSize  = 0x20

	schemaStringSize = 128
)

////////////////////////////////////////////////////////////////////////////////

// schemaModule is the JSON that cs2-dumper writes for one module. serde_json
// sorts object keys, so the fields of these structs are declared in
// alphabetical order to produce the same output.
type schemaModule struct {
	Classes map[string]schemaClass `json:"classes"`
	Enums   map[string]schemaEnum  `json:"enums"`
}

////////////////////////////////////////////////////////////////////////////////

type schemaClass struct {
	Fields   map[string]int32 `json:"fields"`
	Metadata []schemaMetadata `json:"metadata"`
	Parent   *string          `json:"parent"`
}

////////////////////////////////////////////////////////////////////////////////

type schemaMetadata struct {
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	TypeName *string `json:"type_name,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////

type schemaEnum struct {
	Alignment uint8            `json:"alignment"`
	Members   map[string]int64 `json:"members"`
	Type      string           `json:"type"`
}

////////////////////////////////////////////////////////////////////////////////

// dumpSchemas walks the schema system and returns the classes and enums of
// the type scope named `scopeName`. Unlike upstream, the other type scopes
// are skipped to save DMA reads.
func dumpSchemas(
	ctx context.Context,
	memory *leech.Memory,
	modules []*leech.Module,
	scopeName string,
) (*schemaModule, error) {

	//----------------------------------------------------------------------------//

	module := findModule(modules, "schemasystem.dll")
	if module == nil {
		return nil, errors.New(
			"failed to find module",
			errors.String("name", "schemasystem.dll"),
		)
	}

	img, err := readImage(memory, module)
	if err != nil {
		return nil, errors.New(
			"failed to read module image",
			errors.Error("error", err),
		)
	}

	save, err := img.find(schemaSystemPattern)
	if err != nil {
		return nil, errors.New(
			"failed to find schema system",
			errors.Error("error", err),
		)
	}
	if save == nil {
		return nil, errors.New("schema system pattern is outdated")
	}

	schemaSystem := module.GetBase() + uintptr(save[1])

	//----------------------------------------------------------------------------//

	registrations, err := memory.ReadInt32(schemaSystem + schemaSystemRegistrationCount)
	if err != nil {
		return nil, errors.New(
			"failed to read schema registrations",
			errors.Error("error", err),
		)
	}
	if registrations == 0 {
		return nil, errors.New("schema system has no registrations")
	}

	scopes, err := memory.ReadData(schemaSystem+schemaSystemTypeScopes, utlVectorSize)
	if err != nil {
		return nil, errors.New(
			"failed to read type scopes",
			errors.Error("error", err),
		)
	}

	scopeCount := int32(binary.LittleEndian.Uint32(scopes[utlVectorCount:]))
	scopeData := uintptr(binary.LittleEndian.Uint64(scopes[utlVectorData:]))

	//----------------------------------------------------------------------------//

	var result *schemaModule

	for i := int32(0); i < scopeCount; i++ {

		if ctx.Err() != nil {
			return nil, errors.New(
				"failed to read type scopes",
				errors.Error("error", ctx.Err()),
			)
		}

		scope, err := memory.ReadPtr(scopeData + uintptr(i)*pointerSize)
		if err != nil {
			return nil, errors.New(
				"failed to read type scope",
				errors.Int("index", int(i)),
				errors.Error("error", err),
			)
		}

		name, err := memory.ReadString(scope+typeScopeName, typeScopeNameSize)
		if err != nil {
			return nil, errors.New(
				"failed to read type scope name",
				errors.Int("index", int(i)),
				errors.Error("error", err),
			)
		}

		if name != scopeName {
			continue
		}

		found, err := readTypeScope(ctx, memory, scope)
		if err != nil {
			return nil, errors.New(
				"failed to read type scope",
				errors.String("name", name),
				errors.Error("error", err),
			)
		}

		// Upstream drops empty scopes and keeps the last scope per name
		if len(found.Classes) > 0 || len(found.Enums) > 0 {
			result = found
		}
	}

	if result == nil {
		return nil, errors.New(
			"failed to find type scope",
			errors.String("name", scopeName),
		)
	}

	//----------------------------------------------------------------------------//

	return result, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

func readTypeScope(
	ctx context.Context,
	memory *leech.Memory,
	scope uintptr,
) (*schemaModule, error) {

	//----------------------------------------------------------------------------//

	data, err := memory.ReadData(scope, typeScopeSize)
	if err != nil {
		return nil, errors.New(
			"failed to read type scope",
			errors.Error("error", err),
		)
	}

	classHash := data[typeScopeClassBindings : typeScopeClassBindings+tsHashSize]
	enumHash := data[typeScopeEnumBindings : typeScopeEnumBindings+tsHashSize]

	result := &schemaModule{
		Classes: make(map[string]schemaClass),
		Enums:   make(map[string]schemaEnum),
	}

	//----------------------------------------------------------------------------//

	// Bindings that fail to read are skipped like upstream. Names that
	// collide after `slugify` keep the last binding, also like upstream.
	for _, binding := range readHashElements(memory, classHash) {

		if ctx.Err() != nil {
			return nil, errors.New(
				"failed to read class bindings",
				errors.Error("error", ctx.Err()),
			)
		}

		name, class, err := readClass(memory, binding)
		if err != nil {
			continue
		}

		result.Classes[slugify(name)] = class
	}

	for _, binding := range readHashElements(memory, enumHash) {

		if ctx.Err() != nil {
			return nil, errors.New(
				"failed to read enum bindings",
				errors.Error("error", ctx.Err()),
			)
		}

		name, enum, err := readEnum(memory, binding)
		if err != nil {
			continue
		}

		result.Enums[slugify(name)] = enum
	}

	//----------------------------------------------------------------------------//

	return result, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// readHashElements returns the bindings stored in a `CUtlTSHash`, given the
// bytes of the hash itself. Like upstream, it collects the committed nodes
// of every bucket and then the free list of the memory pool, and drops the
// duplicates. A read that fails ends the list it was walking.
func readHashElements(memory *leech.Memory, hash []byte) []uintptr {

	//----------------------------------------------------------------------------//

	// The counts are signed upstream and converted to usize, so a negative
	// count never stops a list
	usedCount := uint64(int32(binary.LittleEndian.Uint32(hash[tsHashBlocksAllocated:])))
	freeCount := uint64(int32(binary.LittleEndian.Uint32(hash[tsHashPeakAllocated:])))

	//----------------------------------------------------------------------------//

	var allocated []uintptr

	for bucket := 0; bucket < tsHashBucketCount; bucket++ {

		offset := tsHashBuckets + bucket*tsHashBucketSize + tsHashBucketFirstUncommitted
		node := uintptr(binary.LittleEndian.Uint64(hash[offset:]))

		// Upstream only leaves the current bucket once the count is
		// reached, so later buckets still add one element each
		for node != 0 {

			data, err := memory.ReadData(node, tsHashNodeSize)
			if err != nil {
				break
			}

			value := uintptr(binary.LittleEndian.Uint64(data[tsHashNodeData:]))
			if value != 0 {
				allocated = append(allocated, value)
			}

			if uint64(len(allocated)) >= usedCount {
				break
			}

			node = uintptr(binary.LittleEndian.Uint64(data[tsHashNodeNext:]))
		}
	}

	//----------------------------------------------------------------------------//

	var unallocated []uintptr
	blob := uintptr(binary.LittleEndian.Uint64(hash[tsHashFreeBlocks:]))

	for blob != 0 {

		data, err := memory.ReadData(blob, tsHashBlobSize)
		if err != nil {
			break
		}

		value := uintptr(binary.LittleEndian.Uint64(data[tsHashBlobData:]))
		if value != 0 {
			unallocated = append(unallocated, value)
		}

		if uint64(len(unallocated)) >= freeCount {
			break
		}

		blob = uintptr(binary.LittleEndian.Uint64(data[tsHashBlobNext:]))
	}

	//----------------------------------------------------------------------------//

	elements := append(allocated, unallocated...)
	seen := make(map[uintptr]bool, len(elements))
	result := make([]uintptr, 0, len(elements))

	for _, element := range elements {
		if seen[element] {
			continue
		}

		seen[element] = true
		result = append(result, element)
	}

	//----------------------------------------------------------------------------//

	return result

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

func readClass(memory *leech.Memory, binding uintptr) (string, schemaClass, error) {

	//----------------------------------------------------------------------------//

	data, err := memory.ReadData(binding, classInfoSize)
	if err != nil {
		return "", schemaClass{}, errors.New(
			"failed to read class",
			errors.Uintptr("address", binding),
			errors.Error("error", err),
		)
	}

	name, err := readSchemaString(memory, data[classInfoName:])
	if err != nil {
		return "", schemaClass{}, errors.New(
			"failed to read class name",
			errors.Uintptr("address", binding),
			errors.Error("error", err),
		)
	}
	if name == "" {
		return "", schemaClass{}, errors.New(
			"class name is empty",
			errors.Uintptr("address", binding),
		)
	}

	//----------------------------------------------------------------------------//

	fieldCount := int16(binary.LittleEndian.Uint16(data[classInfoFieldCount:]))
	fieldData := uintptr(binary.LittleEndian.Uint64(data[classInfoFields:]))

	fields, err := readClassFields(memory, fieldData, fieldCount)
	if err != nil {
		return "", schemaClass{}, errors.New(
			"failed to read class fields",
			errors.String("name", name),
			errors.Error("error", err),
		)
	}

	metadataCount := int16(binary.LittleEndian.Uint16(data[classInfoMetadataCount:]))
	metadataData := uintptr(binary.LittleEndian.Uint64(data[classInfoMetadata:]))

	metadata, err := readClassMetadata(memory, metadataData, metadataCount)
	if err != nil {
		return "", schemaClass{}, errors.New(
			"failed to read class metadata",
			errors.String("name", name),
			errors.Error("error", err),
		)
	}

	baseClasses := uintptr(binary.LittleEndian.Uint64(data[classInfoBaseClasses:]))

	//----------------------------------------------------------------------------//

	return name, schemaClass{
		Fields:   fields,
		Metadata: metadata,
		Parent:   readParent(memory, baseClasses),
	}, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// readParent returns the name of the first base class. Returns nil when
// there is no base class or any read fails, like upstream.
func readParent(memory *leech.Memory, baseClasses uintptr) *string {

	if baseClasses == 0 {
		return nil
	}

	class, err := memory.ReadPtr(baseClasses + baseClassInfoClass)
	if err != nil {
		return nil
	}

	address, err := memory.ReadPtr(class + baseClassName)
	if err != nil {
		return nil
	}

	name, err := memory.ReadString(address, schemaStringSize)
	if err != nil || name == "" {
		return nil
	}

	return &name
}

////////////////////////////////////////////////////////////////////////////////

// readClassFields returns the offset of every field by name. Upstream also
// reads the field type names, which the JSON output does not use, so they
// are skipped here.
func readClassFields(memory *leech.Memory, fields uintptr, count int16) (map[string]int32, error) {

	result := make(map[string]int32)

	if fields == 0 {
		return result, nil
	}

	for i := int16(0); i < count; i++ {

		data, err := memory.ReadData(fields+uintptr(i)*classFieldSize, classFieldSize)
		if err != nil {
			return nil, errors.New(
				"failed to read class field",
				errors.Int("index", int(i)),
				errors.Error("error", err),
			)
		}

		if binary.LittleEndian.Uint64(data[classFieldType:]) == 0 {
			continue
		}

		name, err := readSchemaString(memory, data[classFieldName:])
		if err != nil {
			return nil, errors.New(
				"failed to read class field name",
				errors.Int("index", int(i)),
				errors.Error("error", err),
			)
		}

		result[name] = int32(binary.LittleEndian.Uint32(data[classFieldOffset:]))
	}

	return result, nil
}

////////////////////////////////////////////////////////////////////////////////

func readClassMetadata(
	memory *leech.Memory,
	metadata uintptr,
	count int16,
) ([]schemaMetadata, error) {

	result := []schemaMetadata{}

	if metadata == 0 {
		return result, nil
	}

	for i := int16(0); i < count; i++ {

		entry, err := readMetadataEntry(memory, metadata+uintptr(i)*metadataEntrySize)
		if err != nil {
			return nil, errors.New(
				"failed to read metadata entry",
				errors.Int("index", int(i)),
				errors.Error("error", err),
			)
		}

		if entry != nil {
			result = append(result, *entry)
		}
	}

	return result, nil
}

////////////////////////////////////////////////////////////////////////////////

// readMetadataEntry reads one static metadata entry. Returns nil for an
// entry without a value, which upstream skips.
func readMetadataEntry(memory *leech.Memory, address uintptr) (*schemaMetadata, error) {

	//----------------------------------------------------------------------------//

	data, err := memory.ReadData(address, metadataEntrySize)
	if err != nil {
		return nil, errors.New(
			"failed to read metadata entry",
			errors.Uintptr("address", address),
			errors.Error("error", err),
		)
	}

	value := uintptr(binary.LittleEndian.Uint64(data[metadataEntryValue:]))
	if value == 0 {
		return nil, nil
	}

	name, err := readSchemaString(memory, data[metadataEntryName:])
	if err != nil {
		return nil, errors.New(
			"failed to read metadata name",
			errors.Uintptr("address", address),
			errors.Error("error", err),
		)
	}

	// Upstream reads the value of every entry, so a failure drops the
	// class even when the value is not used
	valueData, err := memory.ReadData(value, varNameSize)
	if err != nil {
		return nil, errors.New(
			"failed to read metadata value",
			errors.String("name", name),
			errors.Error("error", err),
		)
	}

	//----------------------------------------------------------------------------//

	switch name {
	case "MNetworkChangeCallback":
		callback, err := readSchemaString(memory, valueData)
		if err != nil {
			return nil, errors.New(
				"failed to read change callback name",
				errors.Error("error", err),
			)
		}

		return &schemaMetadata{Name: callback, Type: "NetworkChangeCallback"}, nil

	case "MNetworkVarNames":
		varName, err := readSchemaString(memory, valueData[varNameName:])
		if err != nil {
			return nil, errors.New(
				"failed to read network var name",
				errors.Error("error", err),
			)
		}

		typeName, err := readSchemaString(memory, valueData[varNameTypeName:])
		if err != nil {
			return nil, errors.New(
				"failed to read network var type name",
				errors.Error("error", err),
			)
		}

		typeName = strings.ReplaceAll(typeName, " ", "")
		return &schemaMetadata{Name: varName, Type: "NetworkVarNames", TypeName: &typeName}, nil
	}

	//----------------------------------------------------------------------------//

	return &schemaMetadata{Name: name, Type: "Unknown"}, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

func readEnum(memory *leech.Memory, binding uintptr) (string, schemaEnum, error) {

	//----------------------------------------------------------------------------//

	data, err := memory.ReadData(binding, enumInfoSize)
	if err != nil {
		return "", schemaEnum{}, errors.New(
			"failed to read enum",
			errors.Uintptr("address", binding),
			errors.Error("error", err),
		)
	}

	name, err := readSchemaString(memory, data[enumInfoName:])
	if err != nil {
		return "", schemaEnum{}, errors.New(
			"failed to read enum name",
			errors.Uintptr("address", binding),
			errors.Error("error", err),
		)
	}
	if name == "" {
		return "", schemaEnum{}, errors.New(
			"enum name is empty",
			errors.Uintptr("address", binding),
		)
	}

	//----------------------------------------------------------------------------//

	count := binary.LittleEndian.Uint16(data[enumInfoEnumeratorCount:])
	enumerators := uintptr(binary.LittleEndian.Uint64(data[enumInfoEnumerators:]))

	members, err := readEnumMembers(memory, enumerators, count)
	if err != nil {
		return "", schemaEnum{}, errors.New(
			"failed to read enum members",
			errors.String("name", name),
			errors.Error("error", err),
		)
	}

	alignment := data[enumInfoAlignment]

	var enumType string
	switch alignment {
	case 1:
		enumType = "uint8"

	case 2:
		enumType = "uint16"

	case 4:
		enumType = "uint32"

	case 8:
		enumType = "uint64"

	default:
		enumType = "unknown"
	}

	//----------------------------------------------------------------------------//

	return name, schemaEnum{
		Alignment: alignment,
		Members:   members,
		Type:      enumType,
	}, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

func readEnumMembers(
	memory *leech.Memory,
	enumerators uintptr,
	count uint16,
) (map[string]int64, error) {

	result := make(map[string]int64)

	if enumerators == 0 {
		return result, nil
	}

	for i := uint16(0); i < count; i++ {

		data, err := memory.ReadData(enumerators+uintptr(i)*enumeratorSize, enumeratorSize)
		if err != nil {
			return nil, errors.New(
				"failed to read enum member",
				errors.Int("index", int(i)),
				errors.Error("error", err),
			)
		}

		name, err := readSchemaString(memory, data[enumeratorName:])
		if err != nil {
			return nil, errors.New(
				"failed to read enum member name",
				errors.Int("index", int(i)),
				errors.Error("error", err),
			)
		}

		result[name] = int64(binary.LittleEndian.Uint64(data[enumeratorValue:]))
	}

	return result, nil
}

////////////////////////////////////////////////////////////////////////////////

// readSchemaString reads the string that the pointer at the start of `data`
// points to.
func readSchemaString(memory *leech.Memory, data []byte) (string, error) {

	address := uintptr(binary.LittleEndian.Uint64(data))

	result, err := memory.ReadString(address, schemaStringSize)
	if err != nil {
		return "", errors.New(
			"failed to read string",
			errors.Uintptr("address", address),
			errors.Error("error", err),
		)
	}

	return result, nil
}

////////////////////////////////////////////////////////////////////////////////

// slugify matches cs2-dumper's `slugify`, which turns every character that
// is not alphanumeric into an underscore.
func slugify(name string) string {

	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return r
		}

		return '_'
	}, name)
}
