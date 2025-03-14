package authn

import (
	"errors"
	"net/http"
)

// ErrNoAuthentication implies that the provider *did not* find relevant
// authentication information on the Request
var ErrNoAuthentication = errors.New("not authenticated")

// ErrBadAuthentication implies that the provider *did* find relevant
// authentication information on the Request but it is not valid
var ErrBadAuthentication = errors.New("bad authentication")

type HttpHeaderError struct {
	wrapped error
	headers map[string]string
}

func NewHttpHeaderError(err error, headers map[string]string) error {
	return &HttpHeaderError{
		wrapped: err,
		headers: headers,
	}
}

func (err *HttpHeaderError) Error() string {
	return err.wrapped.Error()
}

func (err *HttpHeaderError) Unwrap() error {
	return err.wrapped
}

func (err *HttpHeaderError) WriteTo(w http.ResponseWriter) {
	for k, v := range err.headers {
		w.Header().Set(k, v)
	}
}
