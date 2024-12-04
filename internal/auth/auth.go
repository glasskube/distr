package auth

import (
	"context"
	"errors"
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
	UserNameKey  = "name"
	UserEmailKey = "email"
	OrgIdKey     = "org"
)

// JWTAuth is for generating/validating JWTs.
// Here we use symmetric encryption for now. This has the downside that the token can not be validated by clients,
// which should be OK for now.
//
// TODO: Maybe migrate to asymmetric encryption at some point.
var JWTAuth = jwtauth.New("HS256", env.JWTSecret(), nil)

func GenerateToken(user types.UserAccount, org types.Organization) (jwt.Token, string, error) {
	now := time.Now()
	claims := map[string]any{
		jwt.IssuedAtKey:   now,
		jwt.NotBeforeKey:  now,
		jwt.ExpirationKey: now.Add(defaultTokenExpiration),
		jwt.SubjectKey:    user.ID,
		UserNameKey:       user.Name,
		UserEmailKey:      user.Email,
		OrgIdKey:          org.ID,
	}
	return JWTAuth.Encode(claims)
}

func CurrentUserId(ctx context.Context) (string, error) {
	if token, _, err := jwtauth.FromContext(ctx); err != nil {
		return "", err
	} else {
		return token.Subject(), nil
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
