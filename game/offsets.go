package game

import (
	"path/filepath"

	"github.com/buger/jsonparser"
)

////////////////////////////////////////////////////////////////////////////////

func (g *Game) GetOffsetsStr(keys ...string) (string, bool) {

	// If the keys exist
	if len(keys) < 2 {
		return "", false
	}

	// Construct key for caching
	key := filepath.Join(keys...)

	// Lock the cache
	g.cacheLock.Lock()
	defer g.cacheLock.Unlock()

	// Check whether the data is cached
	if value, ok := g.strCache[key]; ok {
		return value, true
	}

	// Grab data relating to file
	data, ok := g.offsets[keys[0]]
	if !ok {
		return "", false
	}

	// Try to get the value
	value, err := jsonparser.GetString(data, keys[1:]...)
	if err != nil {
		return "", false
	}

	// Store data in cache
	g.strCache[key] = value
	return value, true
}

////////////////////////////////////////////////////////////////////////////////

func (g *Game) GetOffsetsInt(keys ...string) (uintptr, bool) {

	// If the keys exist
	if len(keys) < 2 {
		return 0, false
	}

	// Construct key for caching
	key := filepath.Join(keys...)

	// Lock the cache
	g.cacheLock.Lock()
	defer g.cacheLock.Unlock()

	// Check whether the data is cached
	if value, ok := g.intCache[key]; ok {
		return value, true
	}

	// Grab data relating to file
	data, ok := g.offsets[keys[0]]
	if !ok {
		return 0, false
	}

	// Try to get the value
	value, err := jsonparser.GetInt(data, keys[1:]...)
	if err != nil {
		return 0, false
	}

	// Store data in cache
	g.intCache[key] = uintptr(value)
	return uintptr(value), true
}
