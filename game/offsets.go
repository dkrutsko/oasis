package game

import (
	"path/filepath"

	"github.com/buger/jsonparser"
)

////////////////////////////////////////////////////////////////////////////////

func (g *Game) GetOffsetsStr(keys ...string) (string, bool) {

	// If keys exist and offsets are loaded
	if len(keys) == 0 || g.offsets == nil {
		return "", false
	}

	// Construct key for caching
	key := filepath.Join(keys...)
	key = filepath.Join("offsets", key)

	g.cacheLock.RLock()
	// Check whether the data is cached
	if value, ok := g.strCache[key]; ok {
		g.cacheLock.RUnlock()
		return value, true
	}
	g.cacheLock.RUnlock()

	// Try to get the value
	value, err := jsonparser.GetString(g.offsets, keys...)
	if err != nil {
		return "", false
	}

	g.cacheLock.Lock()
	// Store data in cache
	g.strCache[key] = value
	g.cacheLock.Unlock()

	return value, true
}

////////////////////////////////////////////////////////////////////////////////

func (g *Game) GetOffsetsInt(keys ...string) (int64, bool) {

	// If keys exist and offsets are loaded
	if len(keys) == 0 || g.offsets == nil {
		return 0, false
	}

	// Construct key for caching
	key := filepath.Join(keys...)
	key = filepath.Join("offsets", key)

	g.cacheLock.RLock()
	// Check whether the data is cached
	if value, ok := g.intCache[key]; ok {
		g.cacheLock.RUnlock()
		return value, true
	}
	g.cacheLock.RUnlock()

	// Try to get the value
	value, err := jsonparser.GetInt(g.offsets, keys...)
	if err != nil {
		return 0, false
	}

	g.cacheLock.Lock()
	// Store data in cache
	g.intCache[key] = value
	g.cacheLock.Unlock()

	return value, true
}
