package api

import "github.com/distr-sh/distr/internal/types"

type ContextResponse struct {
	User              UserAccountResponse              `json:"user"`
	Organization      OrganizationResponse             `json:"organization"`
	AvailableContexts []types.OrganizationWithUserRole `json:"availableContexts,omitempty"`
}
