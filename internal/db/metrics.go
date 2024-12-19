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
	// TODO index ???
	// TODO created_at not null everywhere

	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		`
		select hours.base_hour, statuses.created_at, statuses.diff_to_prev
		from (
			-- generate the last 24 dates where hours, minute, seconds is 0
			SELECT date_trunc('hour', hour_series) as base_hour
			FROM generate_series(now() - interval '23 hour', now(), interval '1 hour') AS hour_series
		) as hours left join (
			select deployment_id,
				   date_trunc('hour', created_at) as created_at_hour,
				   created_at,
				   created_at - lag(created_at) over (
					   partition by deployment_id order by created_at) as diff_to_prev
			from deploymentstatus
			where created_at > now() - interval '24 hour' and deployment_id = @deploymentId
			order by deployment_id, created_at
		) as statuses on hours.base_hour = statuses.created_at_hour;`,
		pgx.NamedArgs{
			"deploymentId": deploymentId,
		})
	if err != nil {
		return nil, err
	}
	// TODO could be made configurable
	maxAllowedInterval := 10 * time.Second
	expectedPerHour := int((1 * time.Hour).Seconds() / maxAllowedInterval.Seconds())
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
					missingStatuses := int(lastDiff.Seconds() / maxAllowedInterval.Seconds())
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
			if hourChanged {
				// manually calculate diff to start of hour
				diffToPrev = util.PtrTo((*currentCreatedAt).Sub(currentBaseHour))
			}
			missingStatuses := int(diffToPrev.Seconds() / maxAllowedInterval.Seconds())
			currentHourMetric.Unknown = currentHourMetric.Unknown + missingStatuses
		}
		previousRowBaseHour = currentBaseHour
		previousRowCreatedAt = currentCreatedAt
	}

	for rows.Next() {
		if err := rows.Scan(&currentBaseHour, &currentCreatedAt, &diffToPrev); err != nil {
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
		futureSeconds := (time.Until(currentBaseHour.Add(1 * time.Hour))).Seconds()
		currentHourMetric.Total = currentHourMetric.Total - int(futureSeconds/maxAllowedInterval.Seconds())
		metrics[len(metrics)-1] = currentHourMetric
		return metrics, nil
	}
}
