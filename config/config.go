package config

import (
	"flag"
	"sync"
)

////////////////////////////////////////////////////////////////////////////////

type Config struct {
	Debug   bool
	Json    bool
	Version bool
	Addr    string
	Port    uint
}

////////////////////////////////////////////////////////////////////////////////

var (
	instance     *Config
	instanceLock sync.Mutex
)

////////////////////////////////////////////////////////////////////////////////

func LoadConfig() error {

	//----------------------------------------------------------------------------//

	// If already loaded
	// Without using lock
	if instance != nil {
		return nil
	}

	// Lock when loading
	instanceLock.Lock()
	defer instanceLock.Unlock()

	// If already loaded
	if instance != nil {
		return nil
	}

	result := &Config{}

	//----------------------------------------------------------------------------//

	// Define command-line arguments
	flag.BoolVar(&result.Debug, "debug", false, "")
	flag.BoolVar(&result.Json, "json", false, "")
	flag.BoolVar(&result.Version, "version", false, "")
	flag.StringVar(&result.Addr, "addr", "localhost", "")
	flag.UintVar(&result.Port, "port", 8080, "")

	// Use custom output for usage
	flag.Usage = Usage

	// Parse command-line arguments
	flag.Parse()

	//----------------------------------------------------------------------------//

	instance = result
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func GetConfig() *Config {
	return instance
}
