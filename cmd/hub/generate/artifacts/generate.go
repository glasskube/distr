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

	artifact, err := db.GetOrCreateArtifact(ctx, org.ID, "distr")
	util.Must(err)

	av133Digest := &types.ArtifactVersion{
		ArtifactID: artifact.ID,
		Name:       "sha256:ed86a5c340796c0b3ad534b0cb7260a0a64da75f3756be2bb9652fdf53239f23",
		ManifestBlobDigest: types.Digest{
			Algorithm: "sha256",
			Hex:       "ed86a5c340796c0b3ad534b0cb7260a0a64da75f3756be2bb9652fdf53239f23",
		},
		ManifestContentType: "something-about-image+json",
	}
	util.Must(db.CreateArtifactVersion(ctx, av133Digest))
	av133Tagged := &types.ArtifactVersion{
		ArtifactID: artifact.ID,
		Name:       "v1.3.3",
		ManifestBlobDigest: types.Digest{
			Algorithm: "sha256",
			Hex:       "ed86a5c340796c0b3ad534b0cb7260a0a64da75f3756be2bb9652fdf53239f23",
		},
		ManifestContentType: "something-about-image+json",
	}
	util.Must(db.CreateArtifactVersion(ctx, av133Tagged))

	baseBlob := types.Digest{Algorithm: "sha256", Hex: "729f5f5e1dc385b3ba35b1efd7bd362b32fa515d985c4834c23e91899e879a85"}
	baseBlob2 := types.Digest{Algorithm: "sha256", Hex: "8ae816f3bb9218c8fe6322cfbf7a9f8bf86c52c664471037978793496673dbb5"}

	av133DigestBaseBlob := &types.ArtifactVersionPart{ArtifactVersionID: av133Digest.ID, ArtifactBlobDigest: baseBlob}
	util.Must(db.CreateArtifactVersionPart(ctx, av133DigestBaseBlob))
	av133DigestBlob2 := &types.ArtifactVersionPart{ArtifactVersionID: av133Digest.ID, ArtifactBlobDigest: baseBlob2}
	util.Must(db.CreateArtifactVersionPart(ctx, av133DigestBlob2))

	av133TaggedBaseBlob := &types.ArtifactVersionPart{ArtifactVersionID: av133Tagged.ID, ArtifactBlobDigest: baseBlob}
	util.Must(db.CreateArtifactVersionPart(ctx, av133TaggedBaseBlob))
	av133TaggedBlob2 := &types.ArtifactVersionPart{ArtifactVersionID: av133Tagged.ID, ArtifactBlobDigest: baseBlob2}
	util.Must(db.CreateArtifactVersionPart(ctx, av133TaggedBlob2))

	av134Digest := &types.ArtifactVersion{
		ArtifactID: artifact.ID,
		Name:       "sha256:fb86709b0df242e856d37475d0c31c550cb84e036af1c572dd778b5e2f944189",
		ManifestBlobDigest: types.Digest{
			Algorithm: "sha256",
			Hex:       "fb86709b0df242e856d37475d0c31c550cb84e036af1c572dd778b5e2f944189",
		},
		ManifestContentType: "something-about-image+json",
	}
	util.Must(db.CreateArtifactVersion(ctx, av134Digest))
	av134Tagged := &types.ArtifactVersion{
		ArtifactID: artifact.ID,
		Name:       "v1.3.4",
		ManifestBlobDigest: types.Digest{
			Algorithm: "sha256",
			Hex:       "fb86709b0df242e856d37475d0c31c550cb84e036af1c572dd778b5e2f944189",
		},
		ManifestContentType: "something-about-image+json",
	}
	util.Must(db.CreateArtifactVersion(ctx, av134Tagged))
	av134Latest := &types.ArtifactVersion{
		ArtifactID: artifact.ID,
		Name:       "latest",
		ManifestBlobDigest: types.Digest{
			Algorithm: "sha256",
			Hex:       "fb86709b0df242e856d37475d0c31c550cb84e036af1c572dd778b5e2f944189",
		},
		ManifestContentType: "something-about-image+json",
	}
	util.Must(db.CreateArtifactVersion(ctx, av134Latest))

	av134DigestBaseBlob := &types.ArtifactVersionPart{ArtifactVersionID: av134Digest.ID, ArtifactBlobDigest: baseBlob}
	util.Must(db.CreateArtifactVersionPart(ctx, av134DigestBaseBlob))
	av134DigestBlob2 := &types.ArtifactVersionPart{
		ArtifactVersionID: av134Digest.ID,
		ArtifactBlobDigest: types.Digest{
			Algorithm: "sha256",
			Hex:       "a7116857fe3266b0feae3162b387674b5054516425bbccf5e5d575d5ccdc5124",
		},
	}
	util.Must(db.CreateArtifactVersionPart(ctx, av134DigestBlob2))

	av134TaggedBaseBlob := &types.ArtifactVersionPart{ArtifactVersionID: av134Tagged.ID, ArtifactBlobDigest: baseBlob}
	util.Must(db.CreateArtifactVersionPart(ctx, av134TaggedBaseBlob))
	av134TaggedBlob2 := &types.ArtifactVersionPart{
		ArtifactVersionID: av134Tagged.ID,
		ArtifactBlobDigest: types.Digest{
			Algorithm: "sha256",
			Hex:       "a7116857fe3266b0feae3162b387674b5054516425bbccf5e5d575d5ccdc5124",
		},
	}
	util.Must(db.CreateArtifactVersionPart(ctx, av134TaggedBlob2))

	av134LatestBaseBlob := &types.ArtifactVersionPart{ArtifactVersionID: av134Latest.ID, ArtifactBlobDigest: baseBlob}
	util.Must(db.CreateArtifactVersionPart(ctx, av134LatestBaseBlob))
	av134LatestBlob2 := &types.ArtifactVersionPart{
		ArtifactVersionID: av134Latest.ID,
		ArtifactBlobDigest: types.Digest{
			Algorithm: "sha256",
			Hex:       "a7116857fe3266b0feae3162b387674b5054516425bbccf5e5d575d5ccdc5124",
		},
	}
	util.Must(db.CreateArtifactVersionPart(ctx, av134LatestBlob2))
}
