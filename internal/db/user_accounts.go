package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/glasskube/cloud/internal/apierrors"
	"github.com/glasskube/cloud/internal/auth"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/types"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	userAccountOutputExpr = "u.id, u.created_at, u.email, u.email_verified_at, u.password_hash, u.password_salt, u.name"
)

func CreateUserAccountWithOrganization(
	ctx context.Context,
	userAccount *types.UserAccount,
) (*types.Organization, error) {
	org := types.Organization{
		Name: userAccount.Email,
	}
	if err := CreateUserAccount(ctx, userAccount); err != nil {
		return nil, err
	} else if err := CreateOrganization(ctx, &org); err != nil {
		return nil, err
	} else if err := CreateUserAccountOrganizationAssignment(
		ctx,
		userAccount.ID,
		org.ID,
		types.UserRoleVendor,
	); err != nil {
		return nil, err
	} else {
		return &org, nil
	}
}

func CreateUserAccount(ctx context.Context, userAccount *types.UserAccount) error {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		"INSERT INTO UserAccount AS u (email, password_hash, password_salt, name) "+
			"VALUES (@email, @password_hash, @password_salt, @name) "+
			"RETURNING "+userAccountOutputExpr,
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

func UpateUserAccount(ctx context.Context, userAccount *types.UserAccount) error {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		`UPDATE UserAccount AS u
		SET email = @email,
			name = @name,
			password_hash = @password_hash,
			password_salt = @password_salt
		WHERE id = @id
		RETURNING `+userAccountOutputExpr,
		pgx.NamedArgs{
			"id":            userAccount.ID,
			"email":         userAccount.Email,
			"password_hash": userAccount.PasswordHash,
			"password_salt": userAccount.PasswordSalt,
			"name":          userAccount.Name,
		},
	)
	if err != nil {
		return fmt.Errorf("could not query users: %w", err)
	} else if created, err := pgx.CollectExactlyOneRow[types.UserAccount](rows, pgx.RowToStructByName); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apierrors.ErrNotFound
		} else if pgerr := (*pgconn.PgError)(nil); errors.As(err, &pgerr) && pgerr.Code == pgerrcode.UniqueViolation {
			return fmt.Errorf("can not update user with email %v: %w", userAccount.Email, apierrors.ErrAlreadyExists)
		}
		return fmt.Errorf("could not update user: %w", err)
	} else {
		*userAccount = created
		return nil
	}
}

func CreateUserAccountOrganizationAssignment(ctx context.Context, userId, orgId string, role types.UserRole) error {
	db := internalctx.GetDb(ctx)
	_, err := db.Exec(ctx,
		"INSERT INTO Organization_UserAccount (organization_id, user_account_id, user_role) VALUES (@orgId, @userId, @role)",
		pgx.NamedArgs{"userId": userId, "orgId": orgId, "role": role},
	)
	return err
}

func GetUserAccountsWithOrgID(ctx context.Context, orgId string) ([]types.UserAccountWithUserRole, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		"SELECT "+userAccountOutputExpr+`, j.user_role
		FROM UserAccount u
		INNER JOIN Organization_UserAccount j ON u.id = j.user_account_id
		WHERE j.organization_id = @orgId`,
		pgx.NamedArgs{"orgId": orgId},
	)
	if err != nil {
		return nil, fmt.Errorf("could not query users: %w", err)
	} else if result, err := pgx.CollectRows[types.UserAccountWithUserRole](rows, pgx.RowToStructByName); err != nil {
		return nil, fmt.Errorf("could not map users: %w", err)
	} else {
		return result, nil
	}
}

func GetUserAccountWithID(ctx context.Context, id string) (*types.UserAccount, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		"SELECT "+userAccountOutputExpr+" FROM UserAccount u WHERE u.id = @id",
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
		"SELECT "+userAccountOutputExpr+" FROM UserAccount u WHERE u.email = @email",
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

// GetCurrentUser retrieves the user account ID from the context auth token (subject claim) and returns the
// corresponding UserAccount
//
// TODO: this function should probably be moved to another module and maybe support some kind of result caching.
func GetCurrentUser(ctx context.Context) (*types.UserAccount, error) {
	if userId, err := auth.CurrentUserId(ctx); err != nil {
		return nil, err
	} else if user, err := GetUserAccountWithID(ctx, userId); err != nil {
		return nil, err
	} else {
		return user, nil
	}
}
