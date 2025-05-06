package agentlogs

import "time"

type LogEntry struct {
	Resource  string    `json:"resource"`
	Timestamp time.Time `json:"timestamp"`
	Severity  string    `json:"severity"`
	Body      string    `json:"body"`
}
