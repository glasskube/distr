package db

import (
	"context"
	"errors"

	"github.com/glasskube/distr/internal/apierrors"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func GetLatestPullOfArtifactByUser(ctx context.Context, artifactId uuid.UUID, userId uuid.UUID) (string, error) {
	db := internalctx.GetDb(ctx)
	if rows, err := db.Query(ctx, `
		select av.name
		from ArtifactVersionPull avpl
		join ArtifactVersion av on av.id = avpl.artifact_version_id
		where av.artifact_id = @artifactId and avpl.useraccount_id = @userId
		and av.name not like '%:%'
		order by avpl.created_at desc
		limit 1;
	`, pgx.NamedArgs{
		"artifactId": artifactId,
		"userId":     userId,
	}); err != nil {
		return "", err
	} else if res, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByPos[struct{ Name string }]); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", apierrors.ErrNotFound
		}
		return "", err
	} else {
		return res.Name, nil
	}
}
