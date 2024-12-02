package auth

import (
	"context"
	"errors"
	"time"

	"github.com/glasskube/cloud/internal/db"
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

var JWTAuth = jwtauth.New("HS256", []byte("secret"), nil) // replace with secret key

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

func CurrentUser(ctx context.Context) (*types.UserAccount, error) {
	if token, _, err := jwtauth.FromContext(ctx); err != nil {
		return nil, err
	} else if user, err := db.GetUserAccountWithID(ctx, token.Subject()); err != nil {
		return nil, err
	} else {
		return user, nil
	}
}

func CurrentOrg(ctx context.Context) (*types.Organization, error) {
	if token, _, err := jwtauth.FromContext(ctx); err != nil {
		return nil, err
	} else if orgId, ok := token.Get(OrgIdKey); !ok {
		return nil, errors.New("missing org id in token")
	} else if org, err := db.GetOrganizationWithID(ctx, orgId.(string)); err != nil {
		return nil, err
	} else {
		return org, nil
	}
}
