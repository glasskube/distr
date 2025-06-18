package oidc

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/glasskube/distr/internal/env"
)

func GetRequestSchemeAndHost(r *http.Request) string {
	host := env.Host()
	scheme := "http"
	if strings.HasPrefix(host, "https") {
		scheme = "https"
	}
	return fmt.Sprintf("%v://%v", scheme, r.Host)
}

func getRedirectURL(r *http.Request, provider Provider) string {
	return fmt.Sprintf("%v/api/v1/auth/oidc/%v/callback", GetRequestSchemeAndHost(r), provider)
}
