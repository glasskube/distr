package api

import (
	"time"

	"github.com/google/uuid"
)

type DeploymentTargetNotes struct {
	UpdatedByUserAccountID *uuid.UUID `json:"updatedByUserAccountID,omitempty"`
	UpdatedAt              *time.Time `json:"updatedAt"`
	Notes                  string     `json:"notes"`
}

type DeploymentTargetNotesRequest struct {
	Notes string `json:"notes"`
}
