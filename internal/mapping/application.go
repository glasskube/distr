package mapping

import (
	"github.com/distr-sh/distr/api"
	"github.com/distr-sh/distr/internal/types"
)

func ApplicationToAPI(a types.Application) api.ApplicationResponse {
	return api.ApplicationResponse{
		Application: a,
		ImageUrl:    CreateImageURL(a.ImageID),
	}
}
