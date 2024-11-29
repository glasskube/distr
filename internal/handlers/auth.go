package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/glasskube/cloud/api"
	"github.com/glasskube/cloud/internal/apierrors"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/db"
	"github.com/glasskube/cloud/internal/security"
	"github.com/glasskube/cloud/internal/types"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"go.uber.org/zap"
)

var tokenAuth *jwtauth.JWTAuth

func init() {
	tokenAuth = jwtauth.New("HS256", []byte("secret"), nil) // replace with secret key
}

func AuthRouter(r chi.Router) {
	r.Post("/login", authLoginHandler)
	r.Post("/register", authRegisterHandler)
}

func authLoginHandler(w http.ResponseWriter, r *http.Request) {
	log := internalctx.GetLogger(r.Context())
	if request, err := JsonBody[api.AuthLoginRequest](w, r); err != nil {
		return
	} else if user, err := db.GetUserAccountWithEmail(r.Context(), request.Email); errors.Is(err, apierrors.ErrNotFound) {
		w.WriteHeader(http.StatusBadRequest)
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Warn("user login failed", zap.Error(err))
	} else if err = security.VerifyPassword(*user, request.Password); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "invalid username or password")
	} else if orgs, err := db.GetOrganizationsForUser(r.Context(), user.ID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Warn("user login failed", zap.Error(err))
	} else {
		if len(orgs) < 1 {
			w.WriteHeader(http.StatusInternalServerError)
			log.Error("user has no organizations")
		}
		org := orgs[0]
		claims := map[string]any{
			"sub":   user.ID,
			"name":  user.Name,
			"email": user.Email,
			"org":   org.ID,
		}
		if _, tokenString, err := tokenAuth.Encode(claims); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Warn("token creation failed", zap.Error(err))
		} else {
			_ = json.NewEncoder(w).Encode(api.AuthLoginResponse{Token: tokenString})
		}
	}
}

func authRegisterHandler(w http.ResponseWriter, r *http.Request) {
	log := internalctx.GetLogger(r.Context())
	if request, err := JsonBody[api.AuthRegistrationRequest](w, r); err != nil {
		return
	} else if err := request.Validate(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		userAccount := types.UserAccount{
			Name:     request.Name,
			Email:    request.Email,
			Password: request.Password,
		}
		if err := security.HashPassword(&userAccount); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else if _, err := db.CreateUserAccountWithOrganization(r.Context(), &userAccount); err != nil {
			log.Warn("user registration failed", zap.Error(err))
			if errors.Is(err, apierrors.ErrAlreadyExists) {
				w.WriteHeader(http.StatusBadRequest)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}
