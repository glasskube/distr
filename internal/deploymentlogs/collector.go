package deploymentlogs

import (
	"github.com/distr-sh/distr/api"
	"github.com/google/uuid"
)

type DeploymentIDer interface {
	GetDeploymentID() uuid.UUID
	GetDeploymentRevisionID() uuid.UUID
}

type Collector interface {
	For(DeploymentIDer) DeploymentCollector
	LogRecords() []api.DeploymentLogRecord
}

type DeploymentCollector interface {
	AppendMessage(resource, severity, message string)
}

type collector struct {
	logRecords []api.DeploymentLogRecord
}

func NewCollector() Collector {
	return &collector{}
}

// For implements Collector.
func (c *collector) For(d DeploymentIDer) DeploymentCollector {
	return &deploymentCollector{collector: c, DeploymentIDer: d}
}

// LogRecords implements Collector.
func (c *collector) LogRecords() []api.DeploymentLogRecord {
	return c.logRecords
}

func (c *collector) appendRecord(record api.DeploymentLogRecord) {
	c.logRecords = append(c.logRecords, record)
}

type deploymentCollector struct {
	*collector
	DeploymentIDer
}

// AppendMessage implements DeploymentCollector.
func (d *deploymentCollector) AppendMessage(resource string, severity string, message string) {
	record := NewRecord(d.GetDeploymentID(), d.GetDeploymentRevisionID(), resource, severity, message)
	if record.Body != "" {
		d.appendRecord(record)
	}
}
