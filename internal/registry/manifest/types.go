package manifest

import (
	"github.com/opencontainers/go-digest"
)

type Manifest struct {
	Blob        Blob
	ContentType string
}

type Blob struct {
	Digest digest.Digest
	Size   int64
}
