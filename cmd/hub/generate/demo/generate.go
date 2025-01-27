package main

import (
	"context"
	"time"

	"github.com/glasskube/distr/api"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/security"
	"github.com/glasskube/distr/internal/svc"
	"github.com/glasskube/distr/internal/types"
	"github.com/glasskube/distr/internal/util"
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
	util.Must(db.CreateApplication(ctx, &appMarsBeta, org.ID))
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
		Name: "Mars Travel Calc (Stable)", Type: types.DeploymentTypeDocker,
	}
	util.Must(db.CreateApplication(ctx, &appMarsStable, org.ID))
	util.Must(db.CreateApplicationVersion(ctx, &types.ApplicationVersion{
		ApplicationId: appMarsStable.ID, Name: "v0.3.1", ComposeFileData: []byte("name: Hello World!\n"),
	}))

	appMarsLTS := types.Application{
		Name: "Mars Travel Calc (LTS)", Type: types.DeploymentTypeDocker,
	}
	util.Must(db.CreateApplication(ctx, &appMarsLTS, org.ID))
	appMarsLTSV0299 := types.ApplicationVersion{ApplicationId: appMarsLTS.ID, Name: "v0.29.9"}
	util.Must(db.CreateApplicationVersion(ctx, &appMarsLTSV0299))

	appLaunchDashboard := types.Application{
		Name: "Launch Dashboard", Type: types.DepolymentTypeKubernetes,
	}
	util.Must(db.CreateApplication(ctx, &appLaunchDashboard, org.ID))
	appLaunchDashboardV001 := types.ApplicationVersion{
		ApplicationId: appLaunchDashboard.ID, Name: "v0.0.1",
	}
	util.Must(db.CreateApplicationVersion(ctx, &appLaunchDashboardV001))

	dashboardTest := types.DeploymentTargetWithCreatedBy{
		DeploymentTarget: types.DeploymentTarget{
			Name:           "pmig - Dashboard Testing",
			Type:           types.DeploymentTypeDocker,
			AgentVersionID: util.Require(db.GetCurrentAgentVersion(ctx)).ID,
		},
	}
	util.Must(db.CreateDeploymentTarget(ctx, &dashboardTest, org.ID, pmig.ID))
	util.Must(db.CreateDeploymentTargetStatus(ctx, &dashboardTest.DeploymentTarget, "running"))
	deplRequest := api.DeploymentRequest{
		DeploymentTargetId: dashboardTest.ID, ApplicationVersionId: appLaunchDashboardV001.ID,
	}
	util.Must(db.CreateDeployment(ctx, &deplRequest))
	_, err := db.CreateDeploymentRevision(ctx, &deplRequest)
	util.Must(err)

	calculatorTest := types.DeploymentTargetWithCreatedBy{
		DeploymentTarget: types.DeploymentTarget{
			Name:           "pmig - Calculator Testing",
			Type:           types.DeploymentTypeDocker,
			AgentVersionID: util.Require(db.GetCurrentAgentVersion(ctx)).ID,
		},
	}
	util.Must(db.CreateDeploymentTarget(ctx, &calculatorTest, org.ID, pmig.ID))
	util.Must(db.CreateDeploymentTargetStatus(ctx, &calculatorTest.DeploymentTarget, "running"))
	deplRequest = api.DeploymentRequest{
		DeploymentTargetId: calculatorTest.ID, ApplicationVersionId: appMarsBetaV419.ID,
	}
	util.Must(db.CreateDeployment(ctx, &deplRequest))
	_, err = db.CreateDeploymentRevision(ctx, &deplRequest)
	util.Must(err)

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
		DeploymentTarget: types.DeploymentTarget{
			Name:           "DA - Vienna DC",
			Type:           types.DeploymentTypeDocker,
			Geolocation:    &types.Geolocation{Lat: 48.191166, Lon: 16.3717293},
			AgentVersionID: util.Require(db.GetCurrentAgentVersion(ctx)).ID,
		},
	}
	util.Must(db.CreateDeploymentTarget(ctx, &danubeAerospaceVienna, org.ID, danubeAerospace.ID))
	util.Must(db.CreateDeploymentTargetStatus(ctx, &danubeAerospaceVienna.DeploymentTarget, "running"))
	deplRequest = api.DeploymentRequest{
		DeploymentTargetId: danubeAerospaceVienna.ID, ApplicationVersionId: appMarsBetaV419.ID,
	}
	util.Must(db.CreateDeployment(ctx, &deplRequest))
	_, err = db.CreateDeploymentRevision(ctx, &deplRequest)
	util.Must(err)

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
		DeploymentTarget: types.DeploymentTarget{
			Name:           "LO - Canadian Cluster",
			Type:           types.DeploymentTypeDocker,
			Geolocation:    &types.Geolocation{Lat: 47.6349832, Lon: -122.1410062},
			AgentVersionID: util.Require(db.GetCurrentAgentVersion(ctx)).ID,
		},
	}
	util.Must(db.CreateDeploymentTarget(ctx, &luxOrbitCanada, org.ID, luxOrbit.ID))
	util.Must(db.CreateDeploymentTargetStatus(ctx, &luxOrbitCanada.DeploymentTarget, "running"))
	deplRequest = api.DeploymentRequest{
		DeploymentTargetId: luxOrbitCanada.ID, ApplicationVersionId: appMarsBetaV419.ID,
	}
	util.Must(db.CreateDeployment(ctx, &deplRequest))
	_, err = db.CreateDeploymentRevision(ctx, &deplRequest)
	util.Must(err)

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
		DeploymentTarget: types.DeploymentTarget{
			Name:           "Space K - uk-west-1",
			Type:           types.DepolymentTypeKubernetes,
			Geolocation:    &types.Geolocation{Lat: 55.8578177, Lon: -4.3687363},
			AgentVersionID: util.Require(db.GetCurrentAgentVersion(ctx)).ID,
		},
	}
	util.Must(db.CreateDeploymentTarget(ctx, &spaceKUKWest, org.ID, spaceK.ID))
	util.Must(db.CreateDeploymentTargetStatus(ctx, &spaceKUKWest.DeploymentTarget, "running"))
	deplRequest = api.DeploymentRequest{
		DeploymentTargetId: spaceKUKWest.ID, ApplicationVersionId: appMarsLTSV0299.ID,
	}
	util.Must(db.CreateDeployment(ctx, &deplRequest))
	_, err = db.CreateDeploymentRevision(ctx, &deplRequest)
	util.Must(err)

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
		DeploymentTarget: types.DeploymentTarget{
			Name:           "BSC - Office",
			Type:           types.DeploymentTypeDocker,
			Geolocation:    &types.Geolocation{Lat: 37.76078, Lon: -122.3915258},
			AgentVersionID: util.Require(db.GetCurrentAgentVersion(ctx)).ID,
		},
	}
	util.Must(db.CreateDeploymentTarget(ctx, &baySpaceOffice, org.ID, baySpaceCorp.ID))
	util.Must(db.CreateDeploymentTargetStatus(ctx, &baySpaceOffice.DeploymentTarget, "running"))
	deplRequest = api.DeploymentRequest{
		DeploymentTargetId: baySpaceOffice.ID, ApplicationVersionId: appLaunchDashboardV001.ID,
	}
	util.Must(db.CreateDeployment(ctx, &deplRequest))
	_, err = db.CreateDeploymentRevision(ctx, &deplRequest)
	util.Must(err)

	baySpaceWest := types.DeploymentTargetWithCreatedBy{
		DeploymentTarget: types.DeploymentTarget{
			Name:           "BSC - us-central-1",
			Type:           types.DeploymentTypeDocker,
			Geolocation:    &types.Geolocation{Lat: 39.1929769, Lon: -105.2403348},
			AgentVersionID: util.Require(db.GetCurrentAgentVersion(ctx)).ID,
		},
	}
	util.Must(db.CreateDeploymentTarget(ctx, &baySpaceWest, org.ID, baySpaceCorp.ID))
}
