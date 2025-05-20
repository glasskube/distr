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

func AsApplication(a types.Application) ApplicationResponse {
	return ApplicationResponse{
		Application: a,
		ImageUrl:    WithImageUrl(a.ImageID),
	}
}

type ApplicationsResponse struct {
	types.Application
	ImageUrl string `json:"imageUrl"`
}

func AsApplications(a types.Application) ApplicationsResponse {
	return ApplicationsResponse{
		Application: a,
		ImageUrl:    WithImageUrl(a.ImageID),
	}
}

func MapApplicationsToResponse(applications []types.Application) []ApplicationsResponse {
	result := make([]ApplicationsResponse, len(applications))
	for i, a := range applications {
		result[i] = AsApplications(a)
	}
	return result
}

type PatchApplicationRequest struct {
	Name     *string                          `json:"name,omitempty"`
	Versions []PatchApplicationVersionRequest `json:"versions,omitempty"`
}

type PatchApplicationVersionRequest struct {
	ID         uuid.UUID  `json:"id"`
	ArchivedAt *time.Time `json:"archivedAt,omitempty"`
}
