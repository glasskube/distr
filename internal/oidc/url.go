package oidc

import (
	"fmt"
	"net/http"
)

func GetRequestSchemeAndHost(r *http.Request) string {
	host := r.Host
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%v://%v", scheme, host)
}

func getRedirectURL(r *http.Request, provider Provider) string {
	return fmt.Sprintf("%v/api/v1/auth/oidc/%v/callback", GetRequestSchemeAndHost(r), provider)
}
