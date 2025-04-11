package types

import (
	"slices"
	"time"

	"github.com/glasskube/distr/internal/util"
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
	// Don't forget to update AsUserAccount when adding fields!
}

func (u *UserAccountWithUserRole) AsUserAccount() UserAccount {
	return UserAccount{
		ID:              u.ID,
		CreatedAt:       u.CreatedAt,
		Email:           u.Email,
		EmailVerifiedAt: util.PtrCopy(u.EmailVerifiedAt),
		PasswordHash:    slices.Clone(u.PasswordHash),
		PasswordSalt:    slices.Clone(u.PasswordSalt),
		Name:            u.Name,
		Password:        u.Password,
	}
}
