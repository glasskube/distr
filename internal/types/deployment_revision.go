package types

import "github.com/google/uuid"

type DeploymentRevision struct {
	Base
	DeploymentID         uuid.UUID `db:"deployment_id" json:"deploymentId"`
	ApplicationVersionID uuid.UUID `db:"application_version_id" json:"applicationVersionId"`
	ValuesYaml           []byte    `db:"-" json:"valuesYaml,omitempty"`
	EnvFileData          []byte    `db:"-" json:"-"`
	ForceRestart         bool      `db:"force_restart" json:"forceRestart"`
	IgnoreRevisionSkew   bool      `db:"ignore_revision_skew" json:"ignoreRevisionSkew"`
}
