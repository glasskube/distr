package agentlogs

import (
	"strings"
	"time"

	"github.com/distr-sh/distr/api"
	"github.com/google/uuid"
)

type LogRecordOption func(lr *api.DeploymentLogRecord)

func WithSeverity(severity string) LogRecordOption {
	return func(lr *api.DeploymentLogRecord) {
		lr.Severity = severity
	}
}

func WithTimestamp(ts time.Time) LogRecordOption {
	return func(lr *api.DeploymentLogRecord) {
		lr.Timestamp = ts
	}
}

func NewRecord(deploymentID, deploymentRevisionID uuid.UUID, resource, severity, body string) api.DeploymentLogRecord {
	record := api.DeploymentLogRecord{
		DeploymentID:         deploymentID,
		DeploymentRevisionID: deploymentRevisionID,
		Resource:             resource,
		Timestamp:            time.Now(),
		Severity:             severity,
		Body:                 body,
	}
	messageParts := strings.SplitN(body, " ", 2)
	if len(messageParts) > 1 {
		if ts, err := time.Parse(time.RFC3339Nano, messageParts[0]); err == nil {
			record.Timestamp = ts
			record.Body = strings.TrimSpace(messageParts[1])
		}
	}
	return record
}
