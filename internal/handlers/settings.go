package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/glasskube/cloud/internal/util"

	"github.com/glasskube/cloud/api"
	"github.com/glasskube/cloud/internal/apierrors"
	"github.com/glasskube/cloud/internal/db"
	"github.com/glasskube/cloud/internal/security"
	"github.com/go-chi/chi/v5"
)

func SettingsRouter(r chi.Router) {
	r.Post("/user", updateUserSettingsHandler)
}

func updateUserSettingsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body, err := JsonBody[api.UpdateUserAccountRequest](w, r)
	if err != nil {
		return
	}
	user, err := db.GetCurrentUser(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user.Name = body.Name
	if body.Password != "" {
		user.Password = body.Password
		if err := security.HashPassword(user); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
	if body.EmailVerified && user.EmailVerifiedAt == nil {
		user.EmailVerifiedAt = util.PtrTo(time.Now())
	}

	// TODO demo data verify

	if err := db.UpateUserAccount(ctx, user); errors.Is(err, apierrors.ErrNotFound) {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		RespondJSON(w, user)
	}
}
