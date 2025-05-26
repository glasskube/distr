package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/glasskube/distr/api"
	"github.com/glasskube/distr/internal/apierrors"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const tutorialProgressOutExpr = " uat.tutorial, uat.created_at, uat.events, uat.completed_at "

func GetTutorialProgresses(ctx context.Context, userID, orgID uuid.UUID) (
	[]types.TutorialProgress,
	error,
) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx, `
		SELECT `+tutorialProgressOutExpr+`
		FROM UserAccount_TutorialProgress uat
		WHERE uat.useraccount_id = @userId AND uat.organization_id = @orgId`, pgx.NamedArgs{
		"userId": userID,
		"orgId":  orgID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query tutorial progresses: %w", err)
	}
	if res, err := pgx.CollectRows[types.TutorialProgress](rows, pgx.RowToStructByName); err != nil {
		return nil, err
	} else {
		return res, nil
	}
}

func GetTutorialProgress(ctx context.Context, userID, orgID uuid.UUID, tutorial types.Tutorial) (
	*types.TutorialProgress,
	error,
) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx, `
		SELECT `+tutorialProgressOutExpr+`
		FROM UserAccount_TutorialProgress uat
		WHERE uat.useraccount_id = @userId
			AND uat.tutorial = @tutorial
			AND uat.organization_id = @orgId`, pgx.NamedArgs{
		"userId":   userID,
		"tutorial": tutorial,
		"orgId":    orgID,
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
	userID, orgID uuid.UUID,
	tutorial types.Tutorial,
	progress *api.TutorialProgressRequest,
) (any, error) {
	db := internalctx.GetDb(ctx)
	progress.CreatedAt = time.Now()
	rows, err := db.Query(ctx, `
		INSERT INTO UserAccount_TutorialProgress as uat (useraccount_id, tutorial, events, completed_at, organization_id)
		VALUES (
			@userId,
			@tutorial,
			jsonb_build_array(@event::jsonb), CASE WHEN @markCompleted THEN current_timestamp ELSE NULL END,
		    @orgId
		)
		ON CONFLICT (useraccount_id, tutorial, organization_id) DO UPDATE
			SET events = uat.events::jsonb || @event::jsonb,
			    completed_at = CASE WHEN @markCompleted THEN current_timestamp ELSE uat.completed_at END
		RETURNING `+tutorialProgressOutExpr,
		pgx.NamedArgs{
			"userId":        userID,
			"tutorial":      tutorial,
			"event":         progress.TutorialProgressEvent,
			"markCompleted": progress.MarkCompleted,
			"orgId":         orgID,
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
