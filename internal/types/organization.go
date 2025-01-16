package types

import "time"

type Organization struct {
	Base
	Name string `db:"name" json:"name"`
}

type OrganizationWithUserRole struct {
	Organization
	UserRole UserRole `db:"user_role"`
}

type OrganizationBranding struct {
	Base
	OrganizationID         string    `db:"organization_id" json:"-"`
	UpdatedAt              time.Time `db:"updated_at" json:"updatedAt"`
	UpdatedByUserAccountID *string   `db:"updated_by_user_account_id" json:"-"`
	Title                  *string   `db:"title" json:"title"`
	Description            *string   `db:"description" json:"description"`
	Logo                   []byte    `db:"logo" json:"logo"`
	LogoFileName           *string   `db:"logo_file_name" json:"logoFileName"`
	LogoContentType        *string   `db:"logo_content_type" json:"logoContentType"`
}
