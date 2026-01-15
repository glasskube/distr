package api

import (
	"time"

	"github.com/distr-sh/distr/internal/types"
	"github.com/google/uuid"
)

type ApplicationResponse struct {
	types.Application
	ImageUrl string `json:"imageUrl"`
}

type PatchApplicationRequest struct {
	Name     *string                          `json:"name,omitempty"`
	Versions []PatchApplicationVersionRequest `json:"versions,omitempty"`
}

type PatchApplicationVersionRequest struct {
	ID         uuid.UUID  `json:"id"`
	ArchivedAt *time.Time `json:"archivedAt,omitempty"`
}
