package authinfo

import (
	"github.com/glasskube/distr/internal/types"
	"github.com/google/uuid"
)

type SimpleAuthInfo struct {
	userID         uuid.UUID
	userEmail      string
	organizationID *uuid.UUID
	emailVerified  bool
	userRole       *types.UserRole
	rawToken       any
}

// CurrentOrgID implements AuthInfo.
func (i *SimpleAuthInfo) CurrentOrgID() *uuid.UUID { return i.organizationID }

// CurrentUserEmailVerified implements AuthInfo.
func (i *SimpleAuthInfo) CurrentUserEmailVerified() bool { return i.emailVerified }

// CurrentUserID implements AuthInfo.
func (i *SimpleAuthInfo) CurrentUserID() uuid.UUID { return i.userID }

// CurrentUserEmail implements AuthInfo.
func (i *SimpleAuthInfo) CurrentUserEmail() string { return i.userEmail }

// CurrentUserRole implements AuthInfo.
func (i *SimpleAuthInfo) CurrentUserRole() *types.UserRole { return i.userRole }

// Token implements AuthInfo.
func (i *SimpleAuthInfo) Token() any { return i.rawToken }

func (i *SimpleAuthInfo) CurrentUser() *types.UserAccountWithUserRole {
	panic("SimpleAuthInfo does not contain the current user")
}

func (i *SimpleAuthInfo) CurrentOrg() *types.Organization {
	panic("SimpleAuthInfo does not contain the current org")
}
