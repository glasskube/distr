package types

import "time"

type DeploymentRevisionStatus struct {
	ID                   string               `db:"id" json:"id"`
	CreatedAt            time.Time            `db:"created_at" json:"createdAt"`
	DeploymentRevisionID string               `db:"deployment_revision_id" json:"deploymentRevisionId"`
	Type                 DeploymentStatusType `db:"type" json:"type"`
	Message              string               `db:"message" json:"message"`
}
