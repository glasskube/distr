package types

import (
	"time"

	"github.com/google/uuid"
)

type UserAccount struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	CreatedAt       time.Time  `db:"created_at" json:"createdAt"`
	Email           string     `db:"email" json:"email"`
	EmailVerifiedAt *time.Time `db:"email_verified_at" json:"-"`
	PasswordHash    []byte     `db:"password_hash" json:"-"`
	PasswordSalt    []byte     `db:"password_salt" json:"-"`
	Name            string     `db:"name" json:"name,omitempty"`
	Password        string     `db:"-" json:"-"`
}

type UserAccountWithIcon struct {
	UserAccount
	Icon            []byte  `db:"icon" json:"icon"`
	IconFileName    *string `db:"icon_file_name" json:"iconFileName"`
	IconContentType *string `db:"icon_content_type" json:"iconContentType"`
}

type UserAccountWithUserRole struct {
	// copy+pasted from UserAccount because pgx does not like embedded strucs
	ID              uuid.UUID  `db:"id" json:"id"`
	CreatedAt       time.Time  `db:"created_at" json:"createdAt"`
	Email           string     `db:"email" json:"email"`
	EmailVerifiedAt *time.Time `db:"email_verified_at" json:"-"`
	PasswordHash    []byte     `db:"password_hash" json:"-"`
	PasswordSalt    []byte     `db:"password_salt" json:"-"`
	Name            string     `db:"name" json:"name,omitempty"`
	UserRole        UserRole   `db:"user_role" json:"userRole"` // not copy+pasted
	Password        string     `db:"-" json:"-"`
}
