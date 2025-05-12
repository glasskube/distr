package types

import (
	"encoding/base64"
	"fmt"
	"slices"
	"time"

	"github.com/glasskube/distr/internal/util"
	"github.com/google/uuid"
)

type Organization struct {
	ID               uuid.UUID `db:"id" json:"id"`
	CreatedAt        time.Time `db:"created_at" json:"createdAt"`
	Name             string    `db:"name" json:"name"`
	Slug             *string   `db:"slug" json:"slug"`
	Features         []Feature `db:"features" json:"features"`
	AppDomain        *string   `db:"app_domain" json:"appDomain"`
	RegistryDomain   *string   `db:"registry_domain" json:"registryDomain"`
	EmailFromAddress *string   `db:"email_from_address" json:"emailFromAddress"`
}

func (org *Organization) HasFeature(feature Feature) bool {
	return slices.Contains(org.Features, feature)
}

type OrganizationWithUserRole struct {
	Organization
	UserRole UserRole `db:"user_role"`
}

type OrganizationWithBranding struct {
	Organization
	Branding *OrganizationBranding `db:"branding"`
}

type OrganizationBranding struct {
	ID                     uuid.UUID  `db:"id" json:"id"`
	CreatedAt              time.Time  `db:"created_at" json:"createdAt"`
	OrganizationID         uuid.UUID  `db:"organization_id" json:"-"`
	UpdatedAt              time.Time  `db:"updated_at" json:"updatedAt"`
	UpdatedByUserAccountID *uuid.UUID `db:"updated_by_user_account_id" json:"-"`
	Title                  *string    `db:"title" json:"title"`
	Description            *string    `db:"description" json:"description"`
	Logo                   []byte     `db:"logo" json:"logo"`
	LogoFileName           *string    `db:"logo_file_name" json:"logoFileName"`
	LogoContentType        *string    `db:"logo_content_type" json:"logoContentType"`
}

func (b *OrganizationBranding) LogoDataUrl() *string {
	if b.Logo != nil && b.LogoContentType != nil {
		return util.PtrTo(fmt.Sprintf(
			"data:%s;base64,%s",
			*b.LogoContentType,
			base64.StdEncoding.EncodeToString(b.Logo),
		))
	} else {
		return nil
	}
}
