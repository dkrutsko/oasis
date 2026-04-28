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
