package policy

import "errors"

var (
	ErrAccessDenied  = errors.New("access denied by policy")
	ErrInvalidPolicy = errors.New("invalid policy configuration")
)
