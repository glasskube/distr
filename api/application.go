package api

import "github.com/glasskube/distr/internal/types"

type ApplicationResponse struct {
	types.Application
	ImageUrl string `json:"imageUrl"`
}

func AsApplication(a *types.Application) ApplicationResponse {
	return ApplicationResponse{
		Application: *a,
		ImageUrl:    WithImageUrl(a.ImageID),
	}
}

type ApplicationsResponse struct {
	types.Application
	ImageUrl string `json:"imageUrl"`
}

func AsApplications(a *types.Application) ApplicationsResponse {
	return ApplicationsResponse{
		Application: *a,
		ImageUrl:    WithImageUrl(a.ImageID),
	}
}

func MapApplicationsToResponse(applications []types.Application) []ApplicationsResponse {
	result := make([]ApplicationsResponse, len(applications))
	for i, a := range applications {
		result[i] = AsApplications(&a)
	}
	return result
}
