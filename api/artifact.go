package api

import "github.com/distr-sh/distr/internal/types"

type ArtifactResponse struct {
	types.ArtifactWithTaggedVersion
	ImageUrl string `json:"imageUrl,omitempty"`
}

type ArtifactsResponse struct {
	types.ArtifactWithDownloads
	ImageUrl string `json:"imageUrl,omitempty"`
}
