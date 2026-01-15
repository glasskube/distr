package mapping

import (
	"github.com/distr-sh/distr/api"
	"github.com/distr-sh/distr/internal/types"
	"github.com/distr-sh/distr/internal/util"
)

func UserAccountToAPI(u types.UserAccountWithUserRole) api.UserAccountResponse {
	return api.UserAccountResponse{
		UserAccountWithUserRole: u,
		ImageUrl:                util.CreateImageURL(u.ImageID),
	}
}
