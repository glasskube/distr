package db

import (
	"context"
	"fmt"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/types"
	"github.com/jackc/pgx/v5"
)

func GetUptimeForDeployment(ctx context.Context, deploymentId string) ([]types.UptimeMetric, error) {
	// TODO index ???
	// TODO created_at not null everywhere

	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		`
		select
			date_trunc('hour', d.created_at) as hour,
			extract(epoch from interval '1 hour') / extract(epoch from interval '10 second') as total,
			sum(floor(extract(epoch from d.diff_to_prev) / extract(epoch from interval '10 second'))::int) +
			(CASE WHEN (date_trunc('hour', now()) = date_trunc('hour', max(d.created_at))) THEN (
				floor(extract(epoch from now() - max(d.created_at)) / extract(epoch from interval '10 second'))::int
				) ELSE 0 END) as unknown
		from (
				 select deployment_id,
						created_at,
						created_at - lag(created_at) over (
							partition by deployment_id order by created_at) as diff_to_prev
				 from deploymentstatus
				 where created_at > now() - interval '24 hour'
				 order by deployment_id, created_at
			 ) as d
		where d.deployment_id = @deploymentId
		group by hour, d.deployment_id
		order by 1;`,
		pgx.NamedArgs{
			"deploymentId": deploymentId,
		})
	if err != nil {
		return nil, err
	}
	result, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.UptimeMetric])
	if err != nil {
		return nil, fmt.Errorf("failed to get uptime metrics: %w", err)
	} else {
		return result, nil
	}
}
