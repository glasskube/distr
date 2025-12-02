package api

import "github.com/glasskube/distr/internal/types"

type ArtifactResponse struct {
	types.ArtifactWithTaggedVersion
	ImageUrl string `json:"imageUrl"`
}

type ArtifactsResponse struct {
	types.ArtifactWithDownloads
	ImageUrl string `json:"imageUrl"`
}
