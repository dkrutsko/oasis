// Command extractor converts the collision geometry of a CS2 map VPK into a
// `.tri` file for Oasis. The hull and mesh conversion is based on
// CS2-Phys-Extractor, and spheres and capsules are added as triangles.
//
// Usage: extractor [-debug] [-json] <input.vpk> <output-dir>
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/klauspost/compress/zstd"

	"github.com/dkrutsko/oasis/errors"
	"github.com/dkrutsko/oasis/logger"
)

////////////////////////////////////////////////////////////////////////////////

type exitCodeType int

const (
	exitCodeSuccess exitCodeType = iota
	exitCodeParseArgs
	exitCodeReadVpk
	exitCodeNoWorldPhysics
	exitCodeCreateOutput
	exitCodeCreateDecoder
	exitCodeConvertMap
)

////////////////////////////////////////////////////////////////////////////////

func usage() {

	fmt.Print(
		`
USAGE
-----
	extractor [-debug] [-json] <input.vpk> <output-dir>

ARGUMENTS
---------
	-json (false) - boolean
	 Format log output as JSON lines.

	-debug (false) - boolean
	 Whether to output extended logging information for debugging.
`,
	)
}

////////////////////////////////////////////////////////////////////////////////

func main() {

	//----------------------------------------------------------------------------//

	flagSet := flag.NewFlagSet("", flag.ContinueOnError)

	// Define command-line arguments
	jsonOutput := flagSet.Bool("json", false, "")
	debug := flagSet.Bool("debug", false, "")

	// Use custom output for usage
	flagSet.Usage = usage

	// Parse command-line arguments
	parseErr := flagSet.Parse(os.Args[1:])

	{
		opts := logger.NewOptions()

		if *debug {
			opts.Level = slog.LevelDebug
		} else {
			opts.Level = slog.LevelInfo
		}

		opts.Json = *jsonOutput
		logger.SetLogger("", logger.New(opts))
	}

	// Delay handling the parse error
	// until the logger setup is complete
	if parseErr != nil {
		logger.Err(
			"failed to parse arguments",
			logger.Error("error", parseErr),
		)
		os.Exit(int(exitCodeParseArgs))
	}

	if flagSet.NArg() != 2 {
		usage()
		os.Exit(int(exitCodeParseArgs))
	}

	vpkPath := flagSet.Arg(0)
	outDir := flagSet.Arg(1)

	//----------------------------------------------------------------------------//

	logger.Info(
		"reading vpk",
		logger.String("path", vpkPath),
	)

	file, err := os.Open(vpkPath)
	if err != nil {
		logger.Err(
			"failed to open vpk",
			logger.String("path", vpkPath),
			logger.Error("error", err),
		)
		os.Exit(int(exitCodeReadVpk))
	}

	info, err := file.Stat()
	if err != nil {
		logger.Err(
			"failed to get vpk size",
			logger.String("path", vpkPath),
			logger.Error("error", err),
		)
		os.Exit(int(exitCodeReadVpk))
	}

	vpk, err := ReadVpk(file, info.Size())
	if err != nil {
		logger.Err(
			"failed to read vpk",
			logger.String("path", vpkPath),
			logger.Error("error", err),
		)
		os.Exit(int(exitCodeReadVpk))
	}

	// Vanity and settings VPKs in the maps folder have no world
	// physics, so they get an exit code of their own
	entries := findWorldPhysics(vpk)
	if len(entries) == 0 {
		logger.Err(
			"vpk has no world physics",
			logger.String("path", vpkPath),
		)
		os.Exit(int(exitCodeNoWorldPhysics))
	}

	//----------------------------------------------------------------------------//

	err = os.MkdirAll(outDir, 0o755)
	if err != nil {
		logger.Err(
			"failed to create output directory",
			logger.String("path", outDir),
			logger.Error("error", err),
		)
		os.Exit(int(exitCodeCreateOutput))
	}

	// Version 5 KV3 blocks hold three zstd frames that are
	// decoded at the same time
	zd, err := zstd.NewReader(nil, zstd.WithDecoderConcurrency(3))
	if err != nil {
		logger.Err(
			"failed to create zstd decoder",
			logger.Error("error", err),
		)
		os.Exit(int(exitCodeCreateDecoder))
	}

	//----------------------------------------------------------------------------//

	// Every map is converted even if an earlier one fails, so a
	// single run reports all failures
	failed := false

	for _, entry := range entries {
		mapName := getMapName(entry)

		err := convertWorldPhysics(vpk, entry, mapName, outDir, zd)
		if err != nil {
			logger.Err(
				"failed to convert map",
				logger.String("map", mapName),
				logger.Error("error", err),
			)
			failed = true
		}
	}

	zd.Close()
	file.Close()

	if failed {
		os.Exit(int(exitCodeConvertMap))
	}

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// findWorldPhysics returns the world physics entries of the maps
// in a VPK. Official map VPKs hold exactly one, stored as
// `maps/<name>/world_physics.vmdl_c`.
func findWorldPhysics(vpk *Vpk) []*VpkEntry {

	var result []*VpkEntry

	for i := range vpk.Entries {
		entry := &vpk.Entries[i]

		if entry.Extension != "vmdl_c" || entry.Name != "world_physics" {
			continue
		}

		if getMapName(entry) == "" {
			logger.Warn(
				"skipping world physics outside a map directory",
				logger.String("directory", entry.Directory),
			)
			continue
		}

		result = append(result, entry)
	}

	return result
}

////////////////////////////////////////////////////////////////////////////////

// getMapName returns the name of the map a world physics entry
// belongs to. It returns an empty string when the entry is not
// in `maps/<name>` or the name is not safe to use as a file name.
func getMapName(entry *VpkEntry) string {

	name, found := strings.CutPrefix(entry.Directory, "maps/")
	if !found || name == "" {
		return ""
	}

	// The name becomes the output file name, so only characters
	// that cannot form a path are allowed
	for _, c := range name {
		isLetter := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
		isDigit := c >= '0' && c <= '9'

		if !isLetter && !isDigit && c != '_' && c != '-' {
			return ""
		}
	}

	return name
}

////////////////////////////////////////////////////////////////////////////////

// convertWorldPhysics decodes the physics of one map and writes
// its collision triangles to `<outDir>/<mapName>.tri`.
func convertWorldPhysics(
	vpk *Vpk,
	entry *VpkEntry,
	mapName string,
	outDir string,
	zd *zstd.Decoder,
) error {

	//----------------------------------------------------------------------------//

	logger.Info(
		"converting map",
		logger.String("map", mapName),
	)

	// The physics are the PHYS block of the compiled model
	start := time.Now()

	data, err := vpk.ReadEntry(entry)
	if err != nil {
		return errors.New(
			"failed to read world physics",
			errors.Error("error", err),
		)
	}

	block, err := ReadResourceBlock(data, "PHYS")
	if err != nil {
		return errors.New(
			"failed to read phys block",
			errors.Error("error", err),
		)
	}

	phys, err := DecodeKv3(block, PhysicsKeys, zd)
	if err != nil {
		return errors.New(
			"failed to decode phys block",
			errors.Error("error", err),
		)
	}

	decodeTime := time.Since(start)

	//----------------------------------------------------------------------------//

	// Triangles stream to the file instead of being collected
	start = time.Now()
	path := filepath.Join(outDir, mapName+".tri")

	writer, err := createTriWriter(path)
	if err != nil {
		return err
	}

	result := ConvertPhysics(phys, writer.Write)

	if writer.GetCount() == 0 {
		writer.Abort()
		return errors.New("no collision triangles found")
	}

	err = writer.Commit()
	if err != nil {
		return err
	}

	convertTime := time.Since(start)

	//----------------------------------------------------------------------------//

	logger.Info(
		"wrote tri file",
		logger.String("path", path),
		logger.Int("triangles", writer.GetCount()),
		logger.Int("hulls", result.Hulls),
		logger.Int("meshes", result.Meshes),
		logger.Int("spheres", result.Spheres),
		logger.Int("capsules", result.Capsules),
		logger.Duration("decode", decodeTime),
		logger.Duration("convert", convertTime),
	)

	return nil

	//----------------------------------------------------------------------------//
}
