package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/glasskube/distr/internal/authjwt"
	"github.com/glasskube/distr/internal/authn"
	"github.com/glasskube/distr/internal/authn/authinfo"
	"github.com/glasskube/distr/internal/authn/basic"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/security"
	"github.com/glasskube/distr/internal/types"
	"github.com/google/uuid"
)

func Authenticator() authn.Authenticator[basic.Auth, authinfo.AuthInfo] {
	fn := func(ctx context.Context, basic basic.Auth) (out authinfo.AuthInfo, err error) {
		if targetID, err1 := uuid.Parse(basic.Username); err1 != nil {
			err = fmt.Errorf("%w: %w", authn.ErrBadAuthentication, err1)
		} else if target, err1 := db.GetDeploymentTarget(ctx, targetID, nil); err1 != nil {
			err = err1
		} else if target.AccessKeyHash == nil || target.AccessKeySalt == nil {
			err = errors.New("access key or salt is nil")
		} else if err1 :=
			security.VerifyAccessKey(*target.AccessKeySalt, *target.AccessKeyHash, basic.Password); err1 != nil {
			err = fmt.Errorf("%w: %w", authn.ErrBadAuthentication, err1)
		} else if user, err1 := db.GetUserAccountByID(ctx, target.CreatedByUserAccountID); err1 != nil {
			err = err1
		} else if orgs, err1 := db.GetOrganizationsForUser(ctx, user.ID); err1 != nil {
			err = err1
		} else if len(orgs) != 1 {
			err = fmt.Errorf("user must have exactly one organization")
		} else if orgs[0].UserRole != types.UserRoleCustomer {
			err = fmt.Errorf("user must have role customer")
		} else if _, token, err1 := authjwt.GenerateDefaultToken(*user, orgs[0]); err1 != nil {
			err = err1
		} else {
			err = &tokenError{Token: token}
		}
		return
	}
	return authn.AuthenticatorFunc[basic.Auth, authinfo.AuthInfo](fn)
}

type tokenError struct {
	Token string `json:"token"`
}

var _ authn.WithResponseHeaders = &tokenError{}
var _ authn.WithResponseStatus = &tokenError{}
var _ authn.ResponseBodyWriter = &tokenError{}

// Error implements error.
func (t *tokenError) Error() string {
	return "NOT AN ERROR: agent successful token auth" // :^)
}

// ResponseHeaders implements authn.WithResponseHeaders.
func (t *tokenError) ResponseHeaders() http.Header {
	result := http.Header{}
	result.Add("Content-Type", "application/json")
	return result
}

// ResponseStatus implements authn.WithResponseStatus.
func (t *tokenError) ResponseStatus() int {
	return http.StatusOK
}

// WriteResponse implements authn.ResponseBodyWriter.
func (t *tokenError) WriteResponse(w io.Writer) {
	_ = json.NewEncoder(w).Encode(t)
}
