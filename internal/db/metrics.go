package db

import (
	"context"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/types"
	"github.com/jackc/pgx/v5"
)

func GetUptimeForDeployment(ctx context.Context, deploymentId string) (*types.UptimeMetric, error) {
	// TODO index ???
	// TODO created_at not null everywhere

	db := internalctx.GetDb(ctx)
	row := db.QueryRow(ctx,
		`
		select
			extract(epoch from interval '24 hour') / extract(epoch from interval '10 second') as totalIntervals,
			sum(floor(extract(epoch from d.diff_to_prev) / extract(epoch from interval '10 second'))::int) +
				floor(extract(epoch from now() - max(d.created_at)) / extract(epoch from interval '10 second'))::int
				as timesStatusNotReceived
		from (
				 select deployment_id,
						created_at,
						created_at - lag(created_at, 1, now() - interval '24 hour') over (
							partition by deployment_id order by created_at) as diff_to_prev
				 from deploymentstatus
				 where created_at > now() - interval '24 hour'
				 order by deployment_id, created_at
			 ) as d
		where d.deployment_id = @deploymentId
		group by d.deployment_id;`,
		pgx.NamedArgs{
			"deploymentId": deploymentId,
			"psqlInterval": "10 second",
		})
	var total, unknown int
	if err := row.Scan(&total, &unknown); err != nil {
		return nil, err
	}
	return &types.UptimeMetric{
		Total:   total,
		Unknown: unknown,
	}, nil
}
