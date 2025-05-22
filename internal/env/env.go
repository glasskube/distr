package env

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/glasskube/distr/internal/envparse"
	"github.com/glasskube/distr/internal/envutil"
	"github.com/joho/godotenv"
)

var (
	databaseUrl                         string
	databaseMaxConns                    *int
	jwtSecret                           []byte
	host                                string
	registryHost                        string
	mailerConfig                        MailerConfig
	inviteTokenValidDuration            time.Duration
	resetTokenValidDuration             time.Duration
	agentTokenMaxValidDuration          time.Duration
	agentInterval                       time.Duration
	statusEntriesMaxAge                 *time.Duration
	metricsEntriesMaxAge                *time.Duration
	logRecordEntriesMaxCount            *int
	sentryDSN                           string
	sentryDebug                         bool
	otelExporterSentryEnabled           bool
	otelExporterOtlpEnabled             bool
	enableQueryLogging                  bool
	agentDockerConfig                   []byte
	frontendSentryDSN                   *string
	frontendSentryTraceSampleRate       *float64
	frontendPosthogToken                *string
	frontendPosthogAPIHost              *string
	frontendPosthogUIHost               *string
	userEmailVerificationRequired       bool
	serverShutdownDelayDuration         *time.Duration
	registration                        RegistrationMode
	registryEnabled                     bool
	registryS3Config                    S3Config
	artifactTagsDefaultLimitPerOrg      int
	cleanupDeploymentRevisionStatusCron *string
	cleanupDeploymentTargetStatusCron   *string
)

func Initialize() {
	if currentEnv, ok := os.LookupEnv("DISTR_ENV"); ok {
		fmt.Fprintf(os.Stderr, "environment=%v\n", currentEnv)
		if err := godotenv.Load(currentEnv); err != nil {
			fmt.Fprintf(os.Stderr, "environment not loaded: %v\n", err)
		}
	}

	databaseUrl = envutil.RequireEnv("DATABASE_URL")
	databaseMaxConns = envutil.GetEnvParsedOrNil("DATABASE_MAX_CONNS", strconv.Atoi)
	jwtSecret = envutil.RequireEnvParsed("JWT_SECRET", base64.StdEncoding.DecodeString)
	host = envutil.RequireEnv("DISTR_HOST")
	agentInterval = envutil.GetEnvParsedOrDefault("AGENT_INTERVAL", envparse.PositiveDuration, 5*time.Second)
	statusEntriesMaxAge = envutil.GetEnvParsedOrNil("STATUS_ENTRIES_MAX_AGE", envparse.PositiveDuration)
	metricsEntriesMaxAge = envutil.GetEnvParsedOrNil("METRICS_ENTRIES_MAX_AGE", envparse.PositiveDuration)
	logRecordEntriesMaxCount = envutil.GetEnvParsedOrNil("LOG_RECORD_ENTRIES_MAX_COUNT", envparse.NonNegativeNumber)
	enableQueryLogging = envutil.GetEnvParsedOrDefault("ENABLE_QUERY_LOGGING", strconv.ParseBool, false)
	userEmailVerificationRequired =
		envutil.GetEnvParsedOrDefault("USER_EMAIL_VERIFICATION_REQUIRED", strconv.ParseBool, true)
	serverShutdownDelayDuration = envutil.GetEnvParsedOrNil("SERVER_SHUTDOWN_DELAY_DURATION", envparse.PositiveDuration)
	registration = envutil.GetEnvParsedOrDefault("REGISTRATION", parseRegistrationMode, RegistrationEnabled)
	inviteTokenValidDuration =
		envutil.GetEnvParsedOrDefault("INVITE_TOKEN_VALID_DURATION", envparse.PositiveDuration, 24*time.Hour)
	resetTokenValidDuration =
		envutil.GetEnvParsedOrDefault("RESET_TOKEN_VALID_DURATION", envparse.PositiveDuration, 1*time.Hour)
	agentTokenMaxValidDuration =
		envutil.GetEnvParsedOrDefault("AGENT_TOKEN_MAX_VALID_DURATION", envparse.PositiveDuration, 24*time.Hour)

	mailerConfig.Type = envutil.GetEnvParsedOrDefault("MAILER_TYPE", parseMailerType, MailerTypeUnspecified)
	if mailerConfig.Type != MailerTypeUnspecified {
		mailerConfig.FromAddress = envutil.RequireEnvParsed("MAILER_FROM_ADDRESS", envparse.MailAddress)
	}
	if mailerConfig.Type == MailerTypeSMTP {
		mailerConfig.SmtpConfig = &MailerSMTPConfig{
			Host:     envutil.GetEnv("MAILER_SMTP_HOST"),
			Port:     envutil.RequireEnvParsed("MAILER_SMTP_PORT", strconv.Atoi),
			Username: envutil.GetEnv("MAILER_SMTP_USERNAME"),
			Password: envutil.GetEnv("MAILER_SMTP_PASSWORD"),
		}
	}

	registryEnabled = envutil.GetEnvParsedOrDefault("REGISTRY_ENABLED", strconv.ParseBool, false)
	if registryEnabled {
		registryHost =
			envutil.GetEnvOrDefault("REGISTRY_HOST", host, envutil.GetEnvOpts{DeprecatedAlias: "DISTR_ARTIFACTS_HOST"})
		registryS3Config.Bucket = envutil.RequireEnv("REGISTRY_S3_BUCKET")
		registryS3Config.Region = envutil.RequireEnv("REGISTRY_S3_REGION")
		registryS3Config.Endpoint = envutil.GetEnvOrNil("REGISTRY_S3_ENDPOINT")
		registryS3Config.AccessKeyID = envutil.GetEnvOrNil("REGISTRY_S3_ACCESS_KEY_ID")
		registryS3Config.SecretAccessKey = envutil.GetEnvOrNil("REGISTRY_S3_SECRET_ACCESS_KEY")
		registryS3Config.UsePathStyle = envutil.GetEnvParsedOrDefault("REGISTRY_S3_USE_PATH_STYLE", strconv.ParseBool, false)
		registryS3Config.AllowRedirect = envutil.GetEnvParsedOrDefault("REGISTRY_S3_ALLOW_REDIRECT", strconv.ParseBool, true)
	}
	artifactTagsDefaultLimitPerOrg =
		envutil.GetEnvParsedOrDefault("ARTIFACT_TAGS_DEFAULT_LIMIT_PER_ORG", envparse.NonNegativeNumber, 0)

	sentryDSN = envutil.GetEnv("SENTRY_DSN")
	sentryDebug = envutil.GetEnvParsedOrDefault("SENTRY_DEBUG", strconv.ParseBool, false)
	otelExporterSentryEnabled = envutil.GetEnvParsedOrDefault("OTEL_EXPORTER_SENTRY_ENABLED", strconv.ParseBool, false)
	otelExporterOtlpEnabled = envutil.GetEnvParsedOrDefault("OTEL_EXPORTER_OTLP_ENABLED", strconv.ParseBool, false)
	agentDockerConfig = envutil.GetEnvParsedOrDefault("AGENT_DOCKER_CONFIG", envparse.ByteSlice, nil)
	frontendSentryDSN = envutil.GetEnvOrNil("FRONTEND_SENTRY_DSN")
	frontendSentryTraceSampleRate = envutil.GetEnvParsedOrNil("FRONTEND_SENTRY_TRACE_SAMPLE_RATE", envparse.Float)
	frontendPosthogToken = envutil.GetEnvOrNil("FRONTEND_POSTHOG_TOKEN")
	frontendPosthogAPIHost = envutil.GetEnvOrNil("FRONTEND_POSTHOG_API_HOST")
	frontendPosthogUIHost = envutil.GetEnvOrNil("FRONTEND_POSTHOG_UI_HOST")

	cleanupDeploymentRevisionStatusCron = envutil.GetEnvOrNil("CLEANUP_DEPLOYMENT_REVISION_STATUS_CRON")
	cleanupDeploymentTargetStatusCron = envutil.GetEnvOrNil("CLEANUP_DEPLOYMENT_TARGET_STATUS_CRON")
}

