package inmemory

import (
	"context"
	"maps"
	"slices"
	"sort"

	"github.com/glasskube/distr/internal/registry/manifest"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

type handler struct {
	manifests map[string]map[string]manifest.Manifest
}

func NewManifestHandler() manifest.ManifestHandler {
	return &handler{
		manifests: make(map[string]map[string]manifest.Manifest),
	}
}

// Delete implements manifest.ManifestHandler.
func (h *handler) Delete(ctx context.Context, name string, reference string) error {
	if _, err := h.Get(ctx, name, reference); err != nil {
		return err
	}
	delete(h.manifests[name], reference)
	return nil
}

// Get implements manifest.ManifestHandler.
func (h *handler) Get(ctx context.Context, name string, reference string) (*manifest.Manifest, error) {
	if references, ok := h.manifests[name]; !ok {
		return nil, manifest.ErrNameUnknown
	} else if m, ok := references[reference]; !ok {
		return nil, manifest.ErrManifestUnknown
	} else {
		return &m, nil
	}
}

// List implements manifest.ManifestHandler.
func (h *handler) List(ctx context.Context, n int) ([]string, error) {
	names := slices.Collect(maps.Keys(h.manifests))
	if 0 < n && n < len(names) {
		names = names[:n]
	}
	return names, nil
}

// ListTags implements manifest.ManifestHandler.
func (h *handler) ListTags(ctx context.Context, name string, n int, last string) ([]string, error) {
	referencesMap, ok := h.manifests[name]
	if !ok {
		return nil, manifest.ErrNameUnknown
	}
	var references []string
	for reference := range referencesMap {
		if _, err := v1.NewHash(reference); err == nil {
			continue
		} else {
			references = append(references, reference)
		}
	}

	sort.Strings(references)

	if last != "" {
		for i, reference := range references {
			if reference > last {
				references = references[i:]
				break
			}
		}
	}

	if 0 < n && n < len(references) {
		references = references[:n]
	}

	return references, nil
}

// ListDigests implements manifest.ManifestHandler.
func (h *handler) ListDigests(ctx context.Context, name string) ([]v1.Hash, error) {
	references, ok := h.manifests[name]
	if !ok {
		return nil, manifest.ErrNameUnknown
	}
	var digests []v1.Hash
	for reference := range references {
		if h, err := v1.NewHash(reference); err != nil {
			continue
		} else {
			digests = append(digests, h)
		}
	}
	return digests, nil
}

// Put implements manifest.ManifestHandler.
func (h *handler) Put(ctx context.Context, name string, reference string, m manifest.Manifest, blobs []v1.Hash) error {
	if _, ok := h.manifests[name]; !ok {
		h.manifests[name] = make(map[string]manifest.Manifest)
	}
	h.manifests[name][reference] = m
	return nil
}
