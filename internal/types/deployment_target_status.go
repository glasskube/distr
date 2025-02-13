package types

import (
	"time"

	"github.com/google/uuid"
)

type DeploymentTargetStatus struct {
	// unfortunately Base nested type doesn't work when ApplicationVersion is a nested row in an SQL query
	ID                 uuid.UUID `db:"id" json:"id"`
	CreatedAt          time.Time `db:"created_at" json:"createdAt"`
	Message            string    `db:"message" json:"message"`
	DeploymentTargetID uuid.UUID `db:"deployment_target_id" json:"-"`
}
