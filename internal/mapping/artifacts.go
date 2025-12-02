package mapping

import (
	"github.com/glasskube/distr/api"
	"github.com/glasskube/distr/internal/types"
	"github.com/glasskube/distr/internal/util"
)

func ArtifactToAPI(a types.ArtifactWithTaggedVersion) api.ArtifactResponse {
	return api.ArtifactResponse{
		ArtifactWithTaggedVersion: a,
		ImageUrl:                  util.CreateImageURL(a.ImageID),
	}
}

func ArtifactsWithDownloadsToAPI(a types.ArtifactWithDownloads) api.ArtifactsResponse {
	return api.ArtifactsResponse{
		ArtifactWithDownloads: a,
		ImageUrl:              util.CreateImageURL(a.ImageID),
	}
}
