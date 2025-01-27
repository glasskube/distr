package types

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/glasskube/distr/internal/util"
)

type Organization struct {
	Base
	Name string `db:"name" json:"name"`
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
	ID                     string    `db:"id" json:"id"`
	CreatedAt              time.Time `db:"created_at" json:"createdAt"`
	OrganizationID         string    `db:"organization_id" json:"-"`
	UpdatedAt              time.Time `db:"updated_at" json:"updatedAt"`
	UpdatedByUserAccountID *string   `db:"updated_by_user_account_id" json:"-"`
	Title                  *string   `db:"title" json:"title"`
	Description            *string   `db:"description" json:"description"`
	Logo                   []byte    `db:"logo" json:"logo"`
	LogoFileName           *string   `db:"logo_file_name" json:"logoFileName"`
	LogoContentType        *string   `db:"logo_content_type" json:"logoContentType"`
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
