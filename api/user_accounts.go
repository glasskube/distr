package api

import (
	"github.com/glasskube/cloud/internal/types"
	"github.com/glasskube/cloud/internal/validation"
)

type CreateUserAccountRequest struct {
	Email           string         `json:"email"`
	Name            string         `json:"name"`
	ApplicationName string         `json:"applicationName"`
	UserRole        types.UserRole `json:"userRole"`
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
