package authinfo

import "github.com/glasskube/distr/internal/types"

type SimpleAuthInfo struct {
	userID         string
	userEmail      string
	organizationID *string
	emailVerified  bool
	userRole       *types.UserRole
	rawToken       any
}

// CurrentOrgID implements AuthInfo.
func (i *SimpleAuthInfo) CurrentOrgID() *string { return i.organizationID }

// CurrentUserEmailVerified implements AuthInfo.
func (i *SimpleAuthInfo) CurrentUserEmailVerified() bool { return i.emailVerified }

// CurrentUserID implements AuthInfo.
func (i *SimpleAuthInfo) CurrentUserID() string { return i.userID }

// CurrentUserEmail implements AuthInfo.
func (i *SimpleAuthInfo) CurrentUserEmail() string { return i.userEmail }

// CurrentUserRole implements AuthInfo.
func (i *SimpleAuthInfo) CurrentUserRole() *types.UserRole { return i.userRole }

// Token implements AuthInfo.
func (i *SimpleAuthInfo) Token() any { return i.rawToken }
