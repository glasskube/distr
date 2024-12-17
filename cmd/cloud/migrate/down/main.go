package main

import (
	"context"

	"github.com/glasskube/cloud/internal/migrations"
	"github.com/glasskube/cloud/internal/svc"
	"github.com/glasskube/cloud/internal/util"
)

func main() {
	ctx := context.Background()
	registry := util.Require(svc.NewDefault(ctx))
	defer func() { util.Must(registry.Shutdown()) }()
	util.Must(migrations.Down(registry.GetLogger()))
}
