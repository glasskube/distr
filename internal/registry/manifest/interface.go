package manifest

import (
	"context"
	"io"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

type ManifestHandler interface {
	List(ctx context.Context, name string) ([]v1.Manifest, error)
	Get(ctx context.Context, name string, reference string) (v1.Manifest, error)
	GetReader(ctx context.Context, name string, reference string) (io.Reader, error)
	Put(ctx context.Context, name string, reference string, contentType string, hash v1.Hash) error
	Delete(ctx context.Context, name string, reference string) error
}
