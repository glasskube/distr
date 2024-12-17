package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/glasskube/cloud/api"
	"github.com/glasskube/cloud/internal/apierrors"
	"github.com/glasskube/cloud/internal/auth"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/db"
	"github.com/glasskube/cloud/internal/env"
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
	r.Post("/reset", authResetPasswordHandler)
}

func authLoginHandler(w http.ResponseWriter, r *http.Request) {
	log := internalctx.GetLogger(r.Context())
	if request, err := JsonBody[api.AuthLoginRequest](w, r); err != nil {
		return
	} else if user, err := db.GetUserAccountWithEmail(r.Context(), request.Email); errors.Is(err, apierrors.ErrNotFound) {
		http.Error(w, "invalid username or password", http.StatusBadRequest)
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Warn("user login failed", zap.Error(err))
	} else if err = security.VerifyPassword(*user, request.Password); err != nil {
		http.Error(w, "invalid username or password", http.StatusBadRequest)
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
		var org *types.Organization

		if err := db.RunTx(ctx, pgx.TxOptions{}, func(ctx context.Context) error {
			if err := security.HashPassword(&userAccount); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return err
			} else if org, err = db.CreateUserAccountWithOrganization(ctx, &userAccount); err != nil {
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

		_, token, err := auth.GenerateVerificationTokenValidFor(
			userAccount,
			types.OrganizationWithUserRole{Organization: *org, UserRole: types.UserRoleVendor},
			env.InviteTokenValidDuration(),
		)
		if err != nil {
			log.Error("could not generate verification token for welcome mail", zap.Error(err))
		}

		mailer := internalctx.GetMailer(ctx)
		mail := mail.New(
			mail.To(userAccount.Email),
			mail.Subject("Verify your Glasskube Cloud Email"),
			mail.HtmlBodyTemplate(mailtemplates.VerifyEmail(userAccount, token)),
		)
		if err := mailer.Send(ctx, mail); err != nil {
			log.Error("could not send verification mail",
				zap.Error(err), zap.String("user", userAccount.Email), zap.String("token", token))
		} else {
			log.Info("verification mail has been sent", zap.String("user", userAccount.Email), zap.String("token", token))
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func authResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	mailer := internalctx.GetMailer(ctx)
	if request, err := JsonBody[api.AuthResetPasswordRequest](w, r); err != nil {
		return
	} else if err := request.Validate(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if user, err := db.GetUserAccountWithEmail(ctx, request.Email); err != nil {
		if errors.Is(err, apierrors.ErrNotFound) {
			log.Info("password reset for non-existing user", zap.String("email", request.Email))
			w.WriteHeader(http.StatusNoContent)
		} else {
			log.Warn("could not send reset mail", zap.Error(err))
			http.Error(w, "something went wrong", http.StatusInternalServerError)
		}
	} else if _, token, err := auth.GenerateResetToken(*user); err != nil {
		log.Warn("could not send reset mail", zap.Error(err))
		http.Error(w, "something went wrong", http.StatusInternalServerError)
	} else if err := mailer.Send(ctx, mail.New(
		mail.To(user.Email),
		mail.Subject("Password reset"),
		mail.HtmlBodyTemplate(mailtemplates.PasswordReset(*user, token)),
	)); err != nil {
		log.Warn("could not send reset mail", zap.Error(err))
		http.Error(w, "something went wrong", http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}
