package main

import (
	"context"
	"fmt"
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
	defer func() { _ = registry.Shutdown() }()
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

	kosmoz := types.UserAccount{
		Email:           "jakob.steiner@glasskube.eu",
		Name:            "Jakob Steiner",
		Password:        "asdasdasd",
		EmailVerifiedAt: util.PtrTo(time.Now()),
	}
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
		ComposeFileData: []byte("name: Hello World!\n"),
	}))
	util.Must(db.CreateApplicationVersion(ctx, &types.ApplicationVersion{
		ApplicationId:   app2.ID,
		Name:            "v2",
		ComposeFileData: []byte("name: Hello World!\n"),
	}))
	util.Must(db.CreateApplicationVersion(ctx, &types.ApplicationVersion{
		ApplicationId:   app2.ID,
		Name:            "v3",
		ComposeFileData: []byte("name: Hello World!\n"),
	}))

	app3 := types.Application{Name: "Wizard Security Graph", OrganizationID: org.ID, Type: types.DeploymentTypeDocker}
	util.Must(db.CreateApplication(ctx, &app3))
	av := types.ApplicationVersion{
		ApplicationId:   app3.ID,
		Name:            "v1",
		ComposeFileData: []byte("name: Hello World!\n"),
	}
	util.Must(db.CreateApplicationVersion(ctx, &av))

	podinfoApp := types.Application{Name: "Podinfo", OrganizationID: org.ID, Type: types.DepolymentTypeKubernetes}
	util.Must(db.CreateApplication(ctx, &podinfoApp))
	util.Must(db.CreateApplicationVersion(ctx, &types.ApplicationVersion{
		ApplicationId: podinfoApp.ID,
		Name:          "6.7.1",
		ChartType:     util.PtrTo(types.HelmChartTypeOCI),
		ChartUrl:      util.PtrTo("oci://ghcr.io/stefanprodan/charts/podinfo"),
		ChartVersion:  util.PtrTo("6.7.1"),
		ValuesFileData: []byte(
			"redis:\n  enabled: true\n" +
				"serviceAccount:\n  enabled: true\n",
		),
		TemplateFileData: []byte(
			"replicaCount: 1 # change this if needed\n" +
				"podDisruptionBudget:\n  # only applied if replicaCount > 1\n  maxUnavailable: 1\n",
		),
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

	for idx := range 3 {
		dt := types.DeploymentTargetWithCreatedBy{
			CreatedBy: &types.UserAccountWithUserRole{ID: pmig.ID},
			DeploymentTarget: types.DeploymentTarget{
				OrganizationID: org.ID,
				Name:           fmt.Sprintf("Deployment Target %v", idx),
				Type:           types.DeploymentTypeDocker,
			},
		}
		util.Must(db.CreateDeploymentTarget(ctx, &dt))
		util.Must(db.CreateDeploymentTargetStatus(ctx, &dt.DeploymentTarget, "running"))
		deployment := types.Deployment{
			DeploymentTargetId: dt.ID, ApplicationVersionId: av.ID,
		}
		util.Must(db.CreateDeployment(ctx, &deployment))
		now := time.Now().UTC()
		createdAt := now.Add(-1*24*time.Hour - 30*time.Minute)
		if idx == 2 {
			createdAt = createdAt.Add(12 * time.Hour)
		}
		for createdAt.Before(now) {
			util.Must(db.CreateDeploymentStatusWithCreatedAt(ctx, deployment.ID, "demo status", createdAt))
			if idx == 0 && createdAt.Hour() == 15 && createdAt.Minute() > 50 {
				createdAt = createdAt.Add(15 * time.Minute)
			} else if idx == 1 && createdAt.Hour() == 22 {
				createdAt = createdAt.Add(115 * time.Minute)
			} else {
				createdAt = createdAt.Add(5 * time.Second)
			}
		}
	}
}
