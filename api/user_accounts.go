package api

import (
	"github.com/glasskube/distr/internal/types"
	"github.com/glasskube/distr/internal/validation"
	"github.com/google/uuid"
)

type CreateUserAccountRequest struct {
	Email           string         `json:"email"`
	Name            string         `json:"name"`
	ApplicationName string         `json:"applicationName"`
	UserRole        types.UserRole `json:"userRole"`
}

type CreateUserAccountResponse struct {
	ID        uuid.UUID `json:"id"`
	InviteURL string    `json:"inviteUrl"`
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
