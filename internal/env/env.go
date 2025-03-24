package env

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

var (
	databaseUrl                   string
	jwtSecret                     []byte
	host                          string
	artifactsHost                 string
	mailerConfig                  MailerConfig
	inviteTokenValidDuration      time.Duration
	resetTokenValidDuration       time.Duration
	agentTokenMaxValidDuration    time.Duration
	agentInterval                 time.Duration
	statusEntriesMaxAge           *time.Duration
	sentryDSN                     string
	sentryDebug                   bool
	enableQueryLogging            bool
	agentDockerConfig             []byte
	frontendSentryDSN             *string
	frontendPosthogToken          *string
	frontendPosthogAPIHost        *string
	frontendPosthogUIHost         *string
	userEmailVerificationRequired bool
	serverShutdownDelayDuration   *time.Duration
	registration                  RegistrationMode
	registryEnabled               bool
	registryS3Config              S3Config
)

func init() {
	if currentEnv, ok := os.LookupEnv("DISTR_ENV"); ok {
		fmt.Fprintf(os.Stderr, "environment=%v\n", currentEnv)
		if err := godotenv.Load(currentEnv); err != nil {
			fmt.Fprintf(os.Stderr, "environment not loaded: %v\n", err)
		}
	}

	databaseUrl = requireEnv("DATABASE_URL")
	jwtSecret = requireEnvParsed("JWT_SECRET", base64.StdEncoding.DecodeString)
	host = requireEnv("DISTR_HOST")
	artifactsHost = getEnvOrDefault("DISTR_ARTIFACTS_HOST", host)
	agentInterval = getEnvParsedOrDefault("AGENT_INTERVAL", getPositiveDuration, 5*time.Second)
	statusEntriesMaxAge = getEnvParsedOrNil("STATUS_ENTRIES_MAX_AGE", getPositiveDuration)
	enableQueryLogging = getEnvParsedOrDefault("ENABLE_QUERY_LOGGING", strconv.ParseBool, false)
	userEmailVerificationRequired =
		getEnvParsedOrDefault("USER_EMAIL_VERIFICATION_REQUIRED", strconv.ParseBool, true)
	serverShutdownDelayDuration = getEnvParsedOrNil("SERVER_SHUTDOWN_DELAY_DURATION", getPositiveDuration)
	registration = getEnvParsedOrDefault("REGISTRATION", parseRegistrationMode, RegistrationEnabled)
	inviteTokenValidDuration =
		getEnvParsedOrDefault("INVITE_TOKEN_VALID_DURATION", getPositiveDuration, 24*time.Hour)
	resetTokenValidDuration =
		getEnvParsedOrDefault("RESET_TOKEN_VALID_DURATION", getPositiveDuration, 1*time.Hour)
	agentTokenMaxValidDuration =
		getEnvParsedOrDefault("AGENT_TOKEN_MAX_VALID_DURATION", getPositiveDuration, 24*time.Hour)

	mailerConfig.Type = getEnvParsedOrDefault("MAILER_TYPE", parseMailerType, MailerTypeUnspecified)
	if mailerConfig.Type != MailerTypeUnspecified {
		mailerConfig.FromAddress = requireEnvParsed("MAILER_FROM_ADDRESS", parseMailAddress)
	}
	if mailerConfig.Type == MailerTypeSMTP {
		mailerConfig.SmtpConfig = &MailerSMTPConfig{
			Host:     getEnv("MAILER_SMTP_HOST"),
			Port:     requireEnvParsed("MAILER_SMTP_PORT", strconv.Atoi),
			Username: getEnv("MAILER_SMTP_USERNAME"),
			Password: getEnv("MAILER_SMTP_PASSWORD"),
		}
	}

	registryEnabled = getEnvParsedOrDefault("REGISTRY_ENABLED", strconv.ParseBool, false)
	if registryEnabled {
		registryS3Config.Bucket = requireEnv("REGISTRY_S3_BUCKET")
		registryS3Config.Region = requireEnv("REGISTRY_S3_REGION")
		registryS3Config.Endpoint = getEnvOrNil("REGISTRY_S3_ENDPOINT")
		registryS3Config.AccessKeyID = getEnvOrNil("REGISTRY_S3_ACCESS_KEY_ID")
		registryS3Config.SecretAccessKey = getEnvOrNil("REGISTRY_S3_SECRET_ACCESS_KEY")
		registryS3Config.UsePathStyle = getEnvParsedOrDefault("REGISTRY_S3_USE_PATH_STYLE", strconv.ParseBool, false)
		registryS3Config.AllowRedirect = getEnvParsedOrDefault("REGISTRY_S3_ALLOW_REDIRECT", strconv.ParseBool, true)
	}

	sentryDSN = getEnv("SENTRY_DSN")
	sentryDebug = getEnvParsedOrDefault("SENTRY_DEBUG", strconv.ParseBool, false)
	agentDockerConfig = getEnvParsedOrDefault("AGENT_DOCKER_CONFIG", asByteSlice, nil)
	frontendSentryDSN = getEnvOrNil("FRONTEND_SENTRY_DSN")
	frontendPosthogToken = getEnvOrNil("FRONTEND_POSTHOG_TOKEN")
	frontendPosthogAPIHost = getEnvOrNil("FRONTEND_POSTHOG_API_HOST")
	frontendPosthogUIHost = getEnvOrNil("FRONTEND_POSTHOG_UI_HOST")
}

func DatabaseUrl() string {
	return databaseUrl
}

func JWTSecret() []byte {
	return jwtSecret
}

func Host() string { return host }

func ArtifactsHost() string { return artifactsHost }

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

func AgentDockerConfig() []byte {
	return agentDockerConfig
}

func FrontendSentryDSN() *string {
	return frontendSentryDSN
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
