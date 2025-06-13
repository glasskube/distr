package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/distr/internal/apierrors"
	"github.com/glasskube/distr/internal/authjwt"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/oidc"
	"github.com/glasskube/distr/internal/types"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const redirectToLoginOIDCFailed = "/login?reason=oidc-failed"

func AuthOIDCRouter(r chi.Router) {
	r.Get("/{oidcProvider}", authLoginOidcHandler)
	r.Get("/{oidcProvider}/callback", authLoginOidcCallbackHandler)
}

func authLoginOidcHandler(w http.ResponseWriter, r *http.Request) {
	provider := types.OIDCProvider(r.PathValue("oidcProvider"))
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	if state, err := db.CreateOIDCState(ctx); err != nil {
		sentry.GetHubFromContext(ctx).CaptureException(err)
		log.Error("OIDC state creation failed", zap.Error(err))
		http.Redirect(w, r, redirectToLoginOIDCFailed, http.StatusFound)
		return
	} else {
		redirectURL, err := oidc.GetAuthCodeURL(r, provider, state.String())
		if err != nil {
			http.Redirect(w, r, redirectToLoginOIDCFailed, http.StatusFound)
			return
		}
		http.Redirect(w, r, redirectURL, http.StatusFound)
	}
}

func authLoginOidcCallbackHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)

	if err := verifyOIDCState(r); err != nil {
		sentry.GetHubFromContext(ctx).CaptureException(err)
		log.Warn("could not verify OIDC state", zap.Error(err))
		http.Redirect(w, r, redirectToLoginOIDCFailed, http.StatusFound)
		return
	}

	provider := types.OIDCProvider(r.PathValue("oidcProvider"))
	log = log.With(zap.String("provider", string(provider)))
	code := r.URL.Query().Get("code")

	email, emailVerified, err := oidc.GetEmailForCode(ctx, provider, code, r)
	if err != nil {
		sentry.GetHubFromContext(ctx).CaptureException(err)
		log.Error("OIDC email extraction failed", zap.Error(err))
		http.Redirect(w, r, redirectToLoginOIDCFailed, http.StatusFound)
		return
	}

	err = db.RunTx(ctx, func(ctx context.Context) error {
		user, err := db.GetUserAccountByEmail(ctx, email)
		if errors.Is(err, apierrors.ErrNotFound) {
			http.Redirect(w, r, "/register?reason=oidc-user-not-found&email="+email, http.StatusFound)
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
			// TODO deduplicate (regular login)
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

		if user.EmailVerifiedAt == nil && emailVerified {
			if err = db.UpdateUserAccountEmailVerified(ctx, user); err != nil {
				return err
			}
		}
		if _, tokenString, err := authjwt.GenerateDefaultToken(*user, org); err != nil {
			return fmt.Errorf("token creation failed: %w", err)
		} else if err = db.UpdateUserAccountLastLoggedIn(ctx, user.ID); err != nil {
			return err
		} else {
			http.Redirect(w, r, fmt.Sprintf("%v/login?jwt=%v", oidc.GetRequestSchemeAndHost(r), tokenString), http.StatusFound)
			return nil
		}
	})
	if err != nil {
		sentry.GetHubFromContext(ctx).CaptureException(err)
		log.Warn("user login failed", zap.Error(err))
		http.Redirect(w, r, redirectToLoginOIDCFailed, http.StatusFound)
	}
}

func verifyOIDCState(r *http.Request) error {
	if state, err := uuid.Parse(r.URL.Query().Get("state")); err != nil {
		return err
	} else if createdAt, err := db.DeleteOIDCState(r.Context(), state); err != nil {
		return err
	} else if createdAt.Before(time.Now().UTC().Add(-1 * time.Minute)) {
		return fmt.Errorf("got an OIDC state that is too old: %v, created_at: %v, now: %v",
			state, createdAt, time.Now().UTC())
	}
	return nil
}
