package authjwt

import (
	"maps"
	"time"

	"github.com/glasskube/distr/api"

	"github.com/glasskube/distr/internal/env"
	"github.com/glasskube/distr/internal/types"
	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
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
	UserImageURLKey      = "image_url"
	OrgIdKey             = "org"
	PasswordResetKey     = "password_reset"

	audienceUserValue  = "user"
	audienceAgentValue = "agent"
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
	return generateUserToken(user, nil, env.ResetTokenValidDuration(), map[string]any{
		PasswordResetKey:     true,
		UserEmailVerifiedKey: true,
	})
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
		jwt.SubjectKey:       user.ID.String(),
		jwt.AudienceKey:      audienceUserValue,
		UserNameKey:          user.Name,
		UserEmailKey:         user.Email,
		UserEmailVerifiedKey: !env.UserEmailVerificationRequired() || user.EmailVerifiedAt != nil,
	}
	if user.ImageID != nil {
		claims[UserImageURLKey] = api.WithImageUrl(user.ImageID)
	}
	if org != nil {
		claims[UserRoleKey] = org.UserRole
		claims[OrgIdKey] = org.ID.String()
	}
	maps.Copy(claims, extraClaims)
	return JWTAuth.Encode(claims)
}

func GenerateAgentTokenValidFor(targetID, orgID uuid.UUID, validFor time.Duration) (jwt.Token, string, error) {
	now := time.Now()
	claims := map[string]any{
		jwt.IssuedAtKey:   now,
		jwt.NotBeforeKey:  now,
		jwt.ExpirationKey: now.Add(validFor),
		jwt.SubjectKey:    targetID.String(),
		jwt.AudienceKey:   audienceAgentValue,
		OrgIdKey:          orgID.String(),
	}
	return JWTAuth.Encode(claims)
}
