package client

import (
	"encoding/json"
	"net/http"

	"github.com/distr-sh/distr/internal/httpstatus"
)

func JsonResponse[T any](resp *http.Response, err error) (T, error) {
	var result T
	resp, err = httpstatus.CheckStatus(resp, err)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	return result, json.NewDecoder(resp.Body).Decode(&result)
}
