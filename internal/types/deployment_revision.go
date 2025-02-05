package types

type DeploymentRevision struct {
	Base
	DeploymentID         string `db:"deployment_id" json:"deploymentId"`
	ApplicationVersionId string `db:"application_version_id" json:"applicationVersionId"`
	ValuesYaml           []byte `db:"-" json:"valuesYaml,omitempty"`
	EnvFileData          []byte `db:"-" json:"-"`
}
