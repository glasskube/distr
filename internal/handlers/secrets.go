package handlers

import (
	"errors"
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/distr/api"
	"github.com/glasskube/distr/internal/apierrors"
	"github.com/glasskube/distr/internal/auth"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/mapping"
	"github.com/glasskube/distr/internal/middleware"
	"github.com/oaswrap/spec/adapter/chiopenapi"
	"github.com/oaswrap/spec/option"
	"go.uber.org/zap"
)

func SecretsRouter(r chiopenapi.Router) {
	r.WithOptions(option.GroupTags("Secrets"))

	r.Use(middleware.RequireOrgAndRole)

	r.Get("/", getSecretsHandler()).
		With(option.Description("List all secrets")).
		With(option.Response(http.StatusOK, []api.SecretWithoutValue{}))

	r.Route("/{key}", func(r chiopenapi.Router) {
		r.Use(middleware.RequireReadWriteOrAdmin)

		r.Put("/", putSecretHandler()).
			With(option.Description("Create or update a secret")).
			With(option.Request(api.CreateUpdateSecretRequest{})).
			With(option.Response(http.StatusOK, api.SecretWithoutValue{}))
		r.Delete("/", deleteSecretHandler()).
			With(option.Description("Delete a secret")).
			With(option.Request(api.DeleteSecretRequest{}))
	})
}

func getSecretsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		auth := auth.Authentication.Require(ctx)

		secrets, err := db.GetSecrets(ctx, *auth.CurrentOrgID(), auth.CurrentCustomerOrgID())

		if err != nil {
			internalctx.GetLogger(ctx).Error("failed to get secrets", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		} else {
			RespondJSON(w, mapping.List(secrets, mapping.SecretToAPI))
		}
	}
}

func putSecretHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		auth := auth.Authentication.Require(ctx)
		body, err := JsonBody[api.CreateUpdateSecretRequest](w, r)
		if err != nil {
			return
		}
		body.Key = r.PathValue("key")

		secret, err := db.CreateOrUpdateSecret(
			ctx,
			body.Key,
			body.Value,
			*auth.CurrentOrgID(),
			auth.CurrentCustomerOrgID(),
			auth.CurrentUserID(),
		)

		if err != nil {
			internalctx.GetLogger(ctx).Error("failed to create/update secret", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		} else {
			RespondJSON(w, mapping.SecretToAPI(*secret))
		}
	}
}

func deleteSecretHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		auth := auth.Authentication.Require(ctx)
		key := r.PathValue("key")

		err := db.DeleteSecret(ctx, key, *auth.CurrentOrgID(), auth.CurrentCustomerOrgID())

		if err != nil {
			if errors.Is(err, apierrors.ErrNotFound) {
				http.Error(w, "Secret not found", http.StatusNotFound)
			} else {
				internalctx.GetLogger(ctx).Error("failed to delete secret", zap.Error(err))
				sentry.GetHubFromContext(ctx).CaptureException(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}
