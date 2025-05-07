package agentlogs

import (
	"strings"
	"time"

	"github.com/glasskube/distr/api"
	"github.com/google/uuid"
)

type LogRecordOption func(lr *api.LogRecord)

func WithSeverity(severity string) LogRecordOption {
	return func(lr *api.LogRecord) {
		lr.Severity = severity
	}
}

func WithTimestamp(ts time.Time) LogRecordOption {
	return func(lr *api.LogRecord) {
		lr.Timestamp = ts
	}
}

func NewRecord(deploymentID, deploymentRevisionID uuid.UUID, resource, severity, body string) api.LogRecord {
	record := api.LogRecord{
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
