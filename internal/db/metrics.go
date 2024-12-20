package db

import (
	"context"
	"time"

	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/types"
	"github.com/glasskube/cloud/internal/util"
	"github.com/jackc/pgx/v5"
)

func GetUptimeForDeployment(ctx context.Context, deploymentId string) ([]types.UptimeMetric, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		`
		SELECT hours.base_hour, statuses.created_at
		FROM (
			-- generate the last 24 dates where hours, minute, seconds is 0;
			-- having a list of all possible hours in the response makes it easier in the go algorithm to handle empty hours
			SELECT date_trunc('hour', hour_series) AS base_hour
			FROM generate_series(now() - INTERVAL '23 hour', now(), INTERVAL '1 hour') AS hour_series
		) AS hours LEFT JOIN (
			SELECT date_trunc('hour', created_at) AS created_at_hour, created_at
			FROM DeploymentStatus
			WHERE created_at > now() - INTERVAL '24 hour' AND deployment_id = @deploymentId
		) AS statuses ON hours.base_hour = statuses.created_at_hour
		ORDER BY hours.base_hour, statuses.created_at;`,
		pgx.NamedArgs{
			"deploymentId": deploymentId,
		})
	if err != nil {
		return nil, err
	}
	// TODO could be made configurable
	maxAllowedInterval := 10 * time.Second
	expectedPerHour := int(time.Hour / maxAllowedInterval)
	metricsIdx := 0
	metrics := make([]types.UptimeMetric, 24)
	var currentHourMetric types.UptimeMetric
	var previousRowBaseHour time.Time
	var previousRowCreatedAt *time.Time
	var currentBaseHour time.Time
	var currentCreatedAt *time.Time
	var diffToPrev *time.Duration

	processRow := func(isLast bool) {
		isFirst := previousRowBaseHour.IsZero()
		hourChanged := !previousRowBaseHour.Equal(currentBaseHour)
		if hourChanged || isLast {
			if !isFirst {
				// except on first and last iteration:
				// manually check duration between last createdAt of the previous hour to the beginning of the new hour
				if previousRowCreatedAt != nil && !isLast {
					lastDiff := currentBaseHour.Sub(*previousRowCreatedAt)
					missingStatuses := int(lastDiff / maxAllowedInterval)
					currentHourMetric.Unknown = currentHourMetric.Unknown + missingStatuses
					// if previousRowCreatedAt is null, the whole previous hour had no status (handled in previous iteration then)
				}
				// save away metrics of last hour
				metrics[metricsIdx] = currentHourMetric
				metricsIdx = metricsIdx + 1
			}
			currentHourMetric = types.UptimeMetric{
				Hour:    currentBaseHour,
				Total:   expectedPerHour,
				Unknown: 0,
			}
		}

		if currentCreatedAt == nil {
			// no status at all in that hour
			currentHourMetric.Unknown = currentHourMetric.Total
		} else {
			if hourChanged || previousRowCreatedAt == nil {
				// calculate diff to start of hour
				diffToPrev = util.PtrTo((*currentCreatedAt).Sub(currentBaseHour))
			} else {
				diffToPrev = util.PtrTo((*currentCreatedAt).Sub(*previousRowCreatedAt))
			}
			missingStatuses := int(*diffToPrev / maxAllowedInterval)
			currentHourMetric.Unknown = currentHourMetric.Unknown + missingStatuses
		}
		previousRowBaseHour = currentBaseHour
		previousRowCreatedAt = currentCreatedAt
	}

	for rows.Next() {
		if err := rows.Scan(&currentBaseHour, &currentCreatedAt); err != nil {
			return nil, err
		}
		processRow(false)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	} else {
		// for the last iteration, we need to "fake" a row at the current point of time
		currentCreatedAt = util.PtrTo(time.Now())
		currentBaseHour = (*currentCreatedAt).Truncate(1 * time.Hour)
		if previousRowCreatedAt != nil {
			diffToPrev = util.PtrTo(currentCreatedAt.Sub(*previousRowCreatedAt))
		} else {
			diffToPrev = util.PtrTo(currentCreatedAt.Sub(previousRowBaseHour))
		}
		processRow(true)
		// last row is the current hour until now, so it does not have the complete total/expected, but a reduced one
		restOfHour := time.Until(currentBaseHour.Add(1 * time.Hour))
		currentHourMetric.Total = currentHourMetric.Total - int(restOfHour/maxAllowedInterval)
		metrics[len(metrics)-1] = currentHourMetric
		return metrics, nil
	}
}
