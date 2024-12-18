package auth

import (
	"context"
	"errors"
	"maps"
	"time"

	"github.com/glasskube/cloud/internal/env"
	"github.com/glasskube/cloud/internal/types"
	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

const (
	defaultTokenExpiration = 24 * time.Hour
)

const (
	UserNameKey          = "name"
	UserEmailKey         = "email"
	UserEmailVerifiedKey = "email_verified"
	UserRoleKey          = "role"
	OrgIdKey             = "org"
	PasswordResetKey     = "password_reset"
)

// JWTAuth is for generating/validating JWTs.
// Here we use symmetric encryption for now. This has the downside that the token can not be validated by clients,
// which should be OK for now.
//
// TODO: Maybe migrate to asymmetric encryption at some point.
var JWTAuth = jwtauth.New("HS256", env.JWTSecret(), nil)

func GenerateDefaultToken(user types.UserAccount, org types.OrganizationWithUserRole) (jwt.Token, string, error) {
	return generateUserToken(user, &org, defaultTokenExpiration, nil)
}

func GenerateResetToken(user types.UserAccount) (jwt.Token, string, error) {
	return generateUserToken(user, nil, env.ResetTokenValidDuration(), map[string]any{PasswordResetKey: true})
}

func GenerateVerificationTokenValidFor(user types.UserAccount) (jwt.Token, string, error) {
	return generateUserToken(user, nil, env.InviteTokenValidDuration(), map[string]any{UserEmailVerifiedKey: true})
}

func generateUserToken(
	user types.UserAccount,
	org *types.OrganizationWithUserRole,
	validFor time.Duration,
	extraClaims map[string]any,
) (jwt.Token, string, error) {
	now := time.Now()
	claims := map[string]any{
		jwt.IssuedAtKey:      now,
		jwt.NotBeforeKey:     now,
		jwt.ExpirationKey:    now.Add(validFor),
		jwt.SubjectKey:       user.ID,
		UserNameKey:          user.Name,
		UserEmailKey:         user.Email,
		UserEmailVerifiedKey: user.EmailVerifiedAt != nil,
	}
	if org != nil {
		claims[UserRoleKey] = org.UserRole
		claims[OrgIdKey] = org.ID
	}
	maps.Copy(claims, extraClaims)
	return JWTAuth.Encode(claims)
}

func GenerateAgentTokenValidFor(targetId string, orgId string, validFor time.Duration) (jwt.Token, string, error) {
	now := time.Now()
	claims := map[string]any{
		jwt.IssuedAtKey:   now,
		jwt.NotBeforeKey:  now,
		jwt.ExpirationKey: now.Add(validFor),
		jwt.SubjectKey:    targetId,
		OrgIdKey:          orgId,
	}
	return JWTAuth.Encode(claims)
}

func CurrentUserId(ctx context.Context) (string, error) {
	return CurrentSubject(ctx)
}

func CurrentSubject(ctx context.Context) (string, error) {
	if token, _, err := jwtauth.FromContext(ctx); err != nil {
		return "", err
	} else {
		return token.Subject(), nil
	}
}

func CurrentUserRole(ctx context.Context) (types.UserRole, error) {
	if token, _, err := jwtauth.FromContext(ctx); err != nil {
		return "", err
	} else if userRole, ok := token.Get(UserRoleKey); !ok {
		return "", errors.New("missing user role in token")
	} else {
		return types.UserRole(userRole.(string)), nil
	}
}

func CurrentOrgId(ctx context.Context) (string, error) {
	if token, _, err := jwtauth.FromContext(ctx); err != nil {
		return "", err
	} else if orgId, ok := token.Get(OrgIdKey); !ok {
		return "", errors.New("missing org id in token")
	} else {
		return orgId.(string), nil
	}
}

func CurrentUserEmailVerified(ctx context.Context) (bool, error) {
	if token, _, err := jwtauth.FromContext(ctx); err != nil {
		return false, err
	} else if verified, ok := token.Get(UserEmailVerifiedKey); !ok {
		return false, nil
	} else {
		return verified.(bool), nil
	}
}
