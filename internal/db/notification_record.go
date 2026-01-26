package db

import (
	"context"

	"github.com/distr-sh/distr/internal/types"
	"github.com/google/uuid"
)

func SaveNotificationRecord(ctx context.Context, record *types.NotificationRecord) error {
	panic("not implemented")
}

func ExistsNotificationRecord(ctx context.Context, configID, previousID uuid.UUID) (bool, error) {
	panic("not implemented")
}
