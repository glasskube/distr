package queryable

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Queryable interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	CopyFrom(
		ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource,
	) (int64, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}
