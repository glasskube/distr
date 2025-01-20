package types

import "time"

type DeploymentTargetStatus struct {
	// unfortunately Base nested type doesn't work when ApplicationVersion is a nested row in an SQL query
	ID                 string    `db:"id" json:"id"`
	CreatedAt          time.Time `db:"created_at" json:"createdAt"`
	Message            string    `db:"message" json:"message"`
	DeploymentTargetId string    `db:"deployment_target_id" json:"-"`
}
