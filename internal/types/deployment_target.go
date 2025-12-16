package types

import (
	"github.com/glasskube/distr/internal/validation"
	"github.com/google/uuid"
)

type DeploymentTarget struct {
	Base
	Name                   string                  `db:"name" json:"name"`
	Type                   DeploymentType          `db:"type" json:"type"`
	AccessKeySalt          *[]byte                 `db:"access_key_salt" json:"-"`
	AccessKeyHash          *[]byte                 `db:"access_key_hash" json:"-"`
	CurrentStatus          *DeploymentTargetStatus `db:"current_status" json:"currentStatus,omitempty"`
	Namespace              *string                 `db:"namespace" json:"namespace,omitempty"`
	Scope                  *DeploymentTargetScope  `db:"scope" json:"scope,omitempty"`
	OrganizationID         uuid.UUID               `db:"organization_id" json:"-"`
	CreatedByUserAccountID uuid.UUID               `db:"created_by_user_account_id" json:"-"`
	CustomerOrganizationID *uuid.UUID              `db:"customer_organization_id" json:"customerOrganizationId,omitempty"`
	AgentVersionID         *uuid.UUID              `db:"agent_version_id" json:"-"`
	ReportedAgentVersionID *uuid.UUID              `db:"reported_agent_version_id" json:"reportedAgentVersionId,omitempty"`
	MetricsEnabled         bool                    `db:"metrics_enabled" json:"metricsEnabled"`
}

func (dt *DeploymentTarget) Validate() error {
	if dt.Type == DeploymentTypeKubernetes {
		if dt.Namespace == nil || *dt.Namespace == "" {
			return validation.NewValidationFailedError(
				"DeploymentTarget with type \"kubernetes\" must not have empty namespace",
			)
		}
		if dt.Scope == nil {
			return validation.NewValidationFailedError(
				"DeploymentTarget with type \"kubernetes\" must not have empty scope",
			)
		}
	}
	return nil
}

type DeploymentTargetWithCreatedBy struct {
	DeploymentTarget
	CreatedBy            *UserAccountWithUserRole       `db:"created_by" json:"createdBy"`
	CustomerOrganization *CustomerOrganization          `db:"customer_organization" json:"customerOrganization,omitempty"`
	Deployments          []DeploymentWithLatestRevision `db:"-" json:"deployments"`
	AgentVersion         AgentVersion                   `db:"agent_version" json:"agentVersion"`
}
