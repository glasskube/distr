package deploymenttargetlogs

import "github.com/distr-sh/distr/api"

type Exporter interface {
	ExportDeploymentTargetLogs(records ...api.DeploymentTargetLogRecord) error
}

type Syncer interface {
	Sync() error
}
