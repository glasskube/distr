package agentclient

import (
	"errors"
	"fmt"
	"net/http"
)

var ErrHttpStatus = errors.New("non-ok http status")

func checkStatus(r *http.Response, err error) (*http.Response, error) {
	if err != nil || statusOK(r) {
		return r, err
	} else {
		return r, fmt.Errorf("%w: %v", ErrHttpStatus, r.Status)
	}
}

func statusOK(r *http.Response) bool {
	return 200 <= r.StatusCode && r.StatusCode < 300
}
