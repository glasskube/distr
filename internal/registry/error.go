// Copyright 2018 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package registry

import (
	"encoding/json"
	"net/http"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type regError struct {
	Status  int
	Code    string
	Message string
	Error   error
}

func (r *regError) Write(resp http.ResponseWriter) error {
	resp.WriteHeader(r.Status)

	type err struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	type wrap struct {
		Errors []err `json:"errors"`
	}
	return json.NewEncoder(resp).Encode(wrap{
		Errors: []err{
			{
				Code:    r.Code,
				Message: r.Message,
			},
		},
	})
}

// regErrInternal returns an internal server error.
func regErrInternal(err error) *regError {
	return &regError{
		Status:  http.StatusInternalServerError,
		Code:    "INTERNAL_SERVER_ERROR",
		Message: err.Error(),
		Error:   err,
	}
}

func regErrManifestInvalid(err error) *regError {
	return &regError{
		Status:  http.StatusBadRequest,
		Code:    "MANIFEST_INVALID",
		Message: err.Error(),
		Error:   err,
	}
}

var regErrBlobUnknown = &regError{
	Status:  http.StatusNotFound,
	Code:    "BLOB_UNKNOWN",
	Message: "Unknown blob",
}

var regErrUnsupported = &regError{
	Status:  http.StatusMethodNotAllowed,
	Code:    "UNSUPPORTED",
	Message: "Unsupported operation",
}

var regErrDigestMismatch = &regError{
	Status:  http.StatusBadRequest,
	Code:    "DIGEST_INVALID",
	Message: "digest does not match contents",
}

var regErrDigestInvalid = &regError{
	Status:  http.StatusBadRequest,
	Code:    "NAME_INVALID",
	Message: "invalid digest",
}

var regErrManifestUnknown = &regError{
	Status:  http.StatusNotFound,
	Code:    "MANIFEST_UNKNOWN",
	Message: "Unknown manifest",
}

var regErrNameUnknown = &regError{
	Status:  http.StatusNotFound,
	Code:    "NAME_UNKNOWN",
	Message: "Unknown name",
}

var regErrMethodUnknown = &regError{
	Status:  http.StatusBadRequest,
	Code:    "METHOD_UNKNOWN",
	Message: "We don't understand your method + url",
}

var regErrDenied = &regError{
	Status:  http.StatusForbidden,
	Code:    "DENIED",
	Message: "Access to the resource has been denied",
}

// inspired by chimiddleware.Recoverer
func registryRecoverer(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				if rvr == http.ErrAbortHandler {
					// we don't recover http.ErrAbortHandler so the response
					// to the client is aborted, this should not be logged
					panic(rvr)
				}

				chimiddleware.PrintPrettyStack(rvr)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
