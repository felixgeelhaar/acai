package localcache

import "errors"

// Infrastructure-level errors for the Granola local cache.
// These are mapped to domain errors in the repository adapter.
var (
	ErrCacheFileNotFound = errors.New("localcache: cache file not found")
	ErrCacheFileCorrupt  = errors.New("localcache: cache file is corrupt or unreadable")
)
