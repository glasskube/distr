package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/glasskube/distr/internal/apierrors"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func GetTutorialProgress(ctx context.Context, userID uuid.UUID, tutorial types.Tutorial) (
	*types.TutorialProgress,
	error,
) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx, `
		SELECT uat.tutorial, uat.created_at, uat.events, uat.completed_at
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

func SaveTutorialProgress(
	ctx context.Context,
	userID uuid.UUID,
	tutorial types.Tutorial,
	progress *types.TutorialProgressRequest,
) (any, error) {
	db := internalctx.GetDb(ctx)
	progress.CreatedAt = time.Now()
	rows, err := db.Query(ctx, `
		INSERT INTO UserAccount_Tutorial as uat (useraccount_id, tutorial, events, completed_at)
		VALUES (
			@userId,
			@tutorial,
			jsonb_build_array(@event::jsonb), CASE WHEN @markCompleted THEN current_timestamp ELSE NULL END
		)
		ON CONFLICT (useraccount_id, tutorial) DO UPDATE
			SET events = uat.events::jsonb || @event::jsonb,
			    completed_at = CASE WHEN @markCompleted THEN current_timestamp ELSE uat.completed_at END
		RETURNING uat.tutorial, uat.created_at, uat.events, uat.completed_at`,
		pgx.NamedArgs{
			"userId":        userID,
			"tutorial":      tutorial,
			"event":         progress.TutorialProgressEvent,
			"markCompleted": progress.MarkCompleted,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	} else if res, err := pgx.CollectExactlyOneRow[types.TutorialProgress](rows, pgx.RowToStructByName); err != nil {
		return nil, err
	} else {
		return res, err
	}
}
