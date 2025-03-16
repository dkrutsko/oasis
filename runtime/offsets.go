package runtime

import (
	"path/filepath"

	"github.com/buger/jsonparser"
)

////////////////////////////////////////////////////////////////////////////////

func (r *Runtime) GetOffsetsStr(keys ...string) (string, bool) {

	// Check for keys and if offsets are loaded
	if len(keys) == 0 || r.offsetsData == nil {
		return "", false
	}

	// Construct key for caching
	key := filepath.Join(keys...)
	key = filepath.Join("offsets", key)

	r.lock.RLock()
	// Check whether the data is cached
	if value, ok := r.strCache[key]; ok {
		r.lock.RUnlock()
		return value, true
	}
	r.lock.RUnlock()

	// Try to get the value
	value, err := jsonparser.GetString(r.offsetsData, keys...)
	if err != nil {
		return "", false
	}

	r.lock.Lock()
	// Store data in cache
	r.strCache[key] = value
	r.lock.Unlock()

	return value, true
}

////////////////////////////////////////////////////////////////////////////////

func (r *Runtime) GetOffsetsInt(keys ...string) (int64, bool) {

	// Check for keys and if offsets are loaded
	if len(keys) == 0 || r.offsetsData == nil {
		return 0, false
	}

	// Construct key for caching
	key := filepath.Join(keys...)
	key = filepath.Join("offsets", key)

	r.lock.RLock()
	// Check whether the data is cached
	if value, ok := r.intCache[key]; ok {
		r.lock.RUnlock()
		return value, true
	}
	r.lock.RUnlock()

	// Try to get the value
	value, err := jsonparser.GetInt(r.offsetsData, keys...)
	if err != nil {
		return 0, false
	}

	r.lock.Lock()
	// Store data in cache
	r.intCache[key] = value
	r.lock.Unlock()

	return value, true
}
