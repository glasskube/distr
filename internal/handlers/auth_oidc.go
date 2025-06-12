package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/getsentry/sentry-go"
	"github.com/glasskube/distr/internal/apierrors"
	"github.com/glasskube/distr/internal/authjwt"
	context2 "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/env"
	"github.com/glasskube/distr/internal/types"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/microsoft"
)

const redirectToLoginOIDCFailed = "/login?reason=oidc-failed"

func AuthOIDCRouter(r chi.Router) {
	ctx := context.Background()
	if env.OIDCGoogleEnabled() {
		// TODO proper (lazy?) initialization (every server instance makes the requests)
		googleProvider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
		if err != nil {
			panic(err)
		}
		googleOidcConfig := &oidc.Config{
			ClientID: *env.OIDCGoogleClientID(),
		}
		googleIDTokenVerifier = googleProvider.Verifier(googleOidcConfig)
	}
	if env.OIDCMicrosoftEnabled() {
		microsoftProvider, err := oidc.NewProvider(ctx, fmt.Sprintf("https://login.microsoftonline.com/%v/v2.0",
			*env.OIDCMicrosoftTenantID()))
		if err != nil {
			panic(err)
		}
		config := &oidc.Config{
			ClientID: *env.OIDCMicrosoftClientID(),
		}
		microsoftIDTokenVerifier = microsoftProvider.Verifier(config)
	}

	r.Get("/{oidcProvider}", authLoginOidcHandler())
	r.Get("/{oidcProvider}/callback", authLoginOidcCallbackHandler)
}

var (
	googleIDTokenVerifier    *oidc.IDTokenVerifier
	microsoftIDTokenVerifier *oidc.IDTokenVerifier
)

func getRequestSchemeAndHost(r *http.Request) string {
	host := r.Host
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%v://%v", scheme, host)
}

func getRedirectURL(r *http.Request, provider types.OIDCProvider) string {
	return fmt.Sprintf("%v/api/v1/auth/oidc/%v/callback", getRequestSchemeAndHost(r), provider)
}

func getGithubOauth2Config(r *http.Request) *oauth2.Config {
	config := &oauth2.Config{
		ClientID:     *env.OIDCGithubClientID(),
		ClientSecret: *env.OIDCGithubClientSecret(),
		RedirectURL:  getRedirectURL(r, types.OIDCProviderGithub),
		Endpoint:     github.Endpoint,
		Scopes:       []string{oidc.ScopeOpenID, "email", "user:email"},
	}
	return config
}

func getGoogleOauth2Config(r *http.Request) *oauth2.Config {
	config := &oauth2.Config{
		ClientID:     *env.OIDCGoogleClientID(),
		ClientSecret: *env.OIDCGoogleClientSecret(),
		RedirectURL:  getRedirectURL(r, types.OIDCProviderGoogle),
		Endpoint:     google.Endpoint,
		Scopes:       []string{oidc.ScopeOpenID, "email"},
	}
	return config
}

func getMicrosoftOauth2Config(r *http.Request) *oauth2.Config {
	config := &oauth2.Config{
		ClientID:     *env.OIDCMicrosoftClientID(),
		ClientSecret: *env.OIDCMicrosoftClientSecret(),
		RedirectURL:  getRedirectURL(r, types.OIDCProviderMicrosoft),
		Endpoint:     microsoft.AzureADEndpoint(*env.OIDCMicrosoftTenantID()),
		Scopes:       []string{oidc.ScopeOpenID, "email"},
	}
	return config
}

func authLoginOidcHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var config *oauth2.Config
		provider := types.OIDCProvider(r.PathValue("oidcProvider"))
		if provider == types.OIDCProviderGithub && env.OIDCGithubEnabled() {
			config = getGithubOauth2Config(r)
		} else if provider == types.OIDCProviderGoogle && env.OIDCGoogleEnabled() {
			config = getGoogleOauth2Config(r)
		} else if provider == types.OIDCProviderMicrosoft && env.OIDCMicrosoftEnabled() {
			config = getMicrosoftOauth2Config(r)
		}
		if config != nil {
			// TODO send some state and in the callback make sure it matches
			http.Redirect(w, r, config.AuthCodeURL(""), http.StatusFound)
		} else {
			http.Redirect(w, r, redirectToLoginOIDCFailed, http.StatusFound)
		}
	}
}

func authLoginOidcCallbackHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := context2.GetLogger(ctx)

	code := r.URL.Query().Get("code")
	// TODO get state too and check it against something

	provider := types.OIDCProvider(r.PathValue("oidcProvider"))
	var config *oauth2.Config
	var idTokenVerifier *oidc.IDTokenVerifier
	var email string
	var emailVerified bool

	if provider == types.OIDCProviderGithub && env.OIDCGithubEnabled() {
		config = getGithubOauth2Config(r)
	} else if provider == types.OIDCProviderGoogle && env.OIDCGoogleEnabled() {
		config = getGoogleOauth2Config(r)
		idTokenVerifier = googleIDTokenVerifier
	} else if provider == types.OIDCProviderMicrosoft && env.OIDCMicrosoftEnabled() {
		config = getMicrosoftOauth2Config(r)
		idTokenVerifier = microsoftIDTokenVerifier
	}

	if config == nil {
		http.Redirect(w, r, redirectToLoginOIDCFailed, http.StatusFound)
		return
	}

	log = log.With(zap.String("provider", string(provider)))
	tokenForCode, err := config.Exchange(ctx, code)
	if err != nil {
		sentry.GetHubFromContext(ctx).CaptureException(err)
		log.Error("token exchange failed", zap.Error(err))
		http.Redirect(w, r, redirectToLoginOIDCFailed, http.StatusFound)
		return
	}

	if idTokenVerifier != nil {
		// at google and microsoft we get an id_token containing the email address
		idTokenStr, ok := tokenForCode.Extra("id_token").(string)
		if !ok {
			sentry.GetHubFromContext(ctx).CaptureException(err)
			log.Error("id_token not found", zap.Error(err))
			http.Redirect(w, r, redirectToLoginOIDCFailed, http.StatusFound)
			return
		}
		idToken, err := idTokenVerifier.Verify(ctx, idTokenStr)
		if err != nil {
			sentry.GetHubFromContext(ctx).CaptureException(err)
			log.Error("failed to verify id_token", zap.Error(err))
			http.Redirect(w, r, redirectToLoginOIDCFailed, http.StatusFound)
			return
		} else {
			var claims struct {
				Email         string `json:"email"`
				EmailVerified bool   `json:"email_verified"`
			}
			if err := idToken.Claims(&claims); err != nil {
				sentry.GetHubFromContext(ctx).CaptureException(err)
				log.Error("failed to get token claims", zap.Error(err))
				http.Redirect(w, r, redirectToLoginOIDCFailed, http.StatusFound)
				return
			} else {
				email = claims.Email
				emailVerified = claims.EmailVerified
			}
		}
	} else if provider == types.OIDCProviderGithub {
		// github doesn't provide the id_token, we need to get the users email addresses via the API
		accessTokenStr, ok := tokenForCode.Extra("access_token").(string)
		if !ok {
			sentry.GetHubFromContext(ctx).CaptureException(err)
			log.Error("no access_token found", zap.Error(err))
			http.Redirect(w, r, redirectToLoginOIDCFailed, http.StatusFound)
			return
		}
		// TODO also test with organization account?
		if ghEmail, err := getEmailFromGithubAccessToken(accessTokenStr); err != nil {
			sentry.GetHubFromContext(ctx).CaptureException(err)
			log.Error("failed to get github emails", zap.Error(err))
			http.Redirect(w, r, redirectToLoginOIDCFailed, http.StatusFound)
			return
		} else {
			email = ghEmail
			emailVerified = true
		}
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
			http.Redirect(w, r, fmt.Sprintf("%v/login?jwt=%v", getRequestSchemeAndHost(r), tokenString), http.StatusFound)
			return nil
		}
	})
	if err != nil {
		sentry.GetHubFromContext(ctx).CaptureException(err)
		log.Warn("user login failed", zap.Error(err))
		http.Redirect(w, r, redirectToLoginOIDCFailed, http.StatusFound)
	}
}

func getEmailFromGithubAccessToken(accessToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch emails: %s", resp.Status)
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	for _, email := range emails {
		if email.Primary && email.Verified {
			return email.Email, nil
		}
	}
	return "", fmt.Errorf("no primary verified email found")
}
