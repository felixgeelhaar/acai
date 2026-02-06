package granola

import "errors"

// Infrastructure-level errors for the Granola API.
// These are mapped to domain errors in the repository adapter.
var (
	ErrNotFound     = errors.New("granola: resource not found")
	ErrRateLimited  = errors.New("granola: rate limited")
	ErrUnauthorized = errors.New("granola: unauthorized")
)
