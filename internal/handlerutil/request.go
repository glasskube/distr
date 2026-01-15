package handlerutil

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/distr-sh/distr/internal/env"
)

func GetRequestSchemeAndHost(r *http.Request) string {
	host := env.Host()
	scheme := "http"
	if strings.HasPrefix(host, "https") {
		scheme = "https"
	}
	return fmt.Sprintf("%v://%v", scheme, r.Host)
}
