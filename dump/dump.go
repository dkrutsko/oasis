package dump

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/dkrutsko/oasis/errors"
	"github.com/dkrutsko/oasis/game"
	"github.com/dkrutsko/oasis/leech"
	"github.com/dkrutsko/oasis/logger"
)

////////////////////////////////////////////////////////////////////////////////

// Read cache for the schema walk, which makes many small reads close to each
// other. The block sizes match the action cache in `game`, the limit is
// higher because the walk touches more memory.
const (
	cacheBlockLength = 16384
	cacheBlockBuffer = 4096
	cacheInitialSize = 5242880
	cacheEnlargeSize = 1048576
	cacheMaximumSize = 67108864
)

////////////////////////////////////////////////////////////////////////////////

// Module whose schemas are written to `client_dll.json`
const clientModule = "client.dll"

////////////////////////////////////////////////////////////////////////////////

// Dump reads the offsets and the `client.dll` schemas from the running game
// through `l`, and writes them to `offsets.json` and `client_dll.json` in
// `directory`, replacing existing files. The files have the same format as
// the ones from cs2-dumper. Nothing is written when either dump fails.
func Dump(ctx context.Context, l *leech.Leech, directory string) error {

	//----------------------------------------------------------------------------//

	start := time.Now()

	processList, err := l.GetProcessList(game.GameProcessName, true)
	if err != nil {
		return errors.New(
			"failed to get process list",
			errors.Error("error", err),
		)
	}

	if len(processList) == 0 {
		return errors.New("game is not running")
	}
	if len(processList) > 1 {
		return errors.New("multiple game processes are running")
	}

	process := processList[0]

	modules, err := process.GetModules(ctx, nil)
	if err != nil {
		return errors.New(
			"failed to get modules",
			errors.Error("error", err),
		)
	}

	memory := process.GetMemory()

	err = memory.CreateCache(
		cacheBlockLength,
		cacheBlockBuffer,
		cacheInitialSize,
		cacheEnlargeSize,
		cacheMaximumSize,
	)
	if err != nil {
		return errors.New(
			"failed to create cache",
			errors.Error("error", err),
		)
	}

	//----------------------------------------------------------------------------//

	offsets, err := dumpOffsets(memory, modules)
	if err != nil {
		return errors.New(
			"failed to dump offsets",
			errors.Error("error", err),
		)
	}

	schemas, err := dumpSchemas(ctx, memory, modules, clientModule)
	if err != nil {
		return errors.New(
			"failed to dump schemas",
			errors.Error("error", err),
		)
	}

	//----------------------------------------------------------------------------//

	offsetsJson, err := encodeJson(offsets)
	if err != nil {
		return errors.New(
			"failed to encode offsets",
			errors.Error("error", err),
		)
	}

	schemasJson, err := encodeJson(map[string]*schemaModule{clientModule: schemas})
	if err != nil {
		return errors.New(
			"failed to encode schemas",
			errors.Error("error", err),
		)
	}

	//----------------------------------------------------------------------------//

	err = os.MkdirAll(directory, 0o755)
	if err != nil {
		return errors.New(
			"failed to create directory",
			errors.String("directory", directory),
			errors.Error("error", err),
		)
	}

	files := []struct {
		name string
		data []byte
	}{
		{"offsets.json", offsetsJson},
		{"client_dll.json", schemasJson},
	}

	for _, file := range files {

		path := filepath.Join(directory, file.name)

		err = os.WriteFile(path, file.data, 0o644)
		if err != nil {
			return errors.New(
				"failed to write file",
				errors.String("path", path),
				errors.Error("error", err),
			)
		}
	}

	//----------------------------------------------------------------------------//

	logger.Info(
		"dumped offsets",
		logger.String("directory", directory),
		logger.Int("classes", len(schemas.Classes)),
		logger.Int("enums", len(schemas.Enums)),
		logger.Elapsed("elapsed", start),
	)

	return nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// encodeJson matches the pretty printer of serde_json, which indents with two
// spaces, leaves HTML characters such as < unescaped and adds no trailing
// newline.
func encodeJson(value any) ([]byte, error) {

	var buf bytes.Buffer

	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")

	err := encoder.Encode(value)
	if err != nil {
		return nil, errors.New(
			"failed to encode json",
			errors.Error("error", err),
		)
	}

	return bytes.TrimSuffix(buf.Bytes(), []byte("\n")), nil
}
