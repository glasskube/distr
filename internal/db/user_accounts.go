package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/glasskube/cloud/internal/apierrors"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/types"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func CreateUserAccountWithOrganization(
	ctx context.Context,
	userAccount *types.UserAccount,
) (*types.Organization, error) {
	if err := CreateUserAccount(ctx, userAccount); err != nil {
		return nil, err
	} else {
		org := types.Organization{
			Name: userAccount.Email,
		}
		if err := CreateOrganization(ctx, &org); err != nil {
			return nil, err
		} else {
			return &org, nil
		}
	}
}

func CreateUserAccount(ctx context.Context, userAccount *types.UserAccount) error {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		"INSERT INTO UserAccount (email, password_hash, password_salt, name) "+
			"VALUES (@email, @password_hash, @password_salt, @name) "+
			"RETURNING id, created_at, email, password_hash, password_salt, name",
		pgx.NamedArgs{
			"email":         userAccount.Email,
			"password_hash": userAccount.PasswordHash,
			"password_salt": userAccount.PasswordSalt,
			"name":          userAccount.Name,
		},
	)
	if err != nil {
		return fmt.Errorf("could not query users: %w", err)
	} else if created, err := pgx.CollectExactlyOneRow[types.UserAccount](rows, pgx.RowToStructByName); err != nil {
		if pgerr := (*pgconn.PgError)(nil); errors.As(err, &pgerr) && pgerr.Code == pgerrcode.UniqueViolation {
			return fmt.Errorf("user account with email %v can not be created: %w", userAccount.Email, apierrors.ErrAlreadyExists)
		}
		return fmt.Errorf("could not create user: %w", err)
	} else {
		*userAccount = created
		return nil
	}
}

func GetUserAccountWithID(ctx context.Context, id string) (*types.UserAccount, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		"SELECT id, created_at, email, password_hash, password_salt, name FROM UserAccount WHERE id = @id",
		pgx.NamedArgs{"id": id},
	)
	if err != nil {
		return nil, fmt.Errorf("could not query users: %w", err)
	} else if userAccount, err := pgx.CollectExactlyOneRow[types.UserAccount](rows, pgx.RowToStructByName); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apierrors.ErrNotFound
		} else {
			return nil, fmt.Errorf("could not map user: %w", err)
		}
	} else {
		return &userAccount, nil
	}
}

func GetUserAccountWithEmail(ctx context.Context, email string) (*types.UserAccount, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		"SELECT id, created_at, email, password_hash, password_salt, name FROM UserAccount WHERE email = @email",
		pgx.NamedArgs{"email": email},
	)
	if err != nil {
		return nil, fmt.Errorf("could not query users: %w", err)
	} else if userAccount, err := pgx.CollectExactlyOneRow[types.UserAccount](rows, pgx.RowToStructByName); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apierrors.ErrNotFound
		} else {
			return nil, fmt.Errorf("could not map user: %w", err)
		}
	} else {
		return &userAccount, nil
	}
}
