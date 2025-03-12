package main

import (
	"context"

	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/svc"
	"github.com/glasskube/distr/internal/types"
	"github.com/glasskube/distr/internal/util"
	"github.com/google/uuid"
	"github.com/spf13/pflag"
)

var orgID uuid.UUID

func init() {
	var orgIDStr string
	pflag.StringVar(&orgIDStr, "org", "", "org ID")
	pflag.Parse()
	orgID = uuid.MustParse(orgIDStr)
}

func main() {
	ctx := context.Background()
	registry := util.Require(svc.NewDefault(ctx))
	defer func() { util.Must(registry.Shutdown()) }()
	ctx = internalctx.WithDb(ctx, registry.GetDbPool())

	org, err := db.GetOrganizationByID(ctx, orgID)
	util.Must(err)

	artifact, err := db.GetOrCreateArtifact(ctx, org, "distr")
	util.Must(err)

	av133Digest := &types.ArtifactVersion{
		ArtifactID:          artifact.ID,
		Name:                "sha256:blob-of-v1.3.3",
		ManifestBlobDigest:  "sha256:blob-of-v1.3.3", // I guess it would be the same as the name
		ManifestContentType: "something-about-image+json",
	}
	util.Must(db.CreateArtifactVersion(ctx, av133Digest))
	av133Tagged := &types.ArtifactVersion{
		ArtifactID:          artifact.ID,
		Name:                "v1.3.3",
		ManifestBlobDigest:  "sha256:blob-of-v1.3.3",
		ManifestContentType: "something-about-image+json",
	}
	util.Must(db.CreateArtifactVersion(ctx, av133Tagged))

	baseBlob := "sha256:base-blob"

	av133DigestBaseBlob := &types.ArtifactVersionPart{ArtifactVersionID: av133Digest.ID, ArtifactBlobDigest: baseBlob}
	util.Must(db.CreateArtifactVersionPart(ctx, av133DigestBaseBlob))
	av133DigestBlob2 := &types.ArtifactVersionPart{ArtifactVersionID: av133Digest.ID, ArtifactBlobDigest: "sha256:blob-2"}
	util.Must(db.CreateArtifactVersionPart(ctx, av133DigestBlob2))

	av133TaggedBaseBlob := &types.ArtifactVersionPart{ArtifactVersionID: av133Tagged.ID, ArtifactBlobDigest: baseBlob}
	util.Must(db.CreateArtifactVersionPart(ctx, av133TaggedBaseBlob))
	av133TaggedBlob2 := &types.ArtifactVersionPart{ArtifactVersionID: av133Tagged.ID, ArtifactBlobDigest: "sha256:blob-2"}
	util.Must(db.CreateArtifactVersionPart(ctx, av133TaggedBlob2))

	av134Digest := &types.ArtifactVersion{
		ArtifactID:          artifact.ID,
		Name:                "sha256:blob-of-v1.3.4",
		ManifestBlobDigest:  "sha256:blob-of-v1.3.4", // I guess it would be the same as the name
		ManifestContentType: "something-about-image+json",
	}
	util.Must(db.CreateArtifactVersion(ctx, av134Digest))
	av134Tagged := &types.ArtifactVersion{
		ArtifactID:          artifact.ID,
		Name:                "v1.3.4",
		ManifestBlobDigest:  "sha256:blob-of-v1.3.4",
		ManifestContentType: "something-about-image+json",
	}
	util.Must(db.CreateArtifactVersion(ctx, av134Tagged))
	av134Latest := &types.ArtifactVersion{
		ArtifactID:          artifact.ID,
		Name:                "latest",
		ManifestBlobDigest:  "sha256:blob-of-v1.3.4",
		ManifestContentType: "something-about-image+json",
	}
	util.Must(db.CreateArtifactVersion(ctx, av134Latest))

	av134DigestBaseBlob := &types.ArtifactVersionPart{ArtifactVersionID: av134Digest.ID, ArtifactBlobDigest: baseBlob}
	util.Must(db.CreateArtifactVersionPart(ctx, av134DigestBaseBlob))
	av134DigestBlob2 := &types.ArtifactVersionPart{ArtifactVersionID: av134Digest.ID,
		ArtifactBlobDigest: "sha256:blob-2-of-1.3.4"}
	util.Must(db.CreateArtifactVersionPart(ctx, av134DigestBlob2))

	av134TaggedBaseBlob := &types.ArtifactVersionPart{ArtifactVersionID: av134Tagged.ID, ArtifactBlobDigest: baseBlob}
	util.Must(db.CreateArtifactVersionPart(ctx, av134TaggedBaseBlob))
	av134TaggedBlob2 := &types.ArtifactVersionPart{ArtifactVersionID: av134Tagged.ID,
		ArtifactBlobDigest: "sha256:blob-2-of-1.3.4"}
	util.Must(db.CreateArtifactVersionPart(ctx, av134TaggedBlob2))

	av134LatestBaseBlob := &types.ArtifactVersionPart{ArtifactVersionID: av134Latest.ID, ArtifactBlobDigest: baseBlob}
	util.Must(db.CreateArtifactVersionPart(ctx, av134LatestBaseBlob))
	av134LatestBlob2 := &types.ArtifactVersionPart{ArtifactVersionID: av134Latest.ID,
		ArtifactBlobDigest: "sha256:blob-2-of-1.3.4"}
	util.Must(db.CreateArtifactVersionPart(ctx, av134LatestBlob2))
}
