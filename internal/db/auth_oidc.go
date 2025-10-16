package db

import (
	"context"
	"fmt"
	"time"

	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/oauth2"
)

func CreateOIDCState(ctx context.Context) (uuid.UUID, string, error) {
	pkceVerifier := oauth2.GenerateVerifier()
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx, "INSERT INTO OIDCState (pkce_code_verifier) VALUES (@pkce_code_verifier) RETURNING id", pgx.NamedArgs{"pkce_code_verifier": pkceVerifier})
	if err != nil {
		return uuid.Nil, "", err
	}
	id, err := pgx.CollectExactlyOneRow(rows, pgx.RowTo[uuid.UUID])
	if err != nil {
		return uuid.Nil, "", err
	}
	return id, pkceVerifier, nil
}

func DeleteOIDCState(ctx context.Context, id uuid.UUID) (string, time.Time, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx, "DELETE FROM OIDCState WHERE id = @id RETURNING created_at", pgx.NamedArgs{"id": id})
	if err != nil {
		return "", time.Time{}, err
	}
	type row struct {
		PKCECodeVerifier string    `db:"pkce_code_verifier"`
		CreatedAt        time.Time `db:"created_at"`
	}
	r, err := pgx.CollectExactlyOneRow(rows, pgx.RowTo[row])
	if err != nil {
		return "", time.Time{}, err
	}
	return r.PKCECodeVerifier, r.CreatedAt, nil
}

func CleanupOIDCStates(ctx context.Context) (int64, error) {
	db := internalctx.GetDb(ctx)
	cmd, err := db.Exec(
		ctx,
		`DELETE FROM OIDCState WHERE current_timestamp - created_at > @maxAge`,
		pgx.NamedArgs{"maxAge": 1 * time.Minute},
	)
	if err != nil {
		return 0, fmt.Errorf("error cleaning up OIDCState: %w", err)
	} else {
		return cmd.RowsAffected(), nil
	}
}
