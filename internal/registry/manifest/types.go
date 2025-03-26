package manifest

import v1 "github.com/google/go-containerregistry/pkg/v1"

type Manifest struct {
	Blob        Blob
	ContentType string
}

type Blob struct {
	Digest v1.Hash
	Size   int64
}
