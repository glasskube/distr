package mapping

import (
	"github.com/glasskube/distr/api"
	"github.com/glasskube/distr/internal/types"
	"github.com/glasskube/distr/internal/util"
)

func UserAccountToAPI(u types.UserAccountWithUserRole) api.UserAccountResponse {
	return api.UserAccountResponse{
		UserAccountWithUserRole: u,
		ImageUrl:                util.CreateImageURL(u.ImageID),
	}
}
