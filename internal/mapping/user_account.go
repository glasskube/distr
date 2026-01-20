package mapping

import (
	"github.com/distr-sh/distr/api"
	"github.com/distr-sh/distr/internal/types"
)

func UserAccountToAPI(u types.UserAccountWithUserRole) api.UserAccountResponse {
	return api.UserAccountResponse{
		UserAccountWithUserRole: u,
		ImageUrl:                CreateImageURL(u.ImageID),
	}
}
