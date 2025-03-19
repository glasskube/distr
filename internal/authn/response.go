package authn

import (
	"io"
	"net/http"
)

type WithResponseHeaders interface {
	ResponseHeaders() http.Header
}

type WithResponseStatus interface {
	ResponseStatus() int
}

type ResponseBodyWriter interface {
	WriteResponse(w io.Writer)
}
