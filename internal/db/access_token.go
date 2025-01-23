package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/glasskube/cloud/internal/apierrors"
	"github.com/glasskube/cloud/internal/authkey"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/types"
	"github.com/jackc/pgx/v5"
)

const accessTokenOutputExpr = `
	tok.id, tok.created_at, tok.expires_at, tok.last_used_at, tok.label, tok.key, tok.user_account_id
`
const accessTokenWithUserAccountOutputExpr = accessTokenOutputExpr + `,
	CASE WHEN u.id IS NOT NULL THEN (` + userAccountOutputExpr + `) END
		AS user_account
`

func CreateAccessToken(ctx context.Context, token *types.AccessToken) error {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		fmt.Sprintf(
			`INSERT INTO AccessToken AS tok (label, expires_at, key, user_account_id)
			VALUES (@label, @expiresAt, @key, @userAccountId)
			RETURNING %v`,
			accessTokenOutputExpr),
		pgx.NamedArgs{
			"label":         token.Label,
			"expiresAt":     token.ExpiresAt,
			"key":           token.Key[:],
			"userAccountId": token.UserAccountID,
		},
	)
	if err != nil {
		return fmt.Errorf("could not create access token: %w", err)
	}
	if res, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.AccessToken]); err != nil {
		return fmt.Errorf("could not create access token: %w", err)
	} else {
		*token = res
		return nil
	}
}

func DeleteAccessToken(ctx context.Context, id string, userID string) error {
	db := internalctx.GetDb(ctx)
	if _, err := db.Exec(
		ctx,
		"DELETE FROM AccessToken WHERE id = @id AND user_account_id = @userId",
		pgx.NamedArgs{"id": id, "userId": userID},
	); err != nil {
		return fmt.Errorf("could not delete token: %w", err)
	}
	return nil
}

func GetAccessTokensByUserAccountID(ctx context.Context, id string) ([]types.AccessToken, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		fmt.Sprintf(`SELECT %v FROM AccessToken tok WHERE tok.user_account_id = @id`, accessTokenOutputExpr),
		pgx.NamedArgs{"id": id},
	)
	if err != nil {
		return nil, fmt.Errorf("error querying access tokens: %w", err)
	}
	if result, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.AccessToken]); err != nil {
		return nil, fmt.Errorf("could not get tokens: %w", err)
	} else {
		return result, nil
	}
}

func GetAccessTokenByKeyUpdatingLastUsed(
	ctx context.Context,
	key authkey.Key,
) (*types.AccessTokenWithUserAccount, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		fmt.Sprintf(
			`WITH updated AS (
				UPDATE AccessToken
				SET last_used_at = now()
				WHERE key = @key AND (expires_at IS NULL OR expires_at > now())
				RETURNING *
			)
			SELECT %v FROM updated tok
			LEFT JOIN UserAccount u ON tok.user_account_id = u.id
			`,
			accessTokenWithUserAccountOutputExpr,
		),
		pgx.NamedArgs{"key": key[:]},
	)
	if err != nil {
		return nil, fmt.Errorf("error querying access token: %w", err)
	}
	if result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.AccessTokenWithUserAccount]); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = apierrors.ErrNotFound
		}
		return nil, fmt.Errorf("could not get token: %w", err)
	} else {
		return &result, nil
	}
}
