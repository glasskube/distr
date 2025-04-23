package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/glasskube/distr/internal/apierrors"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/types"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	fileOutputExpr = "f.id, f.organization_id, f.created_at, f.content_type, f.data, f.file_name, f.file_size"
)

func CreateFile(ctx context.Context, organizationID uuid.UUID, file *types.File) error {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		"INSERT INTO File AS f (organization_id, content_type, data, file_name, file_size) "+
			"VALUES (@organization_id, @content_type, @data, @file_name, @file_size) "+
			"RETURNING "+fileOutputExpr,
		pgx.NamedArgs{
			"organization_id": organizationID,
			"content_type":    file.ContentType,
			"data":            file.Data,
			"file_name":       file.FileName,
			"file_size":       file.FileSize,
		},
	)
	if err != nil {
		return fmt.Errorf("could not query file: %w", err)
	} else if created, err := pgx.CollectExactlyOneRow[types.File](rows, pgx.RowToStructByName); err != nil {
		return fmt.Errorf("could not create file: %w", err)
	} else {
		*file = created
		return nil
	}
}

func GetFileWithID(ctx context.Context, id uuid.UUID) (*types.File, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		"SELECT "+fileOutputExpr+" FROM File f WHERE f.id = @id",
		pgx.NamedArgs{"id": id},
	)
	if err != nil {
		return nil, fmt.Errorf("could not query file: %w", err)
	} else if file, err := pgx.CollectExactlyOneRow[types.File](rows, pgx.RowToStructByName); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apierrors.ErrNotFound
		} else {
			return nil, fmt.Errorf("could not map file: %w", err)
		}
	} else {
		return &file, nil
	}
}

func DeleteFileWithID(ctx context.Context, id uuid.UUID) error {
	db := internalctx.GetDb(ctx)
	cmd, err := db.Exec(ctx, `DELETE FROM file WHERE id = @id`, pgx.NamedArgs{"id": id})
	if err != nil {
		if pgerr := (*pgconn.PgError)(nil); errors.As(err, &pgerr) && pgerr.Code == pgerrcode.ForeignKeyViolation {
			err = fmt.Errorf("%w: %w", apierrors.ErrConflict, err)
		}
	} else if cmd.RowsAffected() == 0 {
		err = apierrors.ErrNotFound
	}

	if err != nil {
		return fmt.Errorf("could not delete File: %w", err)
	}

	return nil
}
