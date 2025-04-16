package authn

import (
	"net/http"
)

type WithResponseHeaders interface {
	ResponseHeaders() http.Header
}
