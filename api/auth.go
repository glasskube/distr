package api

import "errors"

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
		return errors.New("email is empty")
	} else if len(r.Password) < 8 {
		return errors.New("password is too short")
	} else {
		return nil
	}
}

type DeploymentTargetAccessTokenResponse struct {
	ConnectUrl   string `json:"connectUrl"`
	TargetId     string `json:"targetId"`
	TargetSecret string `json:"targetSecret"`
}
