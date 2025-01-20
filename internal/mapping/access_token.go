package mapping

import (
	"github.com/glasskube/cloud/api"
	"github.com/glasskube/cloud/internal/types"
)

func AccessTokenToDTO(model types.AccessToken) api.AccessToken {
	return api.AccessToken{
		ID:         model.ID,
		CreatedAt:  model.CreatedAt,
		ExpiresAt:  model.ExpiresAt,
		LastUsedAt: model.LastUsedAt,
		Label:      model.Label,
	}
}
