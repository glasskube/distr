package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/glasskube/distr/internal/apierrors"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func GetTutorialProgress(ctx context.Context, userID uuid.UUID, tutorial types.Tutorial) (*types.TutorialProgress, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx, `
		SELECT uat.tutorial, uat.created_at, uat.data
		FROM UserAccount_Tutorial uat
		WHERE uat.useraccount_id = @userId AND uat.tutorial = @tutorial`, pgx.NamedArgs{
		"userId":   userID,
		"tutorial": tutorial,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query tutorial progress: %w", err)
	}
	if res, err := pgx.CollectExactlyOneRow[types.TutorialProgress](rows, pgx.RowToStructByName); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apierrors.ErrNotFound
		}
		return nil, err
	} else {
		return &res, nil
	}
}

func SaveTutorialProgress(ctx context.Context, userID uuid.UUID, progress *types.TutorialProgress) error {
	return nil
}
