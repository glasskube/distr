package handlers

import (
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/distr/api"
	"github.com/glasskube/distr/internal/auth"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func ContextRouter(r chi.Router) {
	r.With(middleware.RequireOrgAndRole).Get("/", getContextHandler)
}

func getContextHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	auth := auth.Authentication.Require(ctx)
	if orgs, err := db.GetOrganizationsForUser(ctx, auth.CurrentUserID()); err != nil {
		internalctx.GetLogger(ctx).Error("failed to get organizations", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		var joinDate time.Time
		for _, org := range orgs {
			if org.ID == *auth.CurrentOrgID() {
				joinDate = org.JoinedOrgAt
				break
			}
		}
		RespondJSON(w, api.ContextResponse{
			User: api.AsUserAccount(
				auth.CurrentUser().AsUserAccountWithRole(*auth.CurrentUserRole(), auth.CurrentCustomerOrgID(), joinDate),
			),
			Organization:      *auth.CurrentOrg(),
			AvailableContexts: orgs,
		})
	}
}
