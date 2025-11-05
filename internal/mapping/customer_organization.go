package mapping

import (
	"github.com/glasskube/distr/api"
	"github.com/glasskube/distr/internal/types"
	"github.com/glasskube/distr/internal/util"
)

func CustomerOrganizationToAPI(customerOrganization types.CustomerOrganization) api.CustomerOrganization {
	return api.CustomerOrganization{
		ID:        customerOrganization.ID,
		CreatedAt: customerOrganization.CreatedAt,
		Name:      customerOrganization.Name,
		ImageID:   customerOrganization.ImageID,
		ImageURL:  util.CreateImageURL(customerOrganization.ImageID),
	}
}

func CustomerOrganizationWithUserCountToAPI(
	customerOrganizationWithUserCount types.CustomerOrganizationWithUserCount,
) api.CustomerOrganizationWithUserCount {
	return api.CustomerOrganizationWithUserCount{
		CustomerOrganization: CustomerOrganizationToAPI(customerOrganizationWithUserCount.CustomerOrganization),
		UserCount:            customerOrganizationWithUserCount.UserCount,
	}
}
