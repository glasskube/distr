package api

import (
	"time"

	"github.com/glasskube/distr/internal/types"
	"github.com/google/uuid"
)

type ApplicationResponse struct {
	types.Application
	ImageUrl string `json:"imageUrl"`
}

type PatchApplicationRequest struct {
	// ID is only used for OpenAPI spec generation
	ID       uuid.UUID                        `json:"-" path:"applicationId"`
	Name     *string                          `json:"name,omitempty"`
	Versions []PatchApplicationVersionRequest `json:"versions,omitempty"`
}

type PatchApplicationVersionRequest struct {
	ID         uuid.UUID  `json:"id"`
	ArchivedAt *time.Time `json:"archivedAt,omitempty"`
}
