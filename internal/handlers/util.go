package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func JsonBody[T any](w http.ResponseWriter, r *http.Request) (T, error) {
	var t T
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "bad json")
	}
	return t, err
}
