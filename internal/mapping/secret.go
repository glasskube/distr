package mapping

import (
	"github.com/distr-sh/distr/api"
	"github.com/distr-sh/distr/internal/types"
)

func SecretToAPI(s types.SecretWithUpdatedBy) *api.SecretWithoutValue {
	return &api.SecretWithoutValue{
		ID:                     s.ID,
		CreatedAt:              s.CreatedAt,
		UpdatedAt:              s.UpdatedAt,
		UpdatedBy:              s.UpdatedBy,
		CustomerOrganizationID: s.CustomerOrganizationID,
		Key:                    s.Key,
	}
}
