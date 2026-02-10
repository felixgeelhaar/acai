package localcache

import (
	"encoding/json"
	"fmt"
	"os"
)

// Reader handles file I/O and double-JSON decoding for the Granola cache file.
type Reader struct {
	path string
}

// NewReader creates a Reader for the given cache file path.
func NewReader(path string) *Reader {
	return &Reader{path: path}
}

// Path returns the file path this reader is configured for.
func (r *Reader) Path() string {
	return r.path
}

// Read loads and decodes the cache file.
// The file uses double-JSON encoding: the outer JSON has a "cache" field
// containing a JSON-encoded string that must be decoded a second time.
func (r *Reader) Read() (*CacheState, error) {
	data, err := os.ReadFile(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrCacheFileNotFound, r.path)
		}
		return nil, fmt.Errorf("%w: %v", ErrCacheFileCorrupt, err)
	}

	// First decode: extract the "cache" JSON string
	var envelope CacheFileEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("%w: outer JSON: %v", ErrCacheFileCorrupt, err)
	}
	if envelope.Cache == "" {
		return nil, fmt.Errorf("%w: empty cache field", ErrCacheFileCorrupt)
	}

	// Second decode: parse the inner JSON string
	var state CacheState
	if err := json.Unmarshal([]byte(envelope.Cache), &state); err != nil {
		return nil, fmt.Errorf("%w: inner JSON: %v", ErrCacheFileCorrupt, err)
	}

	return &state, nil
}
