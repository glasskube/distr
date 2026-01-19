package handlers

import (
	"errors"
	"net/http"

	"github.com/distr-sh/distr/api"
	"github.com/distr-sh/distr/internal/apierrors"
	"github.com/distr-sh/distr/internal/auth"
	internalctx "github.com/distr-sh/distr/internal/context"
	"github.com/distr-sh/distr/internal/db"
	"github.com/distr-sh/distr/internal/mapping"
	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
)

func getDeploymentTargetNotesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		auth := auth.Authentication.Require(ctx)
		deploymentTarget := internalctx.GetDeploymentTarget(ctx)

		notes, err := db.GetDeploymentTargetNotes(ctx, deploymentTarget.ID, auth.CurrentCustomerOrgID())
		if err != nil {
			if errors.Is(err, apierrors.ErrNotFound) {
				// empty notes response
				RespondJSON(w, api.DeploymentTargetNotes{})
				return
			}
			internalctx.GetLogger(ctx).Error("failed to get deployment target notes", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, "Failed to get deployment target notes", http.StatusInternalServerError)
			return
		}

		RespondJSON(w, mapping.DeploymentTargetNotesToAPI(*notes))
	}
}

func putDeploymentTargetNotesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		auth := auth.Authentication.Require(ctx)
		deploymentTarget := internalctx.GetDeploymentTarget(ctx)

		request, err := JsonBody[api.DeploymentTargetNotesRequest](w, r)
		if err != nil {
			return
		}

		notes, err := db.CreateOrUpdateDeploymentTargetNotes(
			ctx,
			deploymentTarget.ID,
			auth.CurrentCustomerOrgID(),
			auth.CurrentUserID(),
			request.Notes,
		)
		if err != nil {
			internalctx.GetLogger(ctx).Error("failed to create or update deployment target notes", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, "Failed to create or update deployment target notes", http.StatusInternalServerError)
			return
		}

		RespondJSON(w, mapping.DeploymentTargetNotesToAPI(*notes))
	}
}
