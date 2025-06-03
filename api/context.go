package api

import "github.com/glasskube/distr/internal/types"

type ContextResponse struct {
	User              UserAccountResponse              `json:"user"`
	Organization      types.Organization               `json:"organization"`
	AvailableContexts []types.OrganizationWithUserRole `json:"availableContexts,omitempty"`
}
