package mapping

import (
	"github.com/distr-sh/distr/api"
	"github.com/distr-sh/distr/internal/types"
)

func DeploymentTargetLogRecordToAPI(record types.DeploymentTargetLogRecord) api.DeploymentTargetLogRecord {
	return api.DeploymentTargetLogRecord{
		ID:        record.ID,
		Timestamp: record.Timestamp,
		Severity:  record.Severity,
		Body:      record.Body,
	}
}
