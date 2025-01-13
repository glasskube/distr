package main

import (
	"context"

	"github.com/glasskube/cloud/internal/migrations"
	"github.com/glasskube/cloud/internal/svc"
	"github.com/glasskube/cloud/internal/util"
	"github.com/spf13/pflag"
)

var (
	down bool
	to   uint
)

func init() {
	pflag.BoolVar(&down, "down", down, "run all down migrations")
	pflag.UintVar(&to, "to", to, "run all up/down migrations to reach specified schema revision")
	pflag.Parse()
	if to > 0 && down {
		panic("please use --to OR --down")
	}
}

func main() {
	ctx := context.Background()
	registry := util.Require(svc.NewDefault(ctx))
	defer func() { util.Must(registry.Shutdown()) }()
	if to > 0 {
		registry.GetLogger().Sugar().Infof("run migrations to schema version %v", to)
		util.Must(migrations.Migrate(registry.GetLogger(), to))
	} else if down {
		registry.GetLogger().Info("run DOWN migrations")
		util.Must(migrations.Down(registry.GetLogger()))
	} else {
		registry.GetLogger().Info("run UP migrations")
		util.Must(migrations.Up(registry.GetLogger()))
	}
}
