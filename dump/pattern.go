package dump

import (
	"bytes"
	"debug/pe"
	"encoding/binary"
	"strconv"
	"strings"

	"github.com/dkrutsko/oasis/errors"
	"github.com/dkrutsko/oasis/leech"
)

////////////////////////////////////////////////////////////////////////////////

type atomType uint8

const (
	atomByte atomType = iota
	atomSave
	atomPush
	atomPop
	atomSkip
	atomJump4
	atomReadU8
	atomReadU16
	atomReadU32
)

////////////////////////////////////////////////////////////////////////////////

type atom struct {
	kind  atomType
	value uint32
}

////////////////////////////////////////////////////////////////////////////////

// pattern is a byte pattern in the pelite syntax that cs2-dumper uses. Only
// the subset of the syntax that cs2-dumper needs is supported, so patterns
// can be copied from upstream unchanged:
//
//	48 8b      match a byte
//	?          skip one byte
//	[8]        skip eight bytes
//	'          save the cursor into the next save slot
//	u1 u2 u4   read a value into the next save slot
//	${...}     follow a rel32 operand, match inside, then continue after it
//
// Save slot 0 is always the start of the match.
type pattern struct {
	atoms  []atom
	prefix []byte
	saves  uint32
}

////////////////////////////////////////////////////////////////////////////////

