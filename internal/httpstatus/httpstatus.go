package httpstatus

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var ErrHttpStatus = errors.New("non-ok http status")

func CheckStatus(r *http.Response, err error) (*http.Response, error) {
	if err != nil || StatusOK(r) {
		return r, err
	} else {
		if errorBody, err := io.ReadAll(r.Body); err == nil {
			return r, fmt.Errorf("%w: %v (%v)", ErrHttpStatus, r.Status, strings.TrimSpace(string(errorBody)))
		}
		return r, fmt.Errorf("%w: %v", ErrHttpStatus, r.Status)
	}
}

func StatusOK(r *http.Response) bool {
	return 200 <= r.StatusCode && r.StatusCode < 300
}
