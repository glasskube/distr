// Copyright 2018 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package registry

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/registry/blob"
	"github.com/glasskube/distr/internal/registry/manifest"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/jackc/pgx/v5"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

type catalog struct {
	Repos []string `json:"repositories"`
}

type listTags struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

type manifests struct {
	blobHandler     blob.BlobHandler
	manifestHandler manifest.ManifestHandler
	lock            sync.RWMutex
	log             *zap.SugaredLogger
}

func isManifest(req *http.Request) bool {
	elems := strings.Split(req.URL.Path, "/")
	elems = elems[1:]
	if len(elems) < 4 {
		return false
	}
	return elems[len(elems)-2] == "manifests"
}

func isTags(req *http.Request) bool {
	elems := strings.Split(req.URL.Path, "/")
	elems = elems[1:]
	if len(elems) < 4 {
		return false
	}
	return elems[len(elems)-2] == "tags"
}

func isCatalog(req *http.Request) bool {
	elems := strings.Split(req.URL.Path, "/")
	elems = elems[1:]
	if len(elems) < 2 {
		return false
	}

	return elems[len(elems)-1] == "_catalog"
}

// Returns whether this url should be handled by the referrers handler
func isReferrers(req *http.Request) bool {
	elems := strings.Split(req.URL.Path, "/")
	elems = elems[1:]
	if len(elems) < 4 {
		return false
	}
	return elems[len(elems)-2] == "referrers"
}

