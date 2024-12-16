package main

import (
	"context"
	"time"

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
	defer func() { util.Must(registry.Shutdown()) }()
	ctx = internalctx.WithDb(ctx, registry.GetDbPool())

	org := types.Organization{Name: "Glasskube"}
	util.Must(db.CreateOrganization(ctx, &org))

	pmig := types.UserAccount{
		Email:           "pmig@glasskube.com",
		Name:            "Philip Miglinci",
		Password:        "12345678",
		EmailVerifiedAt: util.PtrTo(time.Now()),
	}
	util.Must(security.HashPassword(&pmig))
	util.Must(db.CreateUserAccount(ctx, &pmig))
	util.Must(db.CreateUserAccountOrganizationAssignment(ctx, pmig.ID, org.ID, types.UserRoleVendor))

	pmigCustomer := types.UserAccount{
		Email:           "pmig+customer@glasskube.com",
		Name:            "Philip Miglinci",
		Password:        "12345678",
		EmailVerifiedAt: util.PtrTo(time.Now()),
	}
	util.Must(security.HashPassword(&pmigCustomer))
	util.Must(db.CreateUserAccount(ctx, &pmigCustomer))
	util.Must(db.CreateUserAccountOrganizationAssignment(ctx, pmigCustomer.ID, org.ID, types.UserRoleCustomer))

	appMarsBeta := types.Application{
		Name: "Fastest way to Mars Calculator (Beta)", OrganizationID: org.ID, Type: types.DeploymentTypeDocker,
	}
	util.Must(db.CreateApplication(ctx, &appMarsBeta))
	util.Must(db.CreateApplicationVersion(ctx, &types.ApplicationVersion{
		ApplicationId: appMarsBeta.ID, Name: "v0.1.0", ComposeFileData: util.PtrTo([]byte("name: Hello World!\n")),
	}))
	util.Must(db.CreateApplicationVersion(ctx, &types.ApplicationVersion{
		ApplicationId: appMarsBeta.ID, Name: "v0.2.0", ComposeFileData: util.PtrTo([]byte("name: Hello World!\n")),
	}))
	util.Must(db.CreateApplicationVersion(ctx, &types.ApplicationVersion{
		ApplicationId: appMarsBeta.ID, Name: "v0.3.0", ComposeFileData: util.PtrTo([]byte("name: Hello World!\n")),
	}))
	appMarsBetaV419 := types.ApplicationVersion{ApplicationId: appMarsBeta.ID, Name: "v4.1.9"}
	util.Must(db.CreateApplicationVersion(ctx, &appMarsBetaV419))

	appMarsStable := types.Application{
		Name: "Fastest way to Mars Calculator (Stable)", OrganizationID: org.ID, Type: types.DeploymentTypeDocker,
	}
	util.Must(db.CreateApplication(ctx, &appMarsStable))
	util.Must(db.CreateApplicationVersion(ctx, &types.ApplicationVersion{
		ApplicationId: appMarsStable.ID, Name: "v0.3.1", ComposeFileData: util.PtrTo([]byte("name: Hello World!\n")),
	}))

	appMarsLTS := types.Application{
		Name: "Fastest way to Mars Calculator (LTS)", OrganizationID: org.ID, Type: types.DeploymentTypeDocker,
	}
	util.Must(db.CreateApplication(ctx, &appMarsLTS))
	appMarsLTSV0299 := types.ApplicationVersion{ApplicationId: appMarsLTS.ID, Name: "v0.29.9"}
	util.Must(db.CreateApplicationVersion(ctx, &appMarsLTSV0299))

	appLaunchDashboard := types.Application{
		Name: "Launch Dashboard", OrganizationID: org.ID, Type: types.DepolymentTypeKubernetes,
	}
	util.Must(db.CreateApplication(ctx, &appLaunchDashboard))
	appLaunchDashboardV001 := types.ApplicationVersion{
		ApplicationId: appMarsLTS.ID, Name: "v0.0.1",
	}
	util.Must(db.CreateApplicationVersion(ctx, &appLaunchDashboardV001))

	gateSpace := types.DeploymentTargetWithCreatedBy{
		CreatedBy: &pmig,
		DeploymentTarget: types.DeploymentTarget{
			OrganizationID: org.ID,
			Name:           "Danube Aerospace",
			Type:           types.DeploymentTypeDocker,
			Geolocation:    &types.Geolocation{Lat: 48.191166, Lon: 16.3717293},
		},
	}
	util.Must(db.CreateDeploymentTarget(ctx, &gateSpace))
	util.Must(db.CreateDeploymentTargetStatus(ctx, &gateSpace.DeploymentTarget, "running"))
	util.Must(db.CreateDeployment(ctx, &types.Deployment{
		DeploymentTargetId: gateSpace.ID, ApplicationVersionId: appMarsBetaV419.ID,
	}))

	lumenOrbit := types.DeploymentTargetWithCreatedBy{
		CreatedBy: &pmig,
		DeploymentTarget: types.DeploymentTarget{
			OrganizationID: org.ID,
			Name:           "Lux Orbit",
			Type:           types.DeploymentTypeDocker,
			Geolocation:    &types.Geolocation{Lat: 47.6349832, Lon: -122.1410062},
		},
	}
	util.Must(db.CreateDeploymentTarget(ctx, &lumenOrbit))
	util.Must(db.CreateDeploymentTargetStatus(ctx, &lumenOrbit.DeploymentTarget, "running"))
	util.Must(db.CreateDeployment(ctx, &types.Deployment{
		DeploymentTargetId: lumenOrbit.ID, ApplicationVersionId: appMarsBetaV419.ID,
	}))

	albaOrbital := types.DeploymentTargetWithCreatedBy{
		CreatedBy: &pmig,
		DeploymentTarget: types.DeploymentTarget{
			OrganizationID: org.ID,
			Name:           "Space K",
			Type:           types.DepolymentTypeKubernetes,
			Geolocation:    &types.Geolocation{Lat: 55.8578177, Lon: -4.3687363},
		},
	}
	util.Must(db.CreateDeploymentTarget(ctx, &albaOrbital))
	util.Must(db.CreateDeploymentTargetStatus(ctx, &albaOrbital.DeploymentTarget, "running"))
	util.Must(db.CreateDeployment(ctx, &types.Deployment{
		DeploymentTargetId: albaOrbital.ID, ApplicationVersionId: appMarsLTSV0299.ID,
	}))

	founderCafe := types.DeploymentTargetWithCreatedBy{
		CreatedBy: &pmig,
		DeploymentTarget: types.DeploymentTarget{
			OrganizationID: org.ID,
			Name:           "Bay Space Corp",
			Type:           types.DeploymentTypeDocker,
			Geolocation:    &types.Geolocation{Lat: 37.76078, Lon: -122.3915258},
		},
	}
	util.Must(db.CreateDeploymentTarget(ctx, &founderCafe))
	util.Must(db.CreateDeploymentTargetStatus(ctx, &founderCafe.DeploymentTarget, "running"))
	util.Must(db.CreateDeployment(ctx, &types.Deployment{
		DeploymentTargetId: founderCafe.ID, ApplicationVersionId: appLaunchDashboardV001.ID,
	}))

	quindar := types.DeploymentTargetWithCreatedBy{
		CreatedBy: &pmig,
		DeploymentTarget: types.DeploymentTarget{
			OrganizationID: org.ID,
			Name:           "Red Target",
			Type:           types.DeploymentTypeDocker,
			Geolocation:    &types.Geolocation{Lat: 39.1929769, Lon: -105.2403348},
		},
	}
	util.Must(db.CreateDeploymentTarget(ctx, &quindar))
}
