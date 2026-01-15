package handlers

import (
	"net/http"
	"time"

	"github.com/distr-sh/distr/api"
	"github.com/distr-sh/distr/internal/auth"
	internalctx "github.com/distr-sh/distr/internal/context"
	"github.com/distr-sh/distr/internal/db"
	"github.com/distr-sh/distr/internal/mapping"
	"github.com/distr-sh/distr/internal/middleware"
	"github.com/getsentry/sentry-go"
	"github.com/oaswrap/spec/adapter/chiopenapi"
	"github.com/oaswrap/spec/option"
	"go.uber.org/zap"
)

func ContextRouter(r chiopenapi.Router) {
	r.WithOptions(option.GroupHidden(true))
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
			User: mapping.UserAccountToAPI(
				auth.CurrentUser().AsUserAccountWithRole(*auth.CurrentUserRole(), auth.CurrentCustomerOrgID(), joinDate),
			),
			Organization:      mapping.OrganizationToAPI(*auth.CurrentOrg()),
			AvailableContexts: orgs,
		})
	}
}
