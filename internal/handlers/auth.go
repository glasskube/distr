package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/glasskube/cloud/api"
	"github.com/glasskube/cloud/internal/apierrors"
	"github.com/glasskube/cloud/internal/auth"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/db"
	"github.com/glasskube/cloud/internal/mail"
	"github.com/glasskube/cloud/internal/mailtemplates"
	"github.com/glasskube/cloud/internal/security"
	"github.com/glasskube/cloud/internal/types"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

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
		} else if len(orgs) > 1 {
			log.Sugar().Warnf("user has %v organizations (currently only one is supported)", len(orgs))
		}
		org := orgs[0]
		if _, tokenString, err := auth.GenerateToken(*user, *org); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Warn("token creation failed", zap.Error(err))
		} else {
			_ = json.NewEncoder(w).Encode(api.AuthLoginResponse{Token: tokenString})
		}
	}
}

func authRegisterHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
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

		if err := db.RunTx(ctx, pgx.TxOptions{}, func(ctx context.Context) error {
			if err := security.HashPassword(&userAccount); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return err
			} else if _, err := db.CreateUserAccountWithOrganization(ctx, &userAccount); err != nil {
				if errors.Is(err, apierrors.ErrAlreadyExists) {
					w.WriteHeader(http.StatusBadRequest)
				} else {
					w.WriteHeader(http.StatusInternalServerError)
				}
				return err
			}
			return nil
		}); err != nil {
			log.Warn("user registration failed", zap.Error(err))
			return
		}

		// TODO generate jwt token containing the email
			mailer := internalctx.GetMailer(ctx)
			mail := mail.New(
				mail.To(userAccount.Email),
				mail.Subject("Verify your Glasskube Cloud Email"),
			mail.HtmlBodyTemplate(mailtemplates.VerifyEmailAtRegistration(userAccount, "/verify?jwt=asdfasdf")),
		)
		if err := mailer.Send(ctx, mail); err != nil {
			log.Error("could not send welcome mail", zap.Error(err))
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
