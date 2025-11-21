package mapping

import (
	"github.com/glasskube/distr/api"
	"github.com/glasskube/distr/internal/types"
	"github.com/glasskube/distr/internal/util"
)

func ApplicationToAPI(a types.Application) api.ApplicationResponse {
	return api.ApplicationResponse{
		Application: a,
		ImageUrl:    util.CreateImageURL(a.ImageID),
	}
}
