package api

import (
	"time"

	"github.com/google/uuid"
)

type LogRecord struct {
	DeploymentID         uuid.UUID `json:"deploymentId"`
	DeploymentRevisionID uuid.UUID `json:"deploymentRevisionId"`
	Resource             string    `json:"resource"`
	Timestamp            time.Time `json:"timestamp"`
	Severity             string    `json:"severity"`
	Body                 string    `json:"body"`
}
