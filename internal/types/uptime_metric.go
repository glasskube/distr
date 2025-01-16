package types

import (
	"time"
)

type UptimeMetric struct {
	Hour    time.Time `json:"hour"`
	Total   int       `json:"total"`
	Unknown int       `json:"unknown"`
}
