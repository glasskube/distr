package types

import (
	"time"

	"github.com/google/uuid"
)

type DeploymentStatusNotificationConfiguration struct {
	ID                     uuid.UUID     `db:"id" json:"id"`
	CreatedAt              time.Time     `db:"created_at" json:"createdAt"`
	OrganizationID         uuid.UUID     `db:"organization_id" json:"organizationId"`
	CustomerOrganizationID *uuid.UUID    `db:"customer_organization_id" json:"customerOrganizationId"`
	Name                   string        `db:"name" json:"name"`
	Enabled                bool          `db:"enabled" json:"enabled"`
	DeploymentTargetIDs    []uuid.UUID   `json:"deploymentTargetIds"`
	UserAccountIDs         []uuid.UUID   `json:"userAccountIds"`
	UserAccounts           []UserAccount `db:"user_accounts" json:"-"`
}
