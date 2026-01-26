package mapping

import (
	"github.com/distr-sh/distr/api"
	"github.com/distr-sh/distr/internal/types"
)

func ArtifactVersionPullToAPI(pull types.ArtifactVersionPull) api.ArtifactVersionPullResponse {
	response := api.ArtifactVersionPullResponse{
		CreatedAt:       pull.CreatedAt,
		RemoteAddress:   pull.RemoteAddress,
		Artifact:        pull.Artifact,
		ArtifactVersion: pull.ArtifactVersion,
	}

	if pull.UserAccount != nil {
		if pull.UserAccount.Name != "" {
			response.UserAccountName = &pull.UserAccount.Name
		}
		response.UserAccountEmail = &pull.UserAccount.Email
	}

	if pull.CustomerOrganization != nil {
		response.CustomerOrganizationName = &pull.CustomerOrganization.Name
	}

	return response
}
