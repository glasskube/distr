package db

import (
	"iter"

	"github.com/jackc/pgx/v5"
)

type SeqE[T any] = iter.Seq2[T, error]

func CollectSeq[T any](rows pgx.Rows, fn pgx.RowToFunc[T]) SeqE[T] {
	return func(yield func(T, error) bool) {
		defer rows.Close()
		for rows.Next() {
			if !yield(fn(rows)) {
				return
			}
		}
		if err := rows.Err(); err != nil {
			var zero T
			_ = yield(zero, err)
		}
	}
}
