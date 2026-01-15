package types

import (
	"time"

	"github.com/distr-sh/distr/internal/authkey"
	"github.com/google/uuid"
)

type AccessToken struct {
	ID             uuid.UUID   `db:"id"`
	CreatedAt      time.Time   `db:"created_at"`
	ExpiresAt      *time.Time  `db:"expires_at"`
	LastUsedAt     *time.Time  `db:"last_used_at"`
	Label          *string     `db:"label"`
	Key            authkey.Key `db:"key"`
	UserAccountID  uuid.UUID   `db:"user_account_id"`
	OrganizationID uuid.UUID   `db:"organization_id"`
}

func (tok AccessToken) HasExpired() bool {
	return tok.ExpiresAt == nil || tok.ExpiresAt.After(time.Now())
}

type AccessTokenWithUserAccount struct {
	AccessToken
	UserAccount            UserAccount `db:"user_account"`
	UserRole               UserRole    `db:"user_role"`
	CustomerOrganizationID *uuid.UUID  `db:"customer_organization_id"`
}
