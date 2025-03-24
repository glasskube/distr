package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/glasskube/distr/internal/env"
	"github.com/glasskube/distr/internal/util"
	"github.com/go-chi/chi/v5"
)

func InternalRouter(r chi.Router) {
	r.Handle("/environment", getFrontendEnvironmentHandler())
}

func getFrontendEnvironmentHandler() http.HandlerFunc {
	// precompute the json response
	frontendEnvJSON := util.Require(json.Marshal(struct {
		SentryDSN      *string `json:"sentryDsn,omitempty"`
		PosthogToken   *string `json:"posthogToken,omitempty"`
		PosthogAPIHost *string `json:"posthogApiHost,omitempty"`
		PosthogUIHost  *string `json:"posthogUiHost,omitempty"`
		RegistryHost   string  `json:"registryHost"`
	}{
		SentryDSN:      env.FrontendSentryDSN(),
		PosthogToken:   env.FrontendPosthogToken(),
		PosthogAPIHost: env.FrontendPosthogAPIHost(),
		PosthogUIHost:  env.FrontendPosthogUIHost(),
		RegistryHost:   env.RegistryHost(),
	}))
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(frontendEnvJSON)
	}
}
