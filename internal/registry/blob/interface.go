package blob

import (
	"context"
	"io"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

// BlobHandler represents a minimal blob storage backend, capable of serving
// blob contents.
type BlobHandler interface {
	// Get gets the blob contents, or errNotFound if the blob wasn't found.
	Get(ctx context.Context, repo string, h v1.Hash, allowRedirect bool) (io.ReadCloser, error)
}

// BlobStatHandler is an extension interface representing a blob storage
// backend that can serve metadata about blobs.
type BlobStatHandler interface {
	// Stat returns the size of the blob, or errNotFound if the blob wasn't
	// found, or redirectError if the blob can be found elsewhere.
	Stat(ctx context.Context, repo string, h v1.Hash) (int64, error)
}

// BlobPutHandler is an extension interface representing a blob storage backend
// that can write blob contents.
type BlobPutHandler interface {
	// Put puts the blob contents.
	//
	// The contents will be verified against the expected size and digest
	// as the contents are read, and an error will be returned if these
	// don't match. Implementations should return that error, or a wrapper
	// around that error, to return the correct error when these don't match.
	Put(ctx context.Context, repo string, h v1.Hash, contentType string, r io.Reader) error
}

// BlobDeleteHandler is an extension interface representing a blob storage
// backend that can delete blob contents.
type BlobDeleteHandler interface {
	// Delete the blob contents.
	Delete(ctx context.Context, repo string, h v1.Hash) error
}
