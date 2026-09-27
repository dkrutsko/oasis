// Command extractor converts the collision geometry of a CS2 map VPK into a
// `.tri` file for Oasis. It is a Go reimplementation of CS2-Phys-Extractor
// and produces byte for byte identical output.
//
// Usage: extractor [-debug] [-json] <input.vpk> <output-dir>
package main

import (
	"bufio"
	"bytes"
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
	exitCodeCreateOutput
	exitCodeCreateDecoder
	exitCodeExtract
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

	err := os.MkdirAll(outDir, 0o755)
	if err != nil {
		logger.Err(
			"failed to create output directory",
			logger.String("path", outDir),
			logger.Error("error", err),
		)
		os.Exit(int(exitCodeCreateOutput))
	}

	// Version 5 KV3 blocks hold three zstd frames that are decoded at once
	zd, err := zstd.NewReader(nil, zstd.WithDecoderConcurrency(3))
	if err != nil {
		logger.Err(
			"failed to create zstd decoder",
			logger.Error("error", err),
		)
		os.Exit(int(exitCodeCreateDecoder))
	}

	//----------------------------------------------------------------------------//

	err = processVpkFile(vpkPath, outDir, zd)
	zd.Close()

	if err != nil {
		logger.Err(
			"failed to extract vpk",
			logger.String("path", vpkPath),
			logger.Error("error", err),
		)
		os.Exit(int(exitCodeExtract))
	}

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

func processVpkFile(vpkPath, outDir string, zd *zstd.Decoder) error {

	logger.Info(
		"processing vpk",
		logger.String("path", vpkPath),
	)

	f, err := os.Open(vpkPath)
	if err != nil {
		return errors.New(
			"failed to open vpk",
			errors.Error("error", err),
		)
	}
	defer f.Close()

	pkg, err := ReadVpk(f)
	if err != nil {
		return errors.New(
			"failed to read vpk",
			errors.Error("error", err),
		)
	}

	// Workshop maps nest the map VPK inside another VPK
	if nested := pkg.GetExtension("vpk"); nested != nil {
		var maps []*VpkEntry
		for _, entry := range nested.Entries {
			if strings.EqualFold(entry.Directory, "maps") {
				maps = append(maps, entry)
			}
		}

		if len(maps) > 0 {
			return processNestedVpk(pkg, pickMainMapVpk(maps), outDir, zd)
		}
	}

	return processMapVpk(pkg, outDir, zd)
}

////////////////////////////////////////////////////////////////////////////////

// pickMainMapVpk skips skybox VPKs and takes the shortest name, keeping the
// first on ties, as the original does.
func pickMainMapVpk(maps []*VpkEntry) *VpkEntry {

	var best *VpkEntry

	for _, entry := range maps {
		name := strings.ToLower(entry.Name)
		if strings.Contains(name, "_3dsky") || strings.Contains(name, "_skybox") || strings.Contains(name, "_sky") {
			continue
		}
		if best == nil || len(entry.Name) < len(best.Name) {
			best = entry
		}
	}

	if best == nil {
		best = maps[0]
	}
	return best
}

////////////////////////////////////////////////////////////////////////////////

func processNestedVpk(parent *Vpk, entry *VpkEntry, outDir string, zd *zstd.Decoder) error {

	data, err := parent.ReadEntry(entry)
	if err != nil {
		return errors.New(
			"failed to extract nested vpk",
			errors.String("name", entry.Name),
			errors.Error("error", err),
		)
	}

	nested, err := ReadVpk(bytes.NewReader(data))
	if err != nil {
		return errors.New(
			"failed to read nested vpk",
			errors.String("name", entry.Name),
			errors.Error("error", err),
		)
	}

	return processMapVpk(nested, outDir, zd)
}

////////////////////////////////////////////////////////////////////////////////

func processMapVpk(pkg *Vpk, outDir string, zd *zstd.Decoder) error {

	//----------------------------------------------------------------------------//

	var groups []*VpkExtension
	for i := range pkg.Extensions {
		if strings.EqualFold(pkg.Extensions[i].Name, "vmdl_c") {
			groups = append(groups, &pkg.Extensions[i])
		}
	}

	//----------------------------------------------------------------------------//

	found := false
	var errs []error

	for _, group := range groups {
		for _, entry := range group.Entries {
			if !strings.Contains(strings.ToLower(entry.Name), "world_physics") {
				continue
			}

			mapName := extractMapName(entry)
			if mapName == "" {
				continue
			}

			found = true

			err := processWorldPhysics(pkg, entry, mapName, outDir, zd)
			if err != nil {
				err = errors.New(
					"failed to convert map",
					errors.String("map", mapName),
					errors.Error("error", err),
				)
				errs = append(errs, err)
			}
		}
	}

	//----------------------------------------------------------------------------//

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	if !found {
		return errors.New("vpk has no world physics")
	}

	return nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// extractMapName takes the directory after "maps" in the entry's path, for
// example "maps/de_dust2" gives "de_dust2".
func extractMapName(entry *VpkEntry) string {

	if entry.Directory != "" {
		parts := strings.Split(strings.ReplaceAll(entry.Directory, `\`, "/"), "/")
		for i, part := range parts {
			if strings.EqualFold(part, "maps") && i+1 < len(parts) {
				return parts[i+1]
			}
		}
		return parts[len(parts)-1]
	}

	name := strings.ReplaceAll(entry.Name, "world_physics", "")
	name = strings.ReplaceAll(name, ".vmdl_c", "")
	return strings.Trim(name, "_.")
}

////////////////////////////////////////////////////////////////////////////////

func processWorldPhysics(pkg *Vpk, entry *VpkEntry, mapName, outDir string, zd *zstd.Decoder) error {

	//----------------------------------------------------------------------------//

	logger.Info(
		"converting map",
		logger.String("map", mapName),
	)

	start := time.Now()

	data, err := pkg.ReadEntry(entry)
	if err != nil {
		return errors.New(
			"failed to read world physics",
			errors.Error("error", err),
		)
	}
	if len(data) == 0 {
		return errors.New("world physics file is empty")
	}

	readTime := time.Since(start)

	//----------------------------------------------------------------------------//

	start = time.Now()

	blocks, err := ReadResourceBlocks(data)
	if err != nil {
		return errors.New(
			"failed to read resource blocks",
			errors.Error("error", err),
		)
	}

	var physBlock *ResourceBlock
	for i := range blocks {
		if blocks[i].Type == "PHYS" {
			physBlock = &blocks[i]
			break
		}
	}
	if physBlock == nil {
		return errors.New("world physics has no phys block")
	}

	phys, err := DecodeKv3(physBlock.Data, PhysicsKeys, zd)
	if err != nil {
		return errors.New(
			"failed to decode phys block",
			errors.Error("error", err),
		)
	}

	decodeTime := time.Since(start)

	//----------------------------------------------------------------------------//

	// Triangles stream straight to the file instead of being collected
	start = time.Now()

	path := filepath.Join(outDir, mapName+".tri")

	out := &triFile{path: path}
	result := ConvertPhysics(phys, out.write)

	err = out.close()
	if err != nil {
		return errors.New(
			"failed to write tri file",
			errors.String("path", path),
			errors.Error("error", err),
		)
	}

	convertTime := time.Since(start)

	if result.Triangles == 0 {
		return errors.New("no collision triangles found")
	}

	//----------------------------------------------------------------------------//

	logger.Info(
		"wrote tri file",
		logger.String("path", path),
		logger.Int("triangles", result.Triangles),
		logger.Int("hulls", result.Hulls),
		logger.Int("meshes", result.Meshes),
		logger.Duration("read", readTime),
		logger.Duration("decode", decodeTime),
		logger.Duration("convert", convertTime),
	)

	return nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// triFile writes triangles to a `.tri` file through a buffer. The file is
// created on the first triangle, so a map without collision triangles
// leaves no file behind, as in the original.
type triFile struct {
	path string
	file *os.File
	buf  *bufio.Writer
	err  error
}

func (t *triFile) write(v1, v2, v3 []byte) {

	if t.err != nil {
		return
	}

	if t.file == nil {
		if t.file, t.err = os.Create(t.path); t.err != nil {
			return
		}
		t.buf = bufio.NewWriterSize(t.file, 1<<20)
	}

	// `bufio.Writer` keeps the first write error, which `close` reports
	t.buf.Write(v1)
	t.buf.Write(v2)
	t.buf.Write(v3)
}

func (t *triFile) close() error {

	if t.file == nil {
		return t.err
	}

	if t.err == nil {
		t.err = t.buf.Flush()
	}
	if err := t.file.Close(); t.err == nil {
		t.err = err
	}
	return t.err
}
