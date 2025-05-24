package manifest

import "errors"

var (
	ErrNameUnknown     = errors.New("unknown name")
	ErrManifestUnknown = errors.New("unknown manifest")
)