func parsePattern(text string) (*pattern, error) {

	//----------------------------------------------------------------------------//

	result := &pattern{
		atoms: []atom{{atomSave, 0}},
		saves: 1,
	}

	depth := 0

	//----------------------------------------------------------------------------//

	for i := 0; i < len(text); i++ {

		switch text[i] {
		case ' ', '\t', '\r', '\n':
			continue

		case '$':
			result.atoms = append(result.atoms, atom{atomJump4, 0})

		case '{':
			last := len(result.atoms) - 1
			if result.atoms[last].kind != atomJump4 {
				return nil, errors.New(
					"pattern stack must follow a jump",
					errors.String("pattern", text),
					errors.Int("index", i),
				)
			}

			// The push records where to continue once the jump returns
			result.atoms[last] = atom{atomPush, 4}
			result.atoms = append(result.atoms, atom{atomJump4, 0})
			depth++

		case '}':
			if depth == 0 {
				return nil, errors.New(
					"pattern stack is unbalanced",
					errors.String("pattern", text),
					errors.Int("index", i),
				)
			}

			result.atoms = append(result.atoms, atom{atomPop, 0})
			depth--

		case '\'':
			result.atoms = append(result.atoms, atom{atomSave, result.saves})
			result.saves++

		case '?':
			result.atoms = append(result.atoms, atom{atomSkip, 1})

		case '[':
			end := strings.IndexByte(text[i:], ']')
			if end < 0 {
				return nil, errors.New(
					"invalid pattern skip",
					errors.String("pattern", text),
					errors.Int("index", i),
				)
			}

			// Ranges such as [2-8] are not used by cs2-dumper
			skip, err := strconv.ParseUint(text[i+1:i+end], 10, 16)
			if err != nil {
				return nil, errors.New(
					"invalid pattern skip",
					errors.String("pattern", text),
					errors.Int("index", i),
					errors.Error("error", err),
				)
			}

			if skip > 0 {
				result.atoms = append(result.atoms, atom{atomSkip, uint32(skip)})
			}

			i += end

		case 'u':
			var width byte
			if i+1 < len(text) {
				width = text[i+1]
			}

			var kind atomType
			switch width {
			case '1':
				kind = atomReadU8

			case '2':
				kind = atomReadU16

			case '4':
				kind = atomReadU32

			default:
				return nil, errors.New(
					"invalid pattern read",
					errors.String("pattern", text),
					errors.Int("index", i),
				)
			}

			result.atoms = append(result.atoms, atom{kind, result.saves})
			result.saves++
			i++

		default:
			if i+1 >= len(text) {
				return nil, errors.New(
					"invalid pattern byte",
					errors.String("pattern", text),
					errors.Int("index", i),
				)
			}

			value, err := strconv.ParseUint(text[i:i+2], 16, 8)
			if err != nil {
				return nil, errors.New(
					"invalid pattern byte",
					errors.String("pattern", text),
					errors.Int("index", i),
					errors.Error("error", err),
				)
			}

			result.atoms = append(result.atoms, atom{atomByte, uint32(value)})
			i++
		}
	}

	if depth != 0 {
		return nil, errors.New(
			"pattern stack is unbalanced",
			errors.String("pattern", text),
		)
	}

	//----------------------------------------------------------------------------//

	// The leading bytes are searched for directly. Every pattern in
	// cs2-dumper starts with at least one.
	for _, a := range result.atoms[1:] {
		if a.kind != atomByte {
			break
		}

		result.prefix = append(result.prefix, byte(a.value))
	}

	if len(result.prefix) == 0 {
		return nil, errors.New(
			"pattern does not start with a byte",
			errors.String("pattern", text),
		)
	}

	//----------------------------------------------------------------------------//

	return result, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// matcher runs the atoms of a pattern at one cursor. It follows pelite's
// interpreter, including a push sharing `pc` with the nested run so that
// matching resumes after the matching pop.
type matcher struct {
	data   []byte
	atoms  []atom
	cursor uint32
	pc     int
}

////////////////////////////////////////////////////////////////////////////////

func (m *matcher) exec(save []uint32) bool {

	for m.pc < len(m.atoms) {

		a := m.atoms[m.pc]
		m.pc++

		switch a.kind {
		case atomByte:
			value, ok := m.read(1)
			if !ok || value != a.value {
				return false
			}

			m.cursor++

		case atomSave:
			save[a.value] = m.cursor

		case atomPush:
			next := m.cursor + a.value
			if !m.exec(save) {
				return false
			}

			m.cursor = next

		case atomPop:
			return true

		case atomSkip:
			m.cursor += a.value

		case atomJump4:
			value, ok := m.read(4)
			if !ok {
				return false
			}

			// The operand is a signed rel32 from the end of itself,
			// uint32 wrapping handles negative values
			m.cursor += value + 4

		case atomReadU8:
			value, ok := m.read(1)
			if !ok {
				return false
			}

			save[a.value] = value
			m.cursor++

		case atomReadU16:
			value, ok := m.read(2)
			if !ok {
				return false
			}

			save[a.value] = value
			m.cursor += 2

		case atomReadU32:
			value, ok := m.read(4)
			if !ok {
				return false
			}

			save[a.value] = value
			m.cursor += 4
		}
	}

	return true
}

////////////////////////////////////////////////////////////////////////////////

func (m *matcher) read(size uint32) (uint32, bool) {

	start := uint64(m.cursor)
	if start+uint64(size) > uint64(len(m.data)) {
		return 0, false
	}

	switch size {
	case 1:
		return uint32(m.data[start]), true

	case 2:
		return uint32(binary.LittleEndian.Uint16(m.data[start:])), true
	}

	return binary.LittleEndian.Uint32(m.data[start:]), true
}

////////////////////////////////////////////////////////////////////////////////

// image is a module copied out of the game. The copy keeps the in-memory
// section layout, so offsets into it are RVAs.
type image struct {
	data      []byte
	codeStart uint32
	codeEnd   uint32
}

////////////////////////////////////////////////////////////////////////////////

func readImage(memory *leech.Memory, module *leech.Module) (*image, error) {

	// Pages that fail to read are zero filled, like memflow does
	data, err := memory.ReadData(module.GetBase(), module.GetSize())
	if err != nil {
		return nil, errors.New(
			"failed to read module",
			errors.String("name", module.GetName()),
			errors.Error("error", err),
		)
	}

	img, err := imageFromBytes(data)
	if err != nil {
		return nil, errors.New(
			"failed to parse module image",
			errors.String("name", module.GetName()),
			errors.Error("error", err),
		)
	}

	return img, nil
}

////////////////////////////////////////////////////////////////////////////////

// imageFromBytes wraps a module copy that keeps the in-memory section
// layout. The headers are laid out the same in memory as on disk, so the
// standard PE parser can read them.
func imageFromBytes(data []byte) (*image, error) {

	file, err := pe.NewFile(bytes.NewReader(data))
	if err != nil {
		return nil, errors.New(
			"failed to parse pe headers",
			errors.Error("error", err),
		)
	}

	header, ok := file.OptionalHeader.(*pe.OptionalHeader64)
	if !ok {
		return nil, errors.New("unsupported pe format")
	}

	// Same code range as pelite's `code_range`, clamped to the copy
	codeEnd := uint64(header.BaseOfCode) + uint64(header.SizeOfCode)
	codeEnd = min(codeEnd, uint64(len(data)))

	return &image{
		data:      data,
		codeStart: header.BaseOfCode,
		codeEnd:   uint32(codeEnd),
	}, nil
}

////////////////////////////////////////////////////////////////////////////////

// find runs a pattern over the code section and returns its save slots. Like
// pelite's `finds_code`, the pattern has to match exactly once, because a
// second match means the pattern is no longer specific enough. Returns nil
// when there is no match or more than one.
func (i *image) find(text string) ([]uint32, error) {

	//----------------------------------------------------------------------------//

	p, err := parsePattern(text)
	if err != nil {
		return nil, errors.New(
			"failed to parse pattern",
			errors.Error("error", err),
		)
	}

	//----------------------------------------------------------------------------//

	var result []uint32
	code := i.data[:i.codeEnd]

	for start := i.codeStart; start < i.codeEnd; {

		index := bytes.Index(code[start:], p.prefix)
		if index < 0 {
			break
		}

		cursor := start + uint32(index)
		start = cursor + 1

		save := make([]uint32, p.saves)
		m := &matcher{data: i.data, atoms: p.atoms, cursor: cursor}

		if !m.exec(save) {
			continue
		}

		if result != nil {
			return nil, nil
		}

		result = save
	}

	//----------------------------------------------------------------------------//

	return result, nil

	//----------------------------------------------------------------------------//
}
