package inmemory

import (
	"bytes"
	"context"
	"io"
	"sync"

	"github.com/distr-sh/distr/internal/registry/and"
	"github.com/distr-sh/distr/internal/registry/blob"
	"github.com/opencontainers/go-digest"
)

type blobHandler struct {
	m    map[string][]byte
	lock sync.Mutex
}

func NewBlobHandler() blob.BlobHandler { return &blobHandler{m: map[string][]byte{}} }

func (m *blobHandler) Stat(_ context.Context, _ string, h digest.Digest) (int64, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	b, found := m.m[h.String()]
	if !found {
		return 0, blob.ErrNotFound
	}
	return int64(len(b)), nil
}

func (m *blobHandler) Get(_ context.Context, _ string, h digest.Digest, _ bool) (io.ReadCloser, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	b, found := m.m[h.String()]
	if !found {
		return nil, blob.ErrNotFound
	}
	return &and.BytesCloser{Reader: bytes.NewReader(b)}, nil
}

func (m *blobHandler) Put(_ context.Context, _ string, h digest.Digest, _ string, r io.Reader) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if rc, ok := r.(io.ReadCloser); ok {
		defer rc.Close()
	}
	all, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	m.m[h.String()] = all
	return nil
}

func (m *blobHandler) Delete(_ context.Context, _ string, h digest.Digest) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if _, found := m.m[h.String()]; !found {
		return blob.ErrNotFound
	}

	delete(m.m, h.String())
	return nil
}