func DatabaseUrl() string {
	return databaseUrl
}

// DatabaseMaxConns allows to override the MaxConns parameter of the pgx pool config.
//
// Note that it should also be possible to set this value via the connection string
// (like this: postgresql://...?pool_max_conns=10), but it doesn't work for some reason.
func DatabaseMaxConns() *int {
	return databaseMaxConns
}

func JWTSecret() []byte {
	return jwtSecret
}

func Host() string { return host }

func RegistryHost() string { return registryHost }

func GetMailerConfig() MailerConfig {
	return mailerConfig
}

func InviteTokenValidDuration() time.Duration {
	return inviteTokenValidDuration
}

func ResetTokenValidDuration() time.Duration {
	return resetTokenValidDuration
}

func AgentTokenMaxValidDuration() time.Duration {
	return agentTokenMaxValidDuration
}

func AgentInterval() time.Duration {
	return agentInterval
}

func SentryDSN() string {
	return sentryDSN
}

func SentryDebug() bool {
	return sentryDebug
}

func EnableQueryLogging() bool {
	return enableQueryLogging
}

func StatusEntriesMaxAge() *time.Duration {
	return statusEntriesMaxAge
}

func MetricsEntriesMaxAge() *time.Duration {
	return metricsEntriesMaxAge
}

func LogRecordEntriesMaxCount() *int {
	return logRecordEntriesMaxCount
}

func AgentDockerConfig() []byte {
	return agentDockerConfig
}

func FrontendSentryDSN() *string {
	return frontendSentryDSN
}

func FrontendSentryTraceSampleRate() *float64 {
	return frontendSentryTraceSampleRate
}

func FrontendPosthogToken() *string {
	return frontendPosthogToken
}

func FrontendPosthogAPIHost() *string {
	return frontendPosthogAPIHost
}
func FrontendPosthogUIHost() *string {
	return frontendPosthogUIHost
}

func UserEmailVerificationRequired() bool {
	return userEmailVerificationRequired
}

func ServerShutdownDelayDuration() *time.Duration {
	return serverShutdownDelayDuration
}

func Registration() RegistrationMode {
	return registration
}

func RegistryEnabled() bool {
	return registryEnabled
}

func RegistryS3Config() S3Config {
	return registryS3Config
}

func ArtifactTagsDefaultLimitPerOrg() int {
	return artifactTagsDefaultLimitPerOrg
}

func OtelExporterSentryEnabled() bool {
	return otelExporterSentryEnabled
}

func OtelExporterOtlpEnabled() bool {
	return otelExporterOtlpEnabled
}

func CleanupDeploymenRevisionStatusCron() *string {
	return cleanupDeploymentRevisionStatusCron
}

func CleanupDeploymenTargetStatusCron() *string {
	return cleanupDeploymentTargetStatusCron
}
