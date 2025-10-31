package authinfo

import (
	"context"
	"fmt"

	"github.com/glasskube/distr/internal/authjwt"
	"github.com/glasskube/distr/internal/authn"
	"github.com/glasskube/distr/internal/types"
	"github.com/glasskube/distr/internal/util"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

func FromUserJWT(token jwt.Token) (*SimpleAuthInfo, error) {
	// TODO: Check if this function is still needed
	// FIXME: SimpleAuthInfo does not contain customer org ID here
	var result SimpleAuthInfo
	if parsedSub, err := uuid.Parse(token.Subject()); err != nil {
		return nil, fmt.Errorf("JWT subject is invalid: %w", err)
	} else {
		result.userID = parsedSub
	}
	result.rawToken = token
	if userEmail, ok := token.Get(authjwt.UserEmailKey); ok {
		result.userEmail = userEmail.(string)
	}
	if userRole, ok := token.Get(authjwt.UserRoleKey); ok {
		result.userRole = util.PtrTo(types.UserRole(userRole.(string)))
	}
	if orgID, ok := token.Get(authjwt.OrgIdKey); ok {
		if parsedOrgID, err := uuid.Parse(orgID.(string)); err != nil {
			return nil, fmt.Errorf("JWT orgId is invalid: %w", err)
		} else {
			result.organizationID = util.PtrTo(parsedOrgID)
		}
	}
	if verified, ok := token.Get(authjwt.UserEmailVerifiedKey); ok {
		result.emailVerified = verified.(bool)
	}
	return &result, nil
}

func UserJWTAuthenticator() authn.Authenticator[jwt.Token, AuthInfo] {
	return authn.AuthenticatorFunc[jwt.Token, AuthInfo](
		func(ctx context.Context, token jwt.Token) (AuthInfo, error) {
			return FromUserJWT(token)
		},
	)
}

func FromAgentJWT(token jwt.Token) (*SimpleAgentAuthInfo, error) {
	var result SimpleAgentAuthInfo
	if parsedSub, err := uuid.Parse(token.Subject()); err != nil {
		return nil, fmt.Errorf("JWT subject is invalid: %w", err)
	} else {
		result.deploymentTargetID = parsedSub
	}
	if orgID, ok := token.Get(authjwt.OrgIdKey); ok {
		if parsedOrgID, err := uuid.Parse(orgID.(string)); err != nil {
			return nil, fmt.Errorf("JWT orgId is invalid: %w", err)
		} else {
			result.organizationID = parsedOrgID
		}
	}
	result.rawToken = token
	return &result, nil
}

func AgentJWTAuthenticator() authn.Authenticator[jwt.Token, AgentAuthInfo] {
	return authn.AuthenticatorFunc[jwt.Token, AgentAuthInfo](
		func(ctx context.Context, token jwt.Token) (AgentAuthInfo, error) {
			return FromAgentJWT(token)
		},
	)
}
