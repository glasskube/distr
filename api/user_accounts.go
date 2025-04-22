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

type UserAccountResponse struct {
	types.UserAccountWithUserRole
	ImageUrl string `json:"imageUrl"`
}

func AsUserAccount(u *types.UserAccountWithUserRole) UserAccountResponse {
	return UserAccountResponse{
		UserAccountWithUserRole: *u,
		ImageUrl:                imageUrl(*u),
	}
}

func imageUrl(u types.UserAccountWithUserRole) string {
	if u.ImageID == nil {
		return ""
	}
	return "/api/v1/files/" + u.ImageID.String()
}

func MapUserAccountsToResponse(userAccounts []types.UserAccountWithUserRole) []UserAccountResponse {
	result := make([]UserAccountResponse, len(userAccounts))
	for i, u := range userAccounts {
		result[i] = AsUserAccount(&u)
	}
	return result
}

type UpdateUserAccountRequest struct {
	Name     string  `json:"name"`
	Password *string `json:"password"`
}

type PatchUserAccountImageRequest struct {
	ImageID uuid.UUID `json:"imageId"`
}

func (r UpdateUserAccountRequest) Validate() error {
	if r.Password != nil {
		if err := validation.ValidatePassword(*r.Password); err != nil {
			return err
		}
	}
	return nil
}
