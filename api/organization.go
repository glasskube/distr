package api

import (
	"github.com/glasskube/distr/internal/types"
)

type CreateUpdateOrganizationRequest struct {
	Name                   string  `json:"name"`
	Slug                   *string `json:"slug"`
	PreConnectScript       *string `json:"preConnectScript"`
	PostConnectScript      *string `json:"postConnectScript"`
	ConnectScriptIsSudo    bool    `json:"connectScriptIsSudo"`
	ArtifactVersionMutable bool    `json:"artifactVersionMutable"`
}

type OrganizationResponse struct {
	types.Organization
	SubscriptionLimits SubscriptionLimits `json:"subscriptionLimits"`
}
