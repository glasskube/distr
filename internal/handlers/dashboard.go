package handlers

import (
	"net/http"
	"slices"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/distr/internal/auth"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/middleware"
	"github.com/glasskube/distr/internal/types"
	"github.com/glasskube/distr/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func DashboardRouter(r chi.Router) {
	r.With(requireUserRoleVendor, middleware.RequireOrgAndRole).Group(func(r chi.Router) {
		r.Get("/artifacts-by-customer", getArtifactsByCustomer)
	})
}

type DashboardArtifact struct {
	Artifact            types.Artifact                `json:"artifact"`
	LatestPulledVersion string                        `json:"latestPulledVersion"`
	AvailableVersions   []types.TaggedArtifactVersion `json:"availableVersions"`
}

type ArtifactsByCustomer struct {
	Customer  types.UserAccountWithUserRole `json:"customer"`
	Artifacts []DashboardArtifact           `json:"artifacts,omitempty"`
}

func getArtifactsByCustomer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	auth := auth.Authentication.Require(ctx)
	if customers, err :=
		db.GetUserAccountsByOrgID(ctx, *auth.CurrentOrgID(), util.PtrTo(types.UserRoleCustomer)); err != nil {
		log.Error("failed to get customers", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	} else if artifacts, err := db.GetArtifactsByOrgID(ctx, *auth.CurrentOrgID()); err != nil {
		log.Error("failed to get artifacts", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	} else {
		var result []ArtifactsByCustomer
		for _, customer := range customers {
			customerRes := ArtifactsByCustomer{Customer: customer}
			for _, artifact := range artifacts {
				if slices.Contains(artifact.DownloadedByUsers, customer.ID) {
					if latestPulled, err := db.GetLatestPullOfArtifactByUser(ctx, artifact.ID, customer.ID); err != nil {
						// TODO
						/*log.Error("failed to get latest artifact pull by user", zap.Error(err))
						sentry.GetHubFromContext(ctx).CaptureException(err)
						http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
						return*/
					} else {
						var licenseOwnerID *uuid.UUID
						if auth.CurrentOrg().HasFeature(types.FeatureLicensing) {
							licenseOwnerID = &customer.ID
						}
						if versions, err := db.GetVersionsForArtifact(ctx, artifact.ID, licenseOwnerID); err != nil {
							// TODO
							/*log.Error("failed to get latest artifact pull by user", zap.Error(err))
							sentry.GetHubFromContext(ctx).CaptureException(err)
							http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
							return*/
						} else {
							customerRes.Artifacts = append(customerRes.Artifacts, DashboardArtifact{
								Artifact:            artifact.Artifact,
								LatestPulledVersion: latestPulled,
								AvailableVersions:   versions,
							})
						}
					}
				}
			}
			result = append(result, customerRes)
		}
		RespondJSON(w, result)
	}
}
