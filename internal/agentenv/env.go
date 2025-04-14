package agentenv

import (
	"strconv"
	"time"

	"github.com/glasskube/distr/internal/envparse"
	"github.com/glasskube/distr/internal/envutil"
)

var (
	AgentVersionID         = envutil.GetEnv("DISTR_AGENT_VERSION_ID")
	Interval               = envutil.GetEnvParsedOrDefault("DISTR_INTERVAL", envparse.PositiveDuration, 5*time.Second)
	DistrRegistryHost      = envutil.GetEnv("DISTR_REGISTRY_HOST")
	DistrRegistryPlainHTTP = envutil.GetEnvParsedOrDefault("DISTR_REGISTRY_PLAIN_HTTP", strconv.ParseBool, false)
)
