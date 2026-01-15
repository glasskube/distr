package oidc

import (
	"fmt"
	"net/http"

	"github.com/distr-sh/distr/internal/handlerutil"
)

func getRedirectURL(r *http.Request, provider Provider) string {
	return fmt.Sprintf("%v/api/v1/auth/oidc/%v/callback", handlerutil.GetRequestSchemeAndHost(r), provider)
}
