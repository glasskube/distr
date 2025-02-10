package handlers

import (
	"fmt"
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/distr/internal/auth"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/types"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func ApplicationLicenseRouter(r chi.Router) {
	r.Get("/", getApplicationLicensesHandler())
}

func getApplicationLicensesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		auth := auth.Authentication.Require(ctx)
		log := internalctx.GetLogger(r.Context())

		var licenses []types.ApplicationLicenseWithVersions
		var err error
		switch *auth.CurrentUserRole() {
		case types.UserRoleCustomer:
			licenses, err = db.GetApplicationLicensesWithOwnerID(ctx, auth.CurrentUserID())
		case types.UserRoleVendor:
			licenses, err = db.GetApplicationLicensesWithOrganizationID(ctx, *auth.CurrentOrgID())
		default:
			panic(fmt.Sprintf("unknown user role: %v", auth.CurrentUserRole()))
		}

		if err != nil {
			sentry.GetHubFromContext(r.Context()).CaptureException(err)
			w.WriteHeader(http.StatusInternalServerError)
			log.Warn("could not get licenses", zap.Error(err))
		} else {
			RespondJSON(w, licenses)
		}
	}
}
