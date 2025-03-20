package manifest

import v1 "github.com/google/go-containerregistry/pkg/v1"

type Manifest struct {
	BlobDigest  v1.Hash
	ContentType string
}
