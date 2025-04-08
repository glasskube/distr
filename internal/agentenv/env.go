package agentenv

import (
	"time"

	"github.com/glasskube/distr/internal/envutil"
)

var (
	AgentVersionID    = envutil.GetEnv("DISTR_AGENT_VERSION_ID")
	DistrRegistryHost = envutil.GetEnv("DISTR_REGISTRY_HOST")
	Interval          = envutil.GetEnvParsedOrDefault("DISTR_INTERVAL", time.ParseDuration, 5*time.Second)
)
