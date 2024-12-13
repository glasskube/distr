package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/glasskube/cloud/internal/middleware"
	"github.com/glasskube/cloud/internal/types"
)

func JsonBody[T any](w http.ResponseWriter, r *http.Request) (T, error) {
	var t T
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	return t, err
}

func RespondJSON(w http.ResponseWriter, data any) {
	w.Header().Add("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

var requireUserRoleVendor = middleware.UserRoleMiddleware(types.UserRoleVendor)
