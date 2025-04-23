package api

import "github.com/glasskube/distr/internal/types"

type ArtifactResponse struct {
	types.ArtifactWithTaggedVersion
	ImageUrl string `json:"imageUrl"`
}

func AsArtifact(a *types.ArtifactWithTaggedVersion) ArtifactResponse {
	return ArtifactResponse{
		ArtifactWithTaggedVersion: *a,
		ImageUrl:                  WithImageUrl(a.ImageID),
	}
}

type ArtifactsResponse struct {
	types.ArtifactWithDownloads
	ImageUrl string `json:"imageUrl"`
}

func AsArtifacts(a *types.ArtifactWithDownloads) ArtifactsResponse {
	return ArtifactsResponse{
		ArtifactWithDownloads: *a,
		ImageUrl:              WithImageUrl(a.ImageID),
	}
}

func MapArtifactsToResponse(artifacts []types.ArtifactWithDownloads) []ArtifactsResponse {
	result := make([]ArtifactsResponse, len(artifacts))
	for i, a := range artifacts {
		result[i] = AsArtifacts(&a)
	}
	return result
}
