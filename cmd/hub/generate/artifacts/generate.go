package main

import (
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

/*
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
		ArtifactID: artifact.ID,
		Name: "sha256:blob-of-v1.3.3",
		ManifestBlobDigest: "sha256:blob-of-v1.3.3", // I guess it would be the same as the name
		ManifestContentType: "something-about-image+json",
	}
	util.Must(db.CreateArtifactVersion(ctx, av133Digest))
	av133 := &types.ArtifactVersion{
		ArtifactID: artifact.ID,
		Name: "v1.3.3",
		ManifestBlobDigest: "sha256:blob-of-v1.3.3",
		ManifestContentType: "something-about-image+json",
	}
	util.Must(db.CreateArtifactVersion(ctx, av133))

	baseBlob := &types.ArtifactBlob{Name: "base0"}
	util.Must(db.CreateArtifactBlob(ctx, baseBlob))

	blob133 := &types.ArtifactBlob{Name: "blob1.3.3", IsLead: true}
	util.Must(db.CreateArtifactBlob(ctx, blob133))

	avp133_1 := &types.ArtifactVersionPart{
		ArtifactVersionID: av133.ID,
		ArtifactBlobID:    baseBlob.ID,
		HashMD5:           baseBlob.ID.String(),
		HashSha1:          baseBlob.ID.String(),
		HashSha256:        baseBlob.ID.String(),
		HashSha512:        baseBlob.ID.String(),
	}
	util.Must(db.CreateArtifactVersionPart(ctx, avp133_1))
	avp133_2 := &types.ArtifactVersionPart{
		ArtifactVersionID: av133.ID,
		ArtifactBlobID:    blob133.ID,
		HashMD5:           blob133.ID.String(),
		HashSha1:          blob133.ID.String(),
		HashSha256:        blob133.ID.String(),
		HashSha512:        blob133.ID.String(),
	}
	util.Must(db.CreateArtifactVersionPart(ctx, avp133_2))

	av134 := &types.ArtifactVersion{ArtifactID: artifact.ID, Name: "v1.3.4"}
	util.Must(db.CreateArtifactVersion(ctx, av134))

	avp134_1 := &types.ArtifactVersionPart{
		ArtifactVersionID: av134.ID,
		ArtifactBlobID:    baseBlob.ID,
		HashMD5:           baseBlob.ID.String(),
		HashSha1:          baseBlob.ID.String(),
		HashSha256:        baseBlob.ID.String(),
		HashSha512:        baseBlob.ID.String(),
	}
	util.Must(db.CreateArtifactVersionPart(ctx, avp134_1))

	blob134 := &types.ArtifactBlob{Name: "blob1.3.4", IsLead: true}
	util.Must(db.CreateArtifactBlob(ctx, blob134))

	avp134_2 := &types.ArtifactVersionPart{
		ArtifactVersionID: av134.ID,
		ArtifactBlobID:    blob134.ID,
		HashMD5:           blob134.ID.String(),
		HashSha1:          blob134.ID.String(),
		HashSha256:        blob134.ID.String(),
		HashSha512:        blob134.ID.String(),
	}
	util.Must(db.CreateArtifactVersionPart(ctx, avp134_2))
}
*/
