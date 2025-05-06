package agentlogs

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type LogCollector interface {
	Collect(resource string, timestamp time.Time, severity string, body string)
}

type resourceLogCollector struct {
	revisionID uuid.UUID
	recorder   LogRecorder
}

func NewCollector(revisionID uuid.UUID, recorder LogRecorder) LogCollector {
	return &resourceLogCollector{revisionID: revisionID, recorder: recorder}
}

func (c *resourceLogCollector) Collect(resource string, timestamp time.Time, severity string, body string) {
	c.recorder.Record(
		context.TODO(),
		c.revisionID,
		[]LogEntry{{Resource: resource, Timestamp: timestamp, Severity: severity, Body: body}},
	)
}
