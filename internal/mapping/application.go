package mapping

import (
	"github.com/distr-sh/distr/api"
	"github.com/distr-sh/distr/internal/types"
	"github.com/distr-sh/distr/internal/util"
)

func ApplicationToAPI(a types.Application) api.ApplicationResponse {
	return api.ApplicationResponse{
		Application: a,
		ImageUrl:    util.CreateImageURL(a.ImageID),
	}
}
