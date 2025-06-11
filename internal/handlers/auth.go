package handlers

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/distr/api"
	"github.com/glasskube/distr/internal/apierrors"
	"github.com/glasskube/distr/internal/auth"
	"github.com/glasskube/distr/internal/authjwt"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/customdomains"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/env"
	"github.com/glasskube/distr/internal/mail"
	"github.com/glasskube/distr/internal/mailsending"
	"github.com/glasskube/distr/internal/mailtemplates"
	"github.com/glasskube/distr/internal/middleware"
	"github.com/glasskube/distr/internal/security"
	"github.com/glasskube/distr/internal/types"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/coreos/go-oidc"
)

var githubOauth2Config *oauth2.Config
var googleOauth2Config *oauth2.Config
var oidcConfig *oidc.Config
var idTokenVerifier *oidc.IDTokenVerifier

func AuthRouter(r chi.Router) {
	githubOauth2Config = &oauth2.Config{ // TODO proper initialization
		ClientID:     *env.OIDCGithubClientID(),
		ClientSecret: *env.OIDCGithubClientSecret(),
		RedirectURL:  "http://localhost:4200/api/v1/auth/login/github/callback",
		Endpoint:     github.Endpoint,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
	ctx := context.Background()

	/*idTokenVerifier = oidc.NewVerifier("https://github.com", oidc.NewRemoteKeySet(context.Background(),
		"https://github.com/login/oauth/.well-known/jwks.json"), &oidc.Config{
		ClientID: *env.OIDCGithubClientID(),
	})*/

	googleOauth2Config = &oauth2.Config{ // TODO proper initialization
		ClientID:     *env.OIDCGoogleClientID(),
		ClientSecret: *env.OIDCGoogleClientSecret(),
		RedirectURL:  "http://localhost:4200/api/v1/auth/login/google/callback",
		Endpoint:     google.Endpoint,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
	provider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		panic(err)
	}
	oidcConfig = &oidc.Config{
		ClientID: *env.OIDCGoogleClientID(),
	}
	idTokenVerifier = provider.Verifier(oidcConfig)

	r.Use(httprate.Limit(
		10,
		1*time.Minute,
		httprate.WithKeyFuncs(httprate.KeyByRealIP, httprate.KeyByEndpoint),
	))
	r.Post("/login", authLoginHandler)
	r.Get("/login/{oidcProvider}", authLoginOidcHandler())
	r.Get("/login/{oidcProvider}/callback", authLoginOidcCallbackHandler)
	r.Route("/register", func(r chi.Router) {
		r.Get("/", authRegisterGetHandler())
		r.Post("/", authRegisterHandler)
	})
	r.Post("/reset", authResetPasswordHandler)
	r.With(middleware.SentryUser, auth.Authentication.Middleware, middleware.RequireOrgAndRole).
		Post("/switch-context", authSwitchContextHandler())
}

func authSwitchContextHandler() func(writer http.ResponseWriter, request *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := internalctx.GetLogger(ctx)
		request, err := JsonBody[api.AuthSwitchContextRequest](w, r)
		if err != nil {
			return
		} else if request.OrganizationID == uuid.Nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		auth := auth.Authentication.Require(ctx)
		if *auth.CurrentOrgID() == request.OrganizationID {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if user, org, err := db.GetUserAccountAndOrg(
			ctx, auth.CurrentUserID(), request.OrganizationID, nil); errors.Is(err, apierrors.ErrNotFound) {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		} else if err != nil {
			sentry.GetHubFromContext(ctx).CaptureException(err)
			log.Error("context switch failed", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		} else if _, tokenString, err := authjwt.GenerateDefaultToken(user.AsUserAccount(), types.OrganizationWithUserRole{
			Organization: *org,
			UserRole:     user.UserRole,
		}); err != nil {
			sentry.GetHubFromContext(ctx).CaptureException(err)
			log.Error("failed to generate token", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		} else {
			RespondJSON(w, api.AuthLoginResponse{Token: tokenString})
		}
	}
}

func authLoginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	request, err := JsonBody[api.AuthLoginRequest](w, r)
	if err != nil {
		return
	}
	err = db.RunTx(ctx, func(ctx context.Context) error {
		user, err := db.GetUserAccountByEmail(ctx, request.Email)
		if errors.Is(err, apierrors.ErrNotFound) {
			http.Error(w, "invalid username or password", http.StatusBadRequest)
			return nil
		} else if err != nil {
			return err
		}
		log = log.With(zap.Any("userId", user.ID))
		if err = security.VerifyPassword(*user, request.Password); err != nil {
			http.Error(w, "invalid username or password", http.StatusBadRequest)
			return nil
		}

		var org types.OrganizationWithUserRole
		orgs, err := db.GetOrganizationsForUser(ctx, user.ID)
		if err != nil {
			return err
		} else if len(orgs) < 1 {
			org.Name = user.Email
			org.UserRole = types.UserRoleVendor
			if err := db.CreateOrganization(ctx, &org.Organization); err != nil {
				return err
			} else if err := db.CreateUserAccountOrganizationAssignment(
				ctx, user.ID, org.ID, org.UserRole); err != nil {
				return err
			}
		} else {
			org = orgs[0]
		}

		if _, tokenString, err := authjwt.GenerateDefaultToken(*user, org); err != nil {
			return fmt.Errorf("token creation failed: %w", err)
		} else if err = db.UpdateUserAccountLastLoggedIn(ctx, user.ID); err != nil {
			return err
		} else {
			RespondJSON(w, api.AuthLoginResponse{Token: tokenString})
			return nil
		}
	})
	if err != nil {
		sentry.GetHubFromContext(ctx).CaptureException(err)
		log.Warn("user login failed", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func authLoginOidcHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if provider := r.PathValue("oidcProvider"); provider == "" {
			http.Error(w, "missing provider", http.StatusBadRequest)
		} else {
			switch provider {
			case "github":
				if !env.OIDCGithubEnabled() {
					http.Error(w, "github login not enabled", http.StatusBadRequest)
					// TODO back to login with this error message
				} else {
					http.Redirect(w, r, githubOauth2Config.AuthCodeURL("STATE-TODO-LOL"), http.StatusFound) // TODO state
				}
			case "google":
				if !env.OIDCGoogleEnabled() {
					http.Error(w, "google login not enabled", http.StatusBadRequest)
					// TODO back to login with this error message
				} else {
					http.Redirect(w, r, googleOauth2Config.AuthCodeURL("STATE-TODO-LOL"), http.StatusFound) // TODO state
				}
			default:
				http.Error(w, "unknown provider", http.StatusBadRequest) // TODO back to login with this error message
			}
		}
	}
}

func authLoginOidcCallbackHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	code := r.URL.Query().Get("code")
	token, err := googleOauth2Config.Exchange(ctx, code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		// TODO back to login with this error message
		return
	}

	// fmt.Fprintf(w, "%v\n", token)
	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No id_token found", http.StatusInternalServerError)
		// TODO back to login with this error message
		return
	}

	// fmt.Fprintf(w, "%v\n", idToken)
	tkn, err := idTokenVerifier.Verify(ctx, idToken)
	if err != nil {
		http.Error(w, "Failed to verify ID token: "+err.Error(), http.StatusInternalServerError)
		// TODO back to login with this error message
		return
	} else {
		var claims struct {
			Email string `json:"email"`
		}
		if err := tkn.Claims(&claims); err != nil {
			http.Error(w, "Failed to get token claims: "+err.Error(), http.StatusInternalServerError)
			// TODO back to login with this error message
			return
		}

		log := internalctx.GetLogger(ctx)
		// login logic
		// TODO should we set the user to email_verified if logged in via social provider?
		err = db.RunTx(ctx, func(ctx context.Context) error {
			user, err := db.GetUserAccountByEmail(ctx, claims.Email)
			if errors.Is(err, apierrors.ErrNotFound) {
				http.Error(w, "invalid username or password", http.StatusBadRequest)
				// TODO back to login with this error message (or send to signup form with this email adress?)
				return nil
			} else if err != nil {
				return err
			}
			log = log.With(zap.Any("userId", user.ID))

			var org types.OrganizationWithUserRole
			orgs, err := db.GetOrganizationsForUser(ctx, user.ID)
			if err != nil {
				return err
			} else if len(orgs) < 1 {
				org.Name = user.Email
				org.UserRole = types.UserRoleVendor
				if err := db.CreateOrganization(ctx, &org.Organization); err != nil {
					return err
				} else if err := db.CreateUserAccountOrganizationAssignment(
					ctx, user.ID, org.ID, org.UserRole); err != nil {
					return err
				}
			} else {
				org = orgs[0]
			}

			if _, tokenString, err := authjwt.GenerateDefaultToken(*user, org); err != nil {
				return fmt.Errorf("token creation failed: %w", err)
			} else if err = db.UpdateUserAccountLastLoggedIn(ctx, user.ID); err != nil {
				return err
			} else {
				http.Redirect(w, r, fmt.Sprintf("http://localhost:4200/?jwt=%v", tokenString), http.StatusFound)
				// TODO cookie instead? otherwise needs support to login properly (not only session storage)
				// RespondJSON(w, api.AuthLoginResponse{Token: tokenString})
				return nil
			}
		})
		if err != nil {
			sentry.GetHubFromContext(ctx).CaptureException(err)
			log.Warn("user login failed", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

	}

	// fmt.Fprintf(w, "Login successful!")
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
		var org *types.Organization

		if err := db.RunTx(ctx, func(ctx context.Context) error {
			if err := security.HashPassword(&userAccount); err != nil {
				sentry.GetHubFromContext(ctx).CaptureException(err)
				w.WriteHeader(http.StatusInternalServerError)
				return err
			} else if org, err = db.CreateUserAccountWithOrganization(ctx, &userAccount); err != nil {
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

		if err := mailsending.SendUserVerificationMail(ctx, userAccount, *org); err != nil {
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
	} else if orgs, err := db.GetOrganizationsForUser(ctx, user.ID); err != nil {
		log.Error("could not send reset mail", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
	} else if _, token, err := authjwt.GenerateResetToken(*user); err != nil {
		log.Error("could not send reset mail", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
	} else {
		var org *types.Organization
		mailOpts := []mail.MailOpt{
			mail.To(user.Email),
			mail.Subject("Password reset"),
		}
		if len(orgs) > 0 {
			org = &orgs[0].Organization
			if from, err := customdomains.EmailFromAddressParsedOrDefault(*org); err == nil {
				mailOpts = append(mailOpts, mail.From(*from))
			} else {
				log.Warn("error parsing custom from address", zap.Error(err))
			}
		}
		mailOpts = append(mailOpts, mail.HtmlBodyTemplate(mailtemplates.PasswordReset(*user, org, token)))
		if err := mailer.Send(ctx, mail.New(mailOpts...)); err != nil {
			log.Warn("could not send reset mail", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, "something went wrong", http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}
