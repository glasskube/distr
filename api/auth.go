package api

import (
	"github.com/glasskube/distr/internal/validation"
)

type AuthLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthLoginResponse struct {
	Token string `json:"token"`
}

type AuthRegistrationRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *AuthRegistrationRequest) Validate() error {
	if r.Email == "" {
		return validation.NewValidationFailedError("email is empty")
	} else if err := validation.ValidatePassword(r.Password); err != nil {
		return err
	}
	return nil
}

type AuthResetPasswordRequest struct {
	Email string `json:"email"`
}

func (r *AuthResetPasswordRequest) Validate() error {
	if r.Email == "" {
		return validation.NewValidationFailedError("email is empty")
	}
	return nil
}
