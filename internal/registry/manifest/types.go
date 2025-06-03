package manifest

import (
	"github.com/opencontainers/go-digest"
)

type Blob struct {
	Digest digest.Digest
	Size   int64
}

type BlobWithData struct {
	Blob
	Data []byte
}

type Manifest struct {
	BlobWithData
	ContentType string
}
