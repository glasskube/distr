package types

import (
	"time"

	"github.com/glasskube/distr/internal/authkey"
)

type AccessToken struct {
	ID            string      `db:"id"`
	CreatedAt     time.Time   `db:"created_at"`
	ExpiresAt     *time.Time  `db:"expires_at"`
	LastUsedAt    *time.Time  `db:"last_used_at"`
	Label         *string     `db:"label"`
	Key           authkey.Key `db:"key"`
	UserAccountID string      `db:"user_account_id"`
}

func (tok AccessToken) HasExpired() bool {
	return tok.ExpiresAt == nil || tok.ExpiresAt.After(time.Now())
}

type AccessTokenWithUserAccount struct {
	AccessToken
	UserAccount UserAccount `db:"user_account"`
}
