package api

type DeploymentTargetAccessTokenResponse struct {
	ConnectUrl   string `json:"connectUrl"`
	TargetId     string `json:"targetId"`
	TargetSecret string `json:"targetSecret"`
}
