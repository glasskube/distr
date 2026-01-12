package api

import (
	"github.com/glasskube/distr/internal/types"
	"github.com/glasskube/distr/internal/validation"
	"github.com/google/uuid"
)

type CreateUserAccountRequest struct {
	Email                  string         `json:"email"`
	Name                   string         `json:"name"`
	UserRole               types.UserRole `json:"userRole"`
	CustomerOrganizationID *uuid.UUID     `json:"customerOrganizationId,omitempty"`
}

type CreateUserAccountResponse struct {
	User      types.UserAccountWithUserRole `json:"user"`
	InviteURL string                        `json:"inviteUrl"`
}

type UserAccountResponse struct {
	types.UserAccountWithUserRole
	ImageUrl string `json:"imageUrl"`
}

type UpdateUserAccountRequest struct {
	Name     string  `json:"name"`
	Password *string `json:"password"`
}

func (r UpdateUserAccountRequest) Validate() error {
	if r.Password != nil {
		if err := validation.ValidatePassword(*r.Password); err != nil {
			return err
		}
	}
	return nil
}

type PatchUserAccountRequest struct {
	UserRole *types.UserRole `json:"userRole"`
}
