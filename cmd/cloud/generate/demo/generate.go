package main

import (
	"context"
	"time"

	"github.com/glasskube/cloud/api"

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

	lw := types.UserAccount{
		Email:           "lw@glasskube.com",
		Name:            "Louis Weston",
		Password:        "12345678",
		EmailVerifiedAt: util.PtrTo(time.Now()),
	}
	util.Must(security.HashPassword(&lw))
	util.Must(db.CreateUserAccount(ctx, &lw))
	util.Must(db.CreateUserAccountOrganizationAssignment(ctx, lw.ID, org.ID, types.UserRoleVendor))

	appMarsBeta := types.Application{
		Name: "Mars Travel Calc (Beta)", OrganizationID: org.ID, Type: types.DeploymentTypeDocker,
	}
	util.Must(db.CreateApplication(ctx, &appMarsBeta))
	util.Must(db.CreateApplicationVersion(ctx, &types.ApplicationVersion{
		ApplicationId: appMarsBeta.ID, Name: "v0.1.0", ComposeFileData: []byte("name: Hello World!\n"),
	}))
	util.Must(db.CreateApplicationVersion(ctx, &types.ApplicationVersion{
		ApplicationId: appMarsBeta.ID, Name: "v0.2.0", ComposeFileData: []byte("name: Hello World!\n"),
	}))
	util.Must(db.CreateApplicationVersion(ctx, &types.ApplicationVersion{
		ApplicationId: appMarsBeta.ID, Name: "v0.3.0", ComposeFileData: []byte("name: Hello World!\n"),
	}))
	appMarsBetaV419 := types.ApplicationVersion{ApplicationId: appMarsBeta.ID, Name: "v4.1.9"}
	util.Must(db.CreateApplicationVersion(ctx, &appMarsBetaV419))

	appMarsStable := types.Application{
		Name: "Mars Travel Calc (Stable)", OrganizationID: org.ID, Type: types.DeploymentTypeDocker,
	}
	util.Must(db.CreateApplication(ctx, &appMarsStable))
	util.Must(db.CreateApplicationVersion(ctx, &types.ApplicationVersion{
		ApplicationId: appMarsStable.ID, Name: "v0.3.1", ComposeFileData: []byte("name: Hello World!\n"),
	}))

	appMarsLTS := types.Application{
		Name: "Mars Travel Calc (LTS)", OrganizationID: org.ID, Type: types.DeploymentTypeDocker,
	}
	util.Must(db.CreateApplication(ctx, &appMarsLTS))
	appMarsLTSV0299 := types.ApplicationVersion{ApplicationId: appMarsLTS.ID, Name: "v0.29.9"}
	util.Must(db.CreateApplicationVersion(ctx, &appMarsLTSV0299))

	appLaunchDashboard := types.Application{
		Name: "Launch Dashboard", OrganizationID: org.ID, Type: types.DepolymentTypeKubernetes,
	}
	util.Must(db.CreateApplication(ctx, &appLaunchDashboard))
	appLaunchDashboardV001 := types.ApplicationVersion{
		ApplicationId: appLaunchDashboard.ID, Name: "v0.0.1",
	}
	util.Must(db.CreateApplicationVersion(ctx, &appLaunchDashboardV001))

	dashboardTest := types.DeploymentTargetWithCreatedBy{
		CreatedBy: &types.UserAccountWithUserRole{ID: pmig.ID},
		DeploymentTarget: types.DeploymentTarget{
			OrganizationID: org.ID,
			Name:           "pmig - Dashboard Testing",
			Type:           types.DeploymentTypeDocker,
		},
	}
	util.Must(db.CreateDeploymentTarget(ctx, &dashboardTest))
	util.Must(db.CreateDeploymentTargetStatus(ctx, &dashboardTest.DeploymentTarget, "running"))
	util.Must(db.CreateDeployment(ctx, &api.DeploymentRequest{
		DeploymentTargetId: dashboardTest.ID, ApplicationVersionId: appLaunchDashboardV001.ID,
	}))

	calculatorTest := types.DeploymentTargetWithCreatedBy{
		CreatedBy: &types.UserAccountWithUserRole{ID: pmig.ID},
		DeploymentTarget: types.DeploymentTarget{
			OrganizationID: org.ID,
			Name:           "pmig - Calculator Testing",
			Type:           types.DeploymentTypeDocker,
		},
	}
	util.Must(db.CreateDeploymentTarget(ctx, &calculatorTest))
	util.Must(db.CreateDeploymentTargetStatus(ctx, &calculatorTest.DeploymentTarget, "running"))
	util.Must(db.CreateDeployment(ctx, &api.DeploymentRequest{
		DeploymentTargetId: calculatorTest.ID, ApplicationVersionId: appMarsBetaV419.ID,
	}))

	danubeAerospace := types.UserAccount{
		Email:           "devops@danube-aerospace.at",
		Name:            "Danube Aerospace",
		Password:        "12345678",
		EmailVerifiedAt: util.PtrTo(time.Now()),
	}
	util.Must(security.HashPassword(&danubeAerospace))
	util.Must(db.CreateUserAccount(ctx, &danubeAerospace))
	util.Must(db.CreateUserAccountOrganizationAssignment(ctx, danubeAerospace.ID, org.ID, types.UserRoleCustomer))

	danubeAerospaceVienna := types.DeploymentTargetWithCreatedBy{
		CreatedBy: &types.UserAccountWithUserRole{ID: danubeAerospace.ID},
		DeploymentTarget: types.DeploymentTarget{
			OrganizationID: org.ID,
			Name:           "DA - Vienna DC",
			Type:           types.DeploymentTypeDocker,
			Geolocation:    &types.Geolocation{Lat: 48.191166, Lon: 16.3717293},
		},
	}
	util.Must(db.CreateDeploymentTarget(ctx, &danubeAerospaceVienna))
	util.Must(db.CreateDeploymentTargetStatus(ctx, &danubeAerospaceVienna.DeploymentTarget, "running"))
	util.Must(db.CreateDeployment(ctx, &api.DeploymentRequest{
		DeploymentTargetId: danubeAerospaceVienna.ID, ApplicationVersionId: appMarsBetaV419.ID,
	}))

	luxOrbit := types.UserAccount{
		Email:           "it@lux-orbit.uk",
		Name:            "Lux Orbit",
		Password:        "12345678",
		EmailVerifiedAt: util.PtrTo(time.Now()),
	}
	util.Must(security.HashPassword(&luxOrbit))
	util.Must(db.CreateUserAccount(ctx, &luxOrbit))
	util.Must(db.CreateUserAccountOrganizationAssignment(ctx, luxOrbit.ID, org.ID, types.UserRoleCustomer))

	luxOrbitCanada := types.DeploymentTargetWithCreatedBy{
		CreatedBy: &types.UserAccountWithUserRole{ID: luxOrbit.ID},
		DeploymentTarget: types.DeploymentTarget{
			OrganizationID: org.ID,
			Name:           "LO - Canadian Cluster",
			Type:           types.DeploymentTypeDocker,
			Geolocation:    &types.Geolocation{Lat: 47.6349832, Lon: -122.1410062},
		},
	}
	util.Must(db.CreateDeploymentTarget(ctx, &luxOrbitCanada))
	util.Must(db.CreateDeploymentTargetStatus(ctx, &luxOrbitCanada.DeploymentTarget, "running"))
	util.Must(db.CreateDeployment(ctx, &api.DeploymentRequest{
		DeploymentTargetId: luxOrbitCanada.ID, ApplicationVersionId: appMarsBetaV419.ID,
	}))

	spaceK := types.UserAccount{
		Email:           "admin@space-k.uk",
		Name:            "Space K",
		Password:        "12345678",
		EmailVerifiedAt: util.PtrTo(time.Now()),
	}
	util.Must(security.HashPassword(&spaceK))
	util.Must(db.CreateUserAccount(ctx, &spaceK))
	util.Must(db.CreateUserAccountOrganizationAssignment(ctx, spaceK.ID, org.ID, types.UserRoleCustomer))

	spaceKUKWest := types.DeploymentTargetWithCreatedBy{
		CreatedBy: &types.UserAccountWithUserRole{ID: spaceK.ID},
		DeploymentTarget: types.DeploymentTarget{
			OrganizationID: org.ID,
			Name:           "Space K - uk-west-1",
			Type:           types.DepolymentTypeKubernetes,
			Geolocation:    &types.Geolocation{Lat: 55.8578177, Lon: -4.3687363},
		},
	}
	util.Must(db.CreateDeploymentTarget(ctx, &spaceKUKWest))
	util.Must(db.CreateDeploymentTargetStatus(ctx, &spaceKUKWest.DeploymentTarget, "running"))
	util.Must(db.CreateDeployment(ctx, &api.DeploymentRequest{
		DeploymentTargetId: spaceKUKWest.ID, ApplicationVersionId: appMarsLTSV0299.ID,
	}))

	baySpaceCorp := types.UserAccount{
		Email:           "admin@bay-space.com",
		Name:            "Bay Space Corp",
		Password:        "12345678",
		EmailVerifiedAt: util.PtrTo(time.Now()),
	}
	util.Must(security.HashPassword(&baySpaceCorp))
	util.Must(db.CreateUserAccount(ctx, &baySpaceCorp))
	util.Must(db.CreateUserAccountOrganizationAssignment(ctx, baySpaceCorp.ID, org.ID, types.UserRoleCustomer))

	baySpaceOffice := types.DeploymentTargetWithCreatedBy{
		CreatedBy: &types.UserAccountWithUserRole{ID: baySpaceCorp.ID},
		DeploymentTarget: types.DeploymentTarget{
			OrganizationID: org.ID,
			Name:           "BSC - Office",
			Type:           types.DeploymentTypeDocker,
			Geolocation:    &types.Geolocation{Lat: 37.76078, Lon: -122.3915258},
		},
	}
	util.Must(db.CreateDeploymentTarget(ctx, &baySpaceOffice))
	util.Must(db.CreateDeploymentTargetStatus(ctx, &baySpaceOffice.DeploymentTarget, "running"))
	util.Must(db.CreateDeployment(ctx, &api.DeploymentRequest{
		DeploymentTargetId: baySpaceOffice.ID, ApplicationVersionId: appLaunchDashboardV001.ID,
	}))

	baySpaceWest := types.DeploymentTargetWithCreatedBy{
		CreatedBy: &types.UserAccountWithUserRole{ID: baySpaceCorp.ID},
		DeploymentTarget: types.DeploymentTarget{
			OrganizationID: org.ID,
			Name:           "BSC - us-central-1",
			Type:           types.DeploymentTypeDocker,
			Geolocation:    &types.Geolocation{Lat: 39.1929769, Lon: -105.2403348},
		},
	}
	util.Must(db.CreateDeploymentTarget(ctx, &baySpaceWest))
}
