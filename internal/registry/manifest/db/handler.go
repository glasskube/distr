package db

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/glasskube/distr/internal/apierrors"
	"github.com/glasskube/distr/internal/auth"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/registry/manifest"
	"github.com/glasskube/distr/internal/types"
	"github.com/glasskube/distr/internal/util"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/jackc/pgx/v5"
)

type handler struct{}

func NewManifestHandler() manifest.ManifestHandler {
	return &handler{}
}

// Delete implements manifest.ManifestHandler.
func (h *handler) Delete(ctx context.Context, name string, reference string) error {
	panic("TODO: implement")
}

// Get implements manifest.ManifestHandler.
func (h *handler) Get(ctx context.Context, name string, reference string) (*manifest.Manifest, error) {
	if orgName, artifactName, err := splitName(name); err != nil {
		return nil, fmt.Errorf("%w: %w", manifest.ErrNameUnknown, err)
	} else if av, err := db.GetArtifactVersion(ctx, orgName, artifactName, reference); err != nil {
		if errors.Is(err, apierrors.ErrNotFound) {
			return nil, fmt.Errorf("%w: %w", manifest.ErrNameUnknown, err)
		}
		return nil, err
	} else {
		return &manifest.Manifest{
			BlobDigest:  v1.Hash(av.ManifestBlobDigest),
			ContentType: av.ManifestContentType,
		}, nil
	}
}

// List implements manifest.ManifestHandler.
func (h *handler) List(ctx context.Context, n int) ([]string, error) {
	auth := auth.ArtifactsAuthentication.Require(ctx)
	var artifacts []types.Artifact
	var err error
	if *auth.CurrentUserRole() == types.UserRoleCustomer {
		artifacts, err = db.GetArtifactsByLicenseOwnerID(ctx, auth.CurrentUserID())
	} else {
		artifacts, err = db.GetArtifactsByOrgID(ctx, *auth.CurrentOrgID())
	}
	if err != nil {
		return nil, err
	}
	result := make([]string, len(artifacts))
	for i, artifact := range artifacts {
		// TODO: use org slug instead
		result[i] = combineName(artifact.OrganizationID.String(), artifact.Name)
	}
	// TODO: move to DB
	if 0 < n && n < len(result) {
		result = result[:n]
	}
	return result, nil
}

// ListDigests implements manifest.ManifestHandler.
func (h *handler) ListDigests(ctx context.Context, name string) ([]v1.Hash, error) {
	if orgName, artifactName, err := splitName(name); err != nil {
		return nil, fmt.Errorf("%w: %w", manifest.ErrNameUnknown, err)
	} else {
		auth := auth.ArtifactsAuthentication.Require(ctx)
		var versions []types.ArtifactVersion
		var err error
		if *auth.CurrentUserRole() == types.UserRoleCustomer {
			versions, err = db.GetLicensedArtifactVersions(ctx, orgName, artifactName, auth.CurrentUserID())
		} else {
			versions, err = db.GetArtifactVersions(ctx, orgName, artifactName)
		}
		if err != nil {
			return nil, err
		}
		var result []v1.Hash
		for _, version := range versions {
			if h, err := v1.NewHash(version.Name); err != nil {
				continue
			} else {
				result = append(result, h)
			}
		}
		return result, nil
	}
}

// ListTags implements manifest.ManifestHandler.
func (h *handler) ListTags(ctx context.Context, name string, n int, last string) ([]string, error) {
	if orgName, artifactName, err := splitName(name); err != nil {
		return nil, fmt.Errorf("%w: %w", manifest.ErrNameUnknown, err)
	} else {
		auth := auth.ArtifactsAuthentication.Require(ctx)
		var versions []types.ArtifactVersion
		var err error
		if *auth.CurrentUserRole() == types.UserRoleCustomer {
			versions, err = db.GetLicensedArtifactVersions(ctx, orgName, artifactName, auth.CurrentUserID())
		} else {
			versions, err = db.GetArtifactVersions(ctx, orgName, artifactName)
		}
		if err != nil {
			return nil, err
		}
		var result []string
		for _, version := range versions {
			// only collect references that are NOT a hash
			if _, err := v1.NewHash(version.Name); err == nil {
				continue
			}
			if last == "" || version.Name > last {
				result = append(result, version.Name)
			}
		}
		if 0 < n && n < len(result) {
			result = result[:n]
		}
		return result, nil
	}
}

// Put implements manifest.ManifestHandler.
func (h *handler) Put(
	ctx context.Context,
	name, reference string,
	manifest manifest.Manifest,
	blobs []v1.Hash,
) error {
	auth := auth.ArtifactsAuthentication.Require(ctx)
	orgName, artifactName, err := splitName(name)
	if err != nil {
		return err
	}
	return db.RunTx(ctx, pgx.TxOptions{}, func(ctx context.Context) error {
		artifact, err := db.GetOrCreateArtifact(ctx, *auth.CurrentOrgID(), artifactName)
		if err != nil {
			return err
		}

		version := types.ArtifactVersion{
			CreatedByUserAccountID: util.PtrTo(auth.CurrentUserID()),
			Name:                   reference,
			ManifestBlobDigest:     types.Digest(manifest.BlobDigest),
			ManifestContentType:    manifest.ContentType,
			ArtifactID:             artifact.ID,
		}

		existingVersion, err := db.GetArtifactVersion(ctx, orgName, artifactName, reference)
		if err != nil {
			if !errors.Is(err, apierrors.ErrNotFound) {
				return err
			}
			if err := db.CreateArtifactVersion(ctx, &version); err != nil {
				return err
			}
		} else if existingVersion.ManifestBlobDigest != version.ManifestBlobDigest ||
			existingVersion.ManifestContentType != version.ManifestContentType {
			return fmt.Errorf("reference already exists with different manifest digest")
		} else {
			version = *existingVersion
		}

		for _, blob := range blobs {
			part := types.ArtifactVersionPart{
				ArtifactVersionID:  version.ID,
				ArtifactBlobDigest: types.Digest(blob),
			}
			if err := db.CreateArtifactVersionPart(ctx, &part); err != nil {
				return err
			}
		}
		return nil
	})
}

func splitName(name string) (string, string, error) {
	if parts := strings.SplitN(name, "/", 2); len(parts) != 2 {
		return "", "", fmt.Errorf("is not a valid artifact name: %v", name)
	} else {
		return parts[0], parts[1], nil
	}
}

func combineName(orgName, artifactName string) string {
	return path.Join(orgName, artifactName)
}