// https://github.com/opencontainers/distribution-spec/blob/master/spec.md#pulling-an-image-manifest
// https://github.com/opencontainers/distribution-spec/blob/master/spec.md#pushing-an-image
func (handler *manifests) handle(resp http.ResponseWriter, req *http.Request) *regError {
	elem := strings.Split(req.URL.Path, "/")
	elem = elem[1:]
	target := elem[len(elem)-1]
	repo := strings.Join(elem[1:len(elem)-2], "/")

	switch req.Method {
	case http.MethodGet:
		handler.lock.RLock()
		defer handler.lock.RUnlock()

		m, err := handler.manifestHandler.Get(req.Context(), repo, target)
		if errors.Is(err, manifest.ErrNameUnknown) {
			return &regError{
				Status:  http.StatusNotFound,
				Code:    "NAME_UNKNOWN",
				Message: "Unknown name",
			}
		} else if errors.Is(err, manifest.ErrManifestUnknown) {
			return &regError{
				Status:  http.StatusNotFound,
				Code:    "MANIFEST_UNKNOWN",
				Message: "Unknown manifest",
			}
		} else if err != nil {
			return &regError{
				Status:  http.StatusNotFound,
				Code:    "INTERNAL_ERROR",
				Message: err.Error(),
			}
		}

		b, err := handler.blobHandler.Get(req.Context(), repo, m.BlobDigest, true)
		if err != nil {
			var rerr blob.RedirectError
			if errors.As(err, &rerr) {
				http.Redirect(resp, req, rerr.Location, rerr.Code)
				return nil
			}
			// TODO: More nuanced
			return &regError{
				Status:  http.StatusNotFound,
				Code:    "MANIFEST_UNKNOWN",
				Message: "Unknown manifest",
			}
		}
		defer b.Close()

		buf := bytes.Buffer{}
		if _, err = io.Copy(&buf, b); err != nil {
			return &regError{
				Status:  http.StatusNotFound,
				Code:    "INTERNAL_ERROR",
				Message: err.Error(),
			}
		}

		resp.Header().Set("Docker-Content-Digest", m.BlobDigest.String())
		resp.Header().Set("Content-Type", m.ContentType)
		resp.Header().Set("Content-Length", fmt.Sprint(buf.Len()))
		resp.WriteHeader(http.StatusOK)
		io.Copy(resp, &buf)
		return nil

	case http.MethodHead:
		handler.lock.RLock()
		defer handler.lock.RUnlock()

		m, err := handler.manifestHandler.Get(req.Context(), repo, target)
		if errors.Is(err, manifest.ErrNameUnknown) {
			return &regError{
				Status:  http.StatusNotFound,
				Code:    "NAME_UNKNOWN",
				Message: "Unknown name",
			}
		} else if errors.Is(err, manifest.ErrManifestUnknown) {
			return &regError{
				Status:  http.StatusNotFound,
				Code:    "MANIFEST_UNKNOWN",
				Message: "Unknown manifest",
			}
		} else if err != nil {
			return &regError{
				Status:  http.StatusNotFound,
				Code:    "INTERNAL_ERROR",
				Message: err.Error(),
			}
		}

		bsh, ok := handler.blobHandler.(blob.BlobStatHandler)
		if !ok {
			return &regError{
				Status:  http.StatusInternalServerError,
				Code:    "INTERNAL_ERROR",
				Message: "cannot stat blob",
			}
		}

		l, err := bsh.Stat(req.Context(), repo, m.BlobDigest)
		if err != nil {
			// TODO: More nuanced
			return &regError{
				Status:  http.StatusNotFound,
				Code:    "MANIFEST_UNKNOWN",
				Message: "Unknown manifest",
			}
		}

		resp.Header().Set("Docker-Content-Digest", m.BlobDigest.String())
		resp.Header().Set("Content-Type", m.ContentType)
		resp.Header().Set("Content-Length", fmt.Sprint(l))
		resp.WriteHeader(http.StatusOK)
		return nil

	case http.MethodPut:
		buf := &bytes.Buffer{}
		io.Copy(buf, req.Body)
		manifestDigest, _, _ := v1.SHA256(bytes.NewReader(buf.Bytes()))
		mf := manifest.Manifest{
			ContentType: req.Header.Get("Content-Type"),
			BlobDigest:  manifestDigest,
		}

		var blobs []v1.Hash

		// If the manifest is a manifest list, check that the manifest
		// list's constituent manifests are already uploaded.
		// This isn't strictly required by the registry API, but some
		// registries require this.
		if types.MediaType(mf.ContentType).IsIndex() {
			if err := func() *regError {
				handler.lock.RLock()
				defer handler.lock.RUnlock()

				im, err := v1.ParseIndexManifest(bytes.NewReader(buf.Bytes()))
				if err != nil {
					return &regError{
						Status:  http.StatusBadRequest,
						Code:    "MANIFEST_INVALID",
						Message: err.Error(),
					}
				}
				for _, desc := range im.Manifests {
					if !desc.MediaType.IsDistributable() {
						continue
					}
					if desc.MediaType.IsIndex() || desc.MediaType.IsImage() {
						if _, err := handler.manifestHandler.Get(req.Context(), repo, desc.Digest.String()); err != nil {
							return &regError{
								Status:  http.StatusNotFound,
								Code:    "MANIFEST_UNKNOWN",
								Message: fmt.Sprintf("Sub-manifest %q not found", desc.Digest),
							}
						}
						blobs = append(blobs, desc.Digest)
					} else {
						// TODO: Probably want to do an existence check for blobs.
						handler.log.Warnf("TODO: Check blobs for %q", desc.Digest)
					}
				}
				return nil
			}(); err != nil {
				return err
			}
		} else if types.MediaType(mf.ContentType).IsImage() {
			if err := func() *regError {
				m, err := v1.ParseManifest(bytes.NewReader(buf.Bytes()))
				if err != nil {
					return &regError{
						Status:  http.StatusBadRequest,
						Code:    "MANIFEST_INVALID",
						Message: err.Error(),
					}
				}
				blobs = append(blobs, m.Config.Digest)
				for _, desc := range m.Layers {
					if !desc.MediaType.IsDistributable() {
						continue
					}
					if desc.MediaType.IsLayer() {
						// TODO: Maybe check if the layer was already uploaded
						blobs = append(blobs, desc.Digest)
					} else {
						handler.log.Warnf("TODO: Check blobs for %q", desc.Digest)
					}
				}
				return nil
			}(); err != nil {
				return err
			}
		}

		handler.lock.Lock()
		defer handler.lock.Unlock()

		if bph, ok := handler.blobHandler.(blob.BlobPutHandler); !ok {
			return &regError{
				Status:  http.StatusInternalServerError,
				Code:    "INTERNAL_ERROR",
				Message: "blob handler is not a BlobPutHandler",
			}
		} else {
			if err := bph.Put(req.Context(), repo, manifestDigest, mf.ContentType, buf); err != nil {
				return &regError{
					Status:  http.StatusInternalServerError,
					Code:    "INTERNAL_ERROR",
					Message: err.Error(),
				}
			}
		}

		// Allow future references by target (tag) and immutable digest.
		// See https://docs.docker.com/engine/reference/commandline/pull/#pull-an-image-by-digest-immutable-identifier.
		err := db.RunTx(req.Context(), pgx.TxOptions{}, func(ctx context.Context) error {
			return multierr.Combine(
				handler.manifestHandler.Put(req.Context(), repo, manifestDigest.String(), mf, blobs),
				handler.manifestHandler.Put(req.Context(), repo, target, mf, blobs),
			)
		})
		if err != nil {
			return &regError{
				Status:  http.StatusInternalServerError,
				Code:    "INTERNAL_ERROR",
				Message: err.Error(),
			}
		}

		resp.Header().Set("Docker-Content-Digest", manifestDigest.String())
		resp.WriteHeader(http.StatusCreated)
		return nil

	case http.MethodDelete:
		handler.lock.Lock()
		defer handler.lock.Unlock()

		if err := handler.manifestHandler.Delete(req.Context(), repo, target); errors.Is(err, manifest.ErrNameUnknown) {
			return &regError{
				Status:  http.StatusNotFound,
				Code:    "NAME_UNKNOWN",
				Message: "Unknown name",
			}
		} else if errors.Is(err, manifest.ErrManifestUnknown) {
			return &regError{
				Status:  http.StatusNotFound,
				Code:    "MANIFEST_UNKNOWN",
				Message: "Unknown manifest",
			}
		} else if err != nil {
			return &regError{
				Status:  http.StatusNotFound,
				Code:    "INTERNAL_ERROR",
				Message: err.Error(),
			}
		}

		resp.WriteHeader(http.StatusAccepted)
		return nil

	default:
		return &regError{
			Status:  http.StatusBadRequest,
			Code:    "METHOD_UNKNOWN",
			Message: "We don't understand your method + url",
		}
	}
}

