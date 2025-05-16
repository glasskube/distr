package types

import (
	"time"

	"github.com/google/uuid"
)

type DeploymentLogRecord struct {
	ID                   uuid.UUID `db:"id"`
	CreatedAt            time.Time `db:"created_at"`
	DeploymentID         uuid.UUID `db:"deployment_id"`
	DeploymentRevisionID uuid.UUID `db:"deployment_revision_id"`
	Resource             string    `db:"resource"`
	Timestamp            time.Time `db:"timestamp"`
	Severity             string    `db:"severity"`
	Body                 string    `db:"body"`
}
