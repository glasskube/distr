package main

import (
	"context"

	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/db"
	"github.com/glasskube/cloud/internal/security"
	"github.com/glasskube/cloud/internal/svc"
	"github.com/glasskube/cloud/internal/types"
	"github.com/glasskube/cloud/internal/util"
)

func main() {
	ctx := context.Background()
	registry := util.Require(svc.NewDefault(ctx))
	defer func() { _ = registry.Shutdown() }()
	ctx = internalctx.WithDb(ctx, registry.GetDbPool())

	org := types.Organization{Name: "Glasskube"}
	util.Must(db.CreateOrganization(ctx, &org))

	pmig := types.UserAccount{Email: "pmig@glasskube.com", Name: "Philip Miglinci", Password: "12345678"}
	util.Must(security.HashPassword(&pmig))
	util.Must(db.CreateUserAccount(ctx, &pmig))
	util.Must(db.CreateUserAccountOrganizationAssignment(ctx, pmig.ID, org.ID, types.UserRoleVendor))

	kosmoz := types.UserAccount{Email: "jakob.steiner@glasskube.eu", Name: "Jakob Steiner", Password: "asdasdasd"}
	util.Must(security.HashPassword(&kosmoz))
	util.Must(db.CreateUserAccount(ctx, &kosmoz))
	util.Must(db.CreateUserAccountOrganizationAssignment(ctx, kosmoz.ID, org.ID, types.UserRoleCustomer))

	app1 := types.Application{Name: "ASAN Mars Explorer", OrganizationID: org.ID, Type: types.DeploymentTypeDocker}
	util.Must(db.CreateApplication(ctx, &app1))
	util.Must(db.CreateApplicationVersion(ctx, &types.ApplicationVersion{
		ApplicationId: app1.ID,
		Name:          "v4.2.0",
	}))

	app2 := types.Application{Name: "Genome Graph Database", OrganizationID: org.ID, Type: types.DeploymentTypeDocker}
	util.Must(db.CreateApplication(ctx, &app2))
	util.Must(db.CreateApplicationVersion(ctx, &types.ApplicationVersion{
		ApplicationId:   app2.ID,
		Name:            "v1",
		ComposeFileData: util.PtrTo([]byte("name: Hello World!\n")),
	}))
	util.Must(db.CreateApplicationVersion(ctx, &types.ApplicationVersion{
		ApplicationId:   app2.ID,
		Name:            "v2",
		ComposeFileData: util.PtrTo([]byte("name: Hello World!\n")),
	}))
	util.Must(db.CreateApplicationVersion(ctx, &types.ApplicationVersion{
		ApplicationId:   app2.ID,
		Name:            "v3",
		ComposeFileData: util.PtrTo([]byte("name: Hello World!\n")),
	}))

	app3 := types.Application{Name: "Wizard Security Graph", OrganizationID: org.ID, Type: types.DeploymentTypeDocker}
	util.Must(db.CreateApplication(ctx, &app3))
	util.Must(db.CreateApplicationVersion(ctx, &types.ApplicationVersion{
		ApplicationId:   app3.ID,
		Name:            "v1",
		ComposeFileData: util.PtrTo([]byte("name: Hello World!\n")),
	}))

	dt1 := types.DeploymentTargetWithCreatedBy{
		CreatedBy: &types.UserAccountWithUserRole{ID: pmig.ID},
		DeploymentTarget: types.DeploymentTarget{
			OrganizationID: org.ID,
			Name:           "Space Center Austria",
			Type:           types.DeploymentTypeDocker,
			Geolocation:    &types.Geolocation{Lat: 48.1956026, Lon: 16.3633028},
		},
	}
	util.Must(db.CreateDeploymentTarget(ctx, &dt1))

	dt2 := types.DeploymentTargetWithCreatedBy{
		CreatedBy: &types.UserAccountWithUserRole{ID: kosmoz.ID},
		DeploymentTarget: types.DeploymentTarget{
			OrganizationID: org.ID,
			Name:           "Edge Location",
			Type:           types.DeploymentTypeDocker,
		},
	}
	util.Must(db.CreateDeploymentTarget(ctx, &dt2))

	dt3 := types.DeploymentTargetWithCreatedBy{
		CreatedBy: &types.UserAccountWithUserRole{ID: kosmoz.ID},
		DeploymentTarget: types.DeploymentTarget{
			OrganizationID: org.ID,
			Name:           "580 Founders Café",
			Type:           types.DeploymentTypeDocker,
			Geolocation:    &types.Geolocation{Lat: 37.758781, Lon: -122.396882},
		},
	}
	util.Must(db.CreateDeploymentTarget(ctx, &dt3))
	util.Must(db.CreateDeploymentTargetStatus(ctx, &dt3.DeploymentTarget, "running"))
}