func (m *manifests) handleTags(resp http.ResponseWriter, req *http.Request) *regError {
	elem := strings.Split(req.URL.Path, "/")
	elem = elem[1:]
	repo := strings.Join(elem[1:len(elem)-2], "/")

	if req.Method == http.MethodGet {
		m.lock.RLock()
		defer m.lock.RUnlock()

		last := req.URL.Query().Get("last")
		n := 10000
		if ns := req.URL.Query().Get("n"); ns != "" {
			if parsed, err := strconv.Atoi(ns); err != nil {
				return &regError{
					Status:  http.StatusBadRequest,
					Code:    "BAD_REQUEST",
					Message: fmt.Sprintf("parsing n: %v", err),
				}
			} else {
				n = parsed
			}
		}

		references, err := m.manifestHandler.ListTags(req.Context(), repo, n, last)
		if errors.Is(err, manifest.ErrNameUnknown) {
			return &regError{
				Status:  http.StatusNotFound,
				Code:    "NAME_UNKNOWN",
				Message: "Unknown name",
			}
		} else if err != nil {
			return &regError{
				Status:  http.StatusNotFound,
				Code:    "INTERNAL_ERROR",
				Message: err.Error(),
			}
		}

		tagsToList := listTags{
			Name: repo,
			Tags: references,
		}

		msg, _ := json.Marshal(tagsToList)
		resp.Header().Set("Content-Length", fmt.Sprint(len(msg)))
		resp.WriteHeader(http.StatusOK)
		io.Copy(resp, bytes.NewReader([]byte(msg)))
		return nil
	}

	return &regError{
		Status:  http.StatusBadRequest,
		Code:    "METHOD_UNKNOWN",
		Message: "We don't understand your method + url",
	}
}

