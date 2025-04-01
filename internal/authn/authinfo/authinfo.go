package authinfo

import (
	"github.com/glasskube/distr/internal/types"
	"github.com/google/uuid"
)

type AuthInfo interface {
	CurrentUserID() uuid.UUID
	CurrentUserEmail() string
	CurrentUserRole() *types.UserRole
	CurrentOrgID() *uuid.UUID
	CurrentUserEmailVerified() bool
	Token() any
	// TODO dont know yet if this belongs here:
	CurrentUser() *types.UserAccount
	CurrentOrg() *types.Organization
}
