package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/glasskube/distr/internal/env"

	"github.com/go-chi/httprate"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/distr/api"
	"github.com/glasskube/distr/internal/apierrors"
	"github.com/glasskube/distr/internal/authjwt"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/mail"
	"github.com/glasskube/distr/internal/mailsending"
	"github.com/glasskube/distr/internal/mailtemplates"
	"github.com/glasskube/distr/internal/security"
	"github.com/glasskube/distr/internal/types"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

func AuthRouter(r chi.Router) {
	r.Use(httprate.Limit(
		10,
		1*time.Minute,
		httprate.WithKeyFuncs(httprate.KeyByRealIP, httprate.KeyByEndpoint),
	))
	r.Post("/login", authLoginHandler)
	r.Route("/register", func(r chi.Router) {
		r.Get("/", authRegisterGetHandler())
		r.Post("/", authRegisterHandler)
	})
	r.Post("/reset", authResetPasswordHandler)
}

func authLoginHandler(w http.ResponseWriter, r *http.Request) {
	log := internalctx.GetLogger(r.Context())
	if request, err := JsonBody[api.AuthLoginRequest](w, r); err != nil {
		return
	} else if user, err := db.GetUserAccountByEmail(r.Context(), request.Email); errors.Is(err, apierrors.ErrNotFound) {
		http.Error(w, "invalid username or password", http.StatusBadRequest)
	} else if err != nil {
		sentry.GetHubFromContext(r.Context()).CaptureException(err)
		w.WriteHeader(http.StatusInternalServerError)
		log.Warn("user login failed", zap.Error(err))
	} else if err = security.VerifyPassword(*user, request.Password); err != nil {
		http.Error(w, "invalid username or password", http.StatusBadRequest)
	} else if orgs, err := db.GetOrganizationsForUser(r.Context(), user.ID); err != nil {
		sentry.GetHubFromContext(r.Context()).CaptureException(err)
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
		if _, tokenString, err := authjwt.GenerateDefaultToken(*user, *org); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Warn("token creation failed", zap.Error(err))
		} else {
			_ = json.NewEncoder(w).Encode(api.AuthLoginResponse{Token: tokenString})
		}
	}
}

func authRegisterGetHandler() http.HandlerFunc {
	ok := env.Registration() == env.RegistrationEnabled
	return func(w http.ResponseWriter, r *http.Request) {
		if !ok {
			w.WriteHeader(http.StatusForbidden)
		}
	}
}

func authRegisterHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)

	if env.Registration() == env.RegistrationDisabled {
		http.Error(w, "registration is disabled", http.StatusForbidden)
		return
	}

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
				sentry.GetHubFromContext(ctx).CaptureException(err)
				w.WriteHeader(http.StatusInternalServerError)
				return err
			} else if _, err = db.CreateUserAccountWithOrganization(ctx, &userAccount); err != nil {
				if errors.Is(err, apierrors.ErrAlreadyExists) {
					w.WriteHeader(http.StatusBadRequest)
				} else {
					sentry.GetHubFromContext(ctx).CaptureException(err)
					w.WriteHeader(http.StatusInternalServerError)
				}
				return err
			}
			return nil
		}); err != nil {
			log.Warn("user registration failed", zap.Error(err))
			return
		}

		if err := mailsending.SendUserVerificationMail(ctx, userAccount); err != nil {
			log.Warn("could not send verification mail", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
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
	} else if user, err := db.GetUserAccountByEmail(ctx, request.Email); err != nil {
		if errors.Is(err, apierrors.ErrNotFound) {
			log.Info("password reset for non-existing user", zap.String("email", request.Email))
			w.WriteHeader(http.StatusNoContent)
		} else {
			log.Warn("could not send reset mail", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, "something went wrong", http.StatusInternalServerError)
		}
	} else if _, token, err := authjwt.GenerateResetToken(*user); err != nil {
		log.Warn("could not send reset mail", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
	} else if err := mailer.Send(ctx, mail.New(
		mail.To(user.Email),
		mail.Subject("Password reset"),
		mail.HtmlBodyTemplate(mailtemplates.PasswordReset(*user, token)),
	)); err != nil {
		log.Warn("could not send reset mail", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}
