package env

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/glasskube/distr/internal/envutil"
	"github.com/joho/godotenv"
)

var (
	databaseUrl                    string
	jwtSecret                      []byte
	host                           string
	registryHost                   string
	mailerConfig                   MailerConfig
	inviteTokenValidDuration       time.Duration
	resetTokenValidDuration        time.Duration
	agentTokenMaxValidDuration     time.Duration
	agentInterval                  time.Duration
	statusEntriesMaxAge            *time.Duration
	sentryDSN                      string
	sentryDebug                    bool
	enableQueryLogging             bool
	agentDockerConfig              []byte
	frontendSentryDSN              *string
	frontendPosthogToken           *string
	frontendPosthogAPIHost         *string
	frontendPosthogUIHost          *string
	userEmailVerificationRequired  bool
	serverShutdownDelayDuration    *time.Duration
	registration                   RegistrationMode
	registryEnabled                bool
	registryS3Config               S3Config
	artifactTagsDefaultLimitPerOrg int
)

func init() {
	if currentEnv, ok := os.LookupEnv("DISTR_ENV"); ok {
		fmt.Fprintf(os.Stderr, "environment=%v\n", currentEnv)
		if err := godotenv.Load(currentEnv); err != nil {
			fmt.Fprintf(os.Stderr, "environment not loaded: %v\n", err)
		}
	}

	databaseUrl = envutil.RequireEnv("DATABASE_URL")
	jwtSecret = envutil.RequireEnvParsed("JWT_SECRET", base64.StdEncoding.DecodeString)
	host = envutil.RequireEnv("DISTR_HOST")
	agentInterval = envutil.GetEnvParsedOrDefault("AGENT_INTERVAL", getPositiveDuration, 5*time.Second)
	statusEntriesMaxAge = envutil.GetEnvParsedOrNil("STATUS_ENTRIES_MAX_AGE", getPositiveDuration)
	enableQueryLogging = envutil.GetEnvParsedOrDefault("ENABLE_QUERY_LOGGING", strconv.ParseBool, false)
	userEmailVerificationRequired =
		envutil.GetEnvParsedOrDefault("USER_EMAIL_VERIFICATION_REQUIRED", strconv.ParseBool, true)
	serverShutdownDelayDuration = envutil.GetEnvParsedOrNil("SERVER_SHUTDOWN_DELAY_DURATION", getPositiveDuration)
	registration = envutil.GetEnvParsedOrDefault("REGISTRATION", parseRegistrationMode, RegistrationEnabled)
	inviteTokenValidDuration =
		envutil.GetEnvParsedOrDefault("INVITE_TOKEN_VALID_DURATION", getPositiveDuration, 24*time.Hour)
	resetTokenValidDuration =
		envutil.GetEnvParsedOrDefault("RESET_TOKEN_VALID_DURATION", getPositiveDuration, 1*time.Hour)
	agentTokenMaxValidDuration =
		envutil.GetEnvParsedOrDefault("AGENT_TOKEN_MAX_VALID_DURATION", getPositiveDuration, 24*time.Hour)

	mailerConfig.Type = envutil.GetEnvParsedOrDefault("MAILER_TYPE", parseMailerType, MailerTypeUnspecified)
	if mailerConfig.Type != MailerTypeUnspecified {
		mailerConfig.FromAddress = envutil.RequireEnvParsed("MAILER_FROM_ADDRESS", parseMailAddress)
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
		envutil.GetEnvParsedOrDefault("ARTIFACT_TAGS_DEFAULT_LIMIT_PER_ORG", getNonNegativeNumber, 0)

	sentryDSN = envutil.GetEnv("SENTRY_DSN")
	sentryDebug = envutil.GetEnvParsedOrDefault("SENTRY_DEBUG", strconv.ParseBool, false)
	agentDockerConfig = envutil.GetEnvParsedOrDefault("AGENT_DOCKER_CONFIG", asByteSlice, nil)
	frontendSentryDSN = envutil.GetEnvOrNil("FRONTEND_SENTRY_DSN")
	frontendPosthogToken = envutil.GetEnvOrNil("FRONTEND_POSTHOG_TOKEN")
	frontendPosthogAPIHost = envutil.GetEnvOrNil("FRONTEND_POSTHOG_API_HOST")
	frontendPosthogUIHost = envutil.GetEnvOrNil("FRONTEND_POSTHOG_UI_HOST")
}

func DatabaseUrl() string {
	return databaseUrl
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

func ArtifactTagsDefaultLimitPerOrg() int {
	return artifactTagsDefaultLimitPerOrg
}
