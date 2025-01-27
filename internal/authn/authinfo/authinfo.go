package authinfo

import "github.com/glasskube/distr/internal/types"

type AuthInfo interface {
	CurrentUserID() string
	CurrentUserEmail() string
	CurrentUserRole() *types.UserRole
	CurrentOrgID() *string
	CurrentUserEmailVerified() bool
}
