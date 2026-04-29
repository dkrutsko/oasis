package config

import (
	"flag"
	"os"
	"sync"

	"github.com/dkrutsko/oasis/errors"
)

////////////////////////////////////////////////////////////////////////////////

// Config holds all parsed command-line options for the application.
type Config struct {
	// Prints the version and exits without further processing.
	Version bool

	// Enables JSON-formatted output for structured logging.
	Json bool

	// Enables verbose debug logging.
	Debug bool

	// Launches the debug viewer to preview overlay output.
	Viewer bool

	// Starts a pprof HTTP server on localhost:6060 for profiling.
	Pprof bool

	// FPGA device index for action/entity reads.
	Action int

	// FPGA device index for camera reads.
	Camera int

	// Target frame rate for action reads (Hz).
	RateAction int

	// Target frame rate for camera reads (Hz).
	RateCamera int

	// Directory containing .tri map collision files.
	// When empty, line-of-sight checks are disabled.
	Maps string
}

////////////////////////////////////////////////////////////////////////////////

var (
	configInstance     *Config
	configInstanceLock sync.RWMutex
)

////////////////////////////////////////////////////////////////////////////////

// LoadConfig parses command-line arguments and returns the resulting `Config`.
// It is thread-safe and caches the result on success so subsequent calls
// return the same instance. When an error occurs, a partially filled `Config`
// is returned. If `Version` is true the config is considered valid and further
// parsing is skipped.
func LoadConfig() (*Config, error) {

	//----------------------------------------------------------------------------//

	// Ensure synchronization
	configInstanceLock.Lock()
	defer configInstanceLock.Unlock()

	// Check if already loaded
	if configInstance != nil {
		return configInstance, nil
	}

	result := &Config{}

	//----------------------------------------------------------------------------//

	flagSet := flag.NewFlagSet("", flag.ContinueOnError)

	// Define command-line arguments
	flagSet.BoolVar(&result.Version, "version", false, "")
	flagSet.BoolVar(&result.Json, "json", false, "")
	flagSet.BoolVar(&result.Debug, "debug", false, "")
	flagSet.BoolVar(&result.Viewer, "viewer", false, "")
	flagSet.BoolVar(&result.Pprof, "pprof", false, "")
	flagSet.IntVar(&result.Action, "action", 0, "")
	flagSet.IntVar(&result.Camera, "camera", 0, "")
	flagSet.StringVar(&result.Maps, "maps", "", "")

	var rateBoth int
	flagSet.IntVar(&rateBoth, "rate", 0, "")
	flagSet.IntVar(&result.RateAction, "rate-action", 60, "")
	flagSet.IntVar(&result.RateCamera, "rate-camera", 90, "")

	// Use custom output for usage
	flagSet.Usage = Usage

	// Parse command-line arguments
	err := flagSet.Parse(os.Args[1:])
	if err != nil {
		return result, errors.New(
			"failed to parse arguments",
			errors.Error("error", err),
		)
	}

	// Only show version
	if result.Version {
		configInstance = result
		return result, nil
	}

	// --rate overrides both if specified
	if rateBoth > 0 {
		result.RateAction = rateBoth
		result.RateCamera = rateBoth
	}

	//----------------------------------------------------------------------------//

	configInstance = result
	return result, nil

	//----------------------------------------------------------------------------//
}

////////////////////////////////////////////////////////////////////////////////

// GetConfig returns the cached `Config`, or nil if `LoadConfig` has not
// yet completed successfully.
func GetConfig() *Config {

	// Ensure synchronization
	configInstanceLock.RLock()
	defer configInstanceLock.RUnlock()

	return configInstance
}
