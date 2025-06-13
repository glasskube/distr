package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/glasskube/distr/internal/env"
	"github.com/glasskube/distr/internal/types"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/microsoft"
)

type OIDCer struct {
	providers map[types.OIDCProvider]*providerContext
}

type providerContext struct {
	OAuth2Config    func(r *http.Request) *oauth2.Config
	IDTokenVerifier *oidc.IDTokenVerifier
}

func NewOIDCer(ctx context.Context, log *zap.Logger) (*OIDCer, error) {
	p := make(map[types.OIDCProvider]*providerContext)
	if env.OIDCGoogleEnabled() {
		log.Info("initializing google OIDC")
		googleProvider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Google OIDC provider: %w", err)
		}
		googleOidcConfig := &oidc.Config{ClientID: *env.OIDCGoogleClientID()}
		p[types.OIDCProviderGoogle] = &providerContext{
			OAuth2Config:    getGoogleOauth2Config,
			IDTokenVerifier: googleProvider.Verifier(googleOidcConfig),
		}
	}
	if env.OIDCMicrosoftEnabled() {
		log.Info("initializing microsoft OIDC")
		microsoftProvider, err := oidc.NewProvider(ctx,
			fmt.Sprintf("https://login.microsoftonline.com/%v/v2.0", *env.OIDCMicrosoftTenantID()))
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Microsoft OIDC provider: %w", err)
		}
		config := &oidc.Config{ClientID: *env.OIDCMicrosoftClientID()}
		p[types.OIDCProviderMicrosoft] = &providerContext{
			OAuth2Config:    getMicrosoftOauth2Config,
			IDTokenVerifier: microsoftProvider.Verifier(config),
		}
	}
	if env.OIDCGithubEnabled() {
		log.Info("initializing github OIDC")
		p[types.OIDCProviderGithub] = &providerContext{
			OAuth2Config:    getGithubOauth2Config,
			IDTokenVerifier: nil,
		}
	}
	return &OIDCer{providers: p}, nil
}

func getGoogleOauth2Config(r *http.Request) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     *env.OIDCGoogleClientID(),
		ClientSecret: *env.OIDCGoogleClientSecret(),
		RedirectURL:  getRedirectURL(r, types.OIDCProviderGoogle),
		Endpoint:     google.Endpoint,
		Scopes:       []string{oidc.ScopeOpenID, "email"},
	}
}

func getMicrosoftOauth2Config(r *http.Request) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     *env.OIDCMicrosoftClientID(),
		ClientSecret: *env.OIDCMicrosoftClientSecret(),
		RedirectURL:  getRedirectURL(r, types.OIDCProviderMicrosoft),
		Endpoint:     microsoft.AzureADEndpoint(*env.OIDCMicrosoftTenantID()),
		Scopes:       []string{oidc.ScopeOpenID, "email"},
	}
}

func getGithubOauth2Config(r *http.Request) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     *env.OIDCGithubClientID(),
		ClientSecret: *env.OIDCGithubClientSecret(),
		RedirectURL:  getRedirectURL(r, types.OIDCProviderGithub),
		Endpoint:     github.Endpoint,
		Scopes:       []string{oidc.ScopeOpenID, "email", "user:email"},
	}
}

// GetEmailForCode exchanges the code for a token and extracts the user's email and verification status.
func (o *OIDCer) GetEmailForCode(
	ctx context.Context, provider types.OIDCProvider, code string, r *http.Request,
) (string, bool, error) {
	prov := o.providers[provider]
	if prov == nil || prov.OAuth2Config == nil {
		return "", false, fmt.Errorf("OIDC provider not configured: %s", provider)
	}
	token, err := prov.OAuth2Config(r).Exchange(ctx, code)
	if err != nil {
		return "", false, fmt.Errorf("token exchange failed: %w", err)
	}

	if prov.IDTokenVerifier != nil {
		idTokenStr, ok := token.Extra("id_token").(string)
		if !ok {
			return "", false, fmt.Errorf("id_token not found in token response")
		}
		idToken, err := prov.IDTokenVerifier.Verify(ctx, idTokenStr)
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
	} else if provider == types.OIDCProviderGithub {
		accessTokenStr, ok := token.Extra("access_token").(string)
		if !ok {
			return "", false, fmt.Errorf("access_token not found in token response")
		}
		email, err := getEmailFromGithubAccessToken(accessTokenStr)
		if err != nil {
			return "", false, fmt.Errorf("failed to get email from github: %w", err)
		}
		return email, true, nil
	}
	return "", false, fmt.Errorf("unsupported OIDC provider: %s", provider)
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

// GetAuthCodeURL returns the OIDC provider's AuthCodeURL for the given state and provider.
func (o *OIDCer) GetAuthCodeURL(r *http.Request, provider types.OIDCProvider, state string) (string, error) {
	prov := o.providers[provider]
	if prov == nil || prov.OAuth2Config == nil {
		return "", fmt.Errorf("OIDC provider not configured: %s", provider)
	}
	return prov.OAuth2Config(r).AuthCodeURL(state), nil
}
