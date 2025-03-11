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

	av := &types.ArtifactVersion{ArtifactID: artifact.ID, Name: "v1.3.3"}
	util.Must(db.CreateArtifactVersion(ctx, av))
}
