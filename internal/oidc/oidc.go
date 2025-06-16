package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/glasskube/distr/internal/env"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/microsoft"
)

type Provider string

const (
	ProviderGithub    Provider = "github"
	ProviderGoogle    Provider = "google"
	ProviderMicrosoft Provider = "microsoft"
)

type EmailExtractorFunc func(context.Context, *oauth2.Token) (string, bool, error)

func verifiedIdTokenEmailExtractor(verifier *oidc.IDTokenVerifier) EmailExtractorFunc {
	return func(ctx context.Context, token *oauth2.Token) (string, bool, error) {
		idTokenStr, ok := token.Extra("id_token").(string)
		if !ok {
			return "", false, fmt.Errorf("id_token not found in token response")
		}
		idToken, err := verifier.Verify(ctx, idTokenStr)
		if err != nil {
			return "", false, fmt.Errorf("failed to verify id_token: %w", err)
		}
		var claims struct {
			Email         string `json:"email"`
			EmailVerified bool   `json:"email_verified"`
		}
		if err := idToken.Claims(&claims); err != nil {
			return "", false, fmt.Errorf("failed to parse id_token claims: %w", err)
		}
		return claims.Email, claims.EmailVerified, nil
	}
}

type providerContext struct {
	OAuth2Config   func(r *http.Request) *oauth2.Config
	EmailExtractor EmailExtractorFunc
}

type OIDCer struct {
	providers map[Provider]*providerContext
}

func NewOIDCer(ctx context.Context, log *zap.Logger) (*OIDCer, error) {
	p := make(map[Provider]*providerContext)
	if env.OIDCGoogleEnabled() {
		log.Info("initializing google OIDC")
		googleProvider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Google OIDC provider: %w", err)
		}
		googleOidcConfig := &oidc.Config{ClientID: *env.OIDCGoogleClientID()}
		googleVerifier := googleProvider.Verifier(googleOidcConfig)
		p[ProviderGoogle] = &providerContext{
			OAuth2Config:   getGoogleOauth2Config,
			EmailExtractor: verifiedIdTokenEmailExtractor(googleVerifier),
		}
	}
	if env.OIDCMicrosoftEnabled() {
		log.Info("initializing microsoft OIDC")
		microsoftProvider, err := oidc.NewProvider(ctx,
			fmt.Sprintf("https://login.microsoftonline.com/%v/v2.0", *env.OIDCMicrosoftTenantID()))
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Microsoft OIDC provider: %w", err)
		}
		microsoftOidcConfig := &oidc.Config{ClientID: *env.OIDCMicrosoftClientID()}
		microsoftVerifier := microsoftProvider.Verifier(microsoftOidcConfig)
		p[ProviderMicrosoft] = &providerContext{
			OAuth2Config:   getMicrosoftOauth2Config,
			EmailExtractor: verifiedIdTokenEmailExtractor(microsoftVerifier),
		}
	}
	if env.OIDCGithubEnabled() {
		log.Info("initializing github OIDC")
		p[ProviderGithub] = &providerContext{
			OAuth2Config:   getGithubOauth2Config,
			EmailExtractor: getEmailFromGithubAccessToken,
		}
	}
	return &OIDCer{providers: p}, nil
}

func getGoogleOauth2Config(r *http.Request) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     *env.OIDCGoogleClientID(),
		ClientSecret: *env.OIDCGoogleClientSecret(),
		RedirectURL:  getRedirectURL(r, ProviderGoogle),
		Endpoint:     google.Endpoint,
		Scopes:       []string{oidc.ScopeOpenID, "email"},
	}
}

func getMicrosoftOauth2Config(r *http.Request) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     *env.OIDCMicrosoftClientID(),
		ClientSecret: *env.OIDCMicrosoftClientSecret(),
		RedirectURL:  getRedirectURL(r, ProviderMicrosoft),
		Endpoint:     microsoft.AzureADEndpoint(*env.OIDCMicrosoftTenantID()),
		Scopes:       []string{oidc.ScopeOpenID, "email"},
	}
}

func getGithubOauth2Config(r *http.Request) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     *env.OIDCGithubClientID(),
		ClientSecret: *env.OIDCGithubClientSecret(),
		RedirectURL:  getRedirectURL(r, ProviderGithub),
		Endpoint:     github.Endpoint,
		Scopes:       []string{oidc.ScopeOpenID, "email", "user:email"},
	}
}

// GetEmailForCode exchanges the code for a token and extracts the user's email and verification status.
func (o *OIDCer) GetEmailForCode(
	ctx context.Context, provider Provider, code string, r *http.Request,
) (string, bool, error) {
	prov := o.providers[provider]
	if prov == nil || prov.OAuth2Config == nil {
		return "", false, fmt.Errorf("OIDC provider not configured: %s", provider)
	}
	token, err := prov.OAuth2Config(r).Exchange(ctx, code)
	if err != nil {
		return "", false, fmt.Errorf("token exchange failed: %w", err)
	}

	if email, verified, err := prov.EmailExtractor(ctx, token); err != nil {
		return "", false, err
	} else {
		return email, verified, nil
	}
}

func getEmailFromGithubAccessToken(ctx context.Context, token *oauth2.Token) (string, bool, error) {
	accessToken, ok := token.Extra("access_token").(string)
	if !ok {
		return "", false, fmt.Errorf("access_token not found in token response")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", false, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", false, fmt.Errorf("failed to fetch emails: %s", resp.Status)
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", false, err
	}

	for _, email := range emails {
		if email.Primary && email.Verified {
			return email.Email, true, nil
		}
	}
	return "", false, fmt.Errorf("no primary verified email found")
}

// GetAuthCodeURL returns the OIDC provider's AuthCodeURL for the given state and provider.
func (o *OIDCer) GetAuthCodeURL(r *http.Request, provider Provider, state string) (string, error) {
	prov := o.providers[provider]
	if prov == nil || prov.OAuth2Config == nil {
		return "", fmt.Errorf("OIDC provider not configured: %s", provider)
	}
	return prov.OAuth2Config(r).AuthCodeURL(state), nil
}