func (m *manifests) handleCatalog(resp http.ResponseWriter, req *http.Request) *regError {
	query := req.URL.Query()
	nStr := query.Get("n")
	n := 10000
	if nStr != "" {
		n, _ = strconv.Atoi(nStr)
	}

	if req.Method == http.MethodGet {
		m.lock.RLock()
		defer m.lock.RUnlock()

		repos, err := m.manifestHandler.List(req.Context(), n)
		if err != nil {
			return &regError{
				Status:  http.StatusNotFound,
				Code:    "INTERNAL_ERROR",
				Message: err.Error(),
			}
		}

		repositoriesToList := catalog{Repos: repos}

		msg, _ := json.Marshal(repositoriesToList)
		resp.Header().Set("Content-Length", fmt.Sprint(len(msg)))
		resp.WriteHeader(http.StatusOK)
		io.Copy(resp, bytes.NewReader([]byte(msg)))
		return nil
	}

	return &regError{
		Status:  http.StatusBadRequest,
		Code:    "METHOD_UNKNOWN",
		Message: "We don't understand your method + url",
	}
}

// TODO: implement handling of artifactType querystring
func (m *manifests) handleReferrers(resp http.ResponseWriter, req *http.Request) *regError {
	// Ensure this is a GET request
	if req.Method != http.MethodGet {
		return &regError{
			Status:  http.StatusBadRequest,
			Code:    "METHOD_UNKNOWN",
			Message: "We don't understand your method + url",
		}
	}

	elem := strings.Split(req.URL.Path, "/")
	elem = elem[1:]
	target := elem[len(elem)-1]
	repo := strings.Join(elem[1:len(elem)-2], "/")

	// Validate that incoming target is a valid digest
	if _, err := v1.NewHash(target); err != nil {
		return &regError{
			Status:  http.StatusBadRequest,
			Code:    "UNSUPPORTED",
			Message: "Target must be a valid digest",
		}
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	digests, err := m.manifestHandler.ListDigests(req.Context(), repo)
	if errors.Is(err, manifest.ErrNameUnknown) {
		return &regError{
			Status:  http.StatusNotFound,
			Code:    "NAME_UNKNOWN",
			Message: "Unknown name",
		}
	} else if err != nil {
		return &regError{
			Status:  http.StatusNotFound,
			Code:    "INTERNAL_ERROR",
			Message: err.Error(),
		}
	}

	im := v1.IndexManifest{
		SchemaVersion: 2,
		MediaType:     types.OCIImageIndex,
		Manifests:     []v1.Descriptor{},
	}
	for _, reference := range digests {
		manifest, err := m.manifestHandler.Get(req.Context(), repo, reference.String())
		if err != nil {
			return &regError{
				Status:  http.StatusNotFound,
				Code:    "INTERNAL_ERROR",
				Message: err.Error(),
			}
		}

		b, err := m.blobHandler.Get(req.Context(), repo, manifest.BlobDigest, false)
		if err != nil {
			return &regError{
				Status:  http.StatusNotFound,
				Code:    "BAD_REQUEST",
				Message: err.Error(),
			}
		}
		defer b.Close()
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, b); err != nil {
			return &regError{
				Status:  http.StatusInternalServerError,
				Code:    "INTERNAL_ERROR",
				Message: err.Error(),
			}
		}

		var refPointer struct {
			Subject *v1.Descriptor `json:"subject"`
		}
		_ = json.Unmarshal(buf.Bytes(), &refPointer)
		if refPointer.Subject == nil {
			continue
		}
		referenceDigest := refPointer.Subject.Digest
		if referenceDigest.String() != target {
			continue
		}
		// At this point, we know the current digest references the target
		var imageAsArtifact struct {
			Config struct {
				MediaType string `json:"mediaType"`
			} `json:"config"`
		}
		_ = json.Unmarshal(buf.Bytes(), &imageAsArtifact)
		im.Manifests = append(im.Manifests, v1.Descriptor{
			MediaType:    types.MediaType(manifest.ContentType),
			Size:         int64(buf.Len()),
			Digest:       reference,
			ArtifactType: imageAsArtifact.Config.MediaType,
		})
	}
	msg, _ := json.Marshal(&im)
	resp.Header().Set("Content-Length", fmt.Sprint(len(msg)))
	resp.Header().Set("Content-Type", string(types.OCIImageIndex))
	resp.WriteHeader(http.StatusOK)
	io.Copy(resp, bytes.NewReader([]byte(msg)))
	return nil
}
