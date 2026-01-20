package types

import (
	"time"

	"github.com/google/uuid"
)

type DeploymentTargetLogRecord struct {
	ID                 uuid.UUID `db:"id"`
	CreatedAt          time.Time `db:"created_at"`
	DeploymentTargetID uuid.UUID `db:"deployment_target_id"`
	Timestamp          time.Time `db:"timestamp"`
	Severity           string    `db:"severity"`
	Body               string    `db:"body"`
}
