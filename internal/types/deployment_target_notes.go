package types

import (
	"time"

	"github.com/google/uuid"
)

type DeploymentTargetNotes struct {
	ID                     uuid.UUID  `db:"id"`
	DeploymentTargetID     uuid.UUID  `db:"deployment_target_id"`
	CustomerOrganizationID *uuid.UUID `db:"customer_organization_id"`
	UpdatedByUserAccountID *uuid.UUID `db:"updated_by_useraccount_id"`
	UpdatedAt              time.Time  `db:"updated_at"`
	Notes                  string     `db:"notes"`
}
