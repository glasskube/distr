package types

import (
	"errors"
)

type DeploymentTarget struct {
	Base
	Name                   string                  `db:"name" json:"name"`
	Type                   DeploymentType          `db:"type" json:"type"`
	Geolocation            *Geolocation            `db:"geolocation" json:"geolocation,omitempty"`
	AccessKeySalt          *[]byte                 `db:"access_key_salt" json:"-"`
	AccessKeyHash          *[]byte                 `db:"access_key_hash" json:"-"`
	CurrentStatus          *DeploymentTargetStatus `db:"current_status" json:"currentStatus,omitempty"`
	Namespace              *string                 `db:"namespace" json:"namespace"`
	OrganizationID         string                  `db:"organization_id" json:"-"`
	CreatedByUserAccountID string                  `db:"created_by_user_account_id" json:"-"`
	AgentVersionID         string                  `db:"agent_version_id" json:"-"`
	ReportedAgentVersionID *string                 `db:"reported_agent_version_id" json:"reportedAgentVersionId,omitempty"`
}

func (dt *DeploymentTarget) Validate() error {
	if dt.Type == DepolymentTypeKubernetes {
		if dt.Namespace == nil || *dt.Namespace == "" {
			return errors.New("DeploymentTarget with type \"kubernetes\" must not have empty namespace")
		}
	}
	return nil
}

type DeploymentTargetWithCreatedBy struct {
	DeploymentTarget
	CreatedBy    *UserAccountWithUserRole      `db:"created_by" json:"createdBy"`
	Deployment   *DeploymentWithLatestRevision `db:"-" json:"deployment,omitempty"`
	AgentVersion AgentVersion                  `db:"agent_version" json:"agentVersion"`
}
