package env

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/mail"
	"os"
	"strconv"
	"time"

	"github.com/glasskube/cloud/internal/util"
	"github.com/joho/godotenv"
)

var (
	databaseUrl                string
	jwtSecret                  []byte
	host                       string
	mailerConfig               MailerConfig
	inviteTokenValidDuration   = 24 * time.Hour
	resetTokenValidDuration    = 1 * time.Hour
	agentTokenMaxValidDuration = 24 * time.Hour
	agentInterval              = 5 * time.Second
	statusEntriesMaxAge        *time.Duration
	sentryDSN                  string
	sentryDebug                bool
	enableQueryLogging         bool
	agentDockerConfig          []byte
	frontendSentryDSN          *string
	frontendPosthogToken       *string
	frontendPosthogAPIHost     *string
	frontendPosthogUIHost      *string
)

func init() {
	if currentEnv, ok := os.LookupEnv("GLASSKUBE_ENV"); ok {
		fmt.Fprintf(os.Stderr, "environment=%v\n", currentEnv)
		if err := godotenv.Load(currentEnv); err != nil {
			fmt.Fprintf(os.Stderr, "environment not loaded: %v\n", err)
		}
	}

	databaseUrl = os.Getenv("DATABASE_URL")

	if decoded, err := base64.StdEncoding.DecodeString(os.Getenv("JWT_SECRET")); err != nil {
		panic(fmt.Errorf("could not decode jwt secret: %w", err))
	} else {
		jwtSecret = decoded
	}
	host = os.Getenv("GLASSKUBE_HOST")
	if host == "" {
		panic(errors.New("can't start, GLASSKUBE_HOST not set"))
	}

	switch os.Getenv("MAILER_TYPE") {
	case "ses":
		mailerConfig.Type = MailerTypeSES
	case "smtp":
		mailerConfig.Type = MailerTypeSMTP
		port, err := strconv.Atoi(os.Getenv("MAILER_SMTP_PORT"))
		if err != nil {
			panic(fmt.Errorf("could not decode smtp port: %w", err))
		}
		mailerConfig.SmtpConfig = &MailerSMTPConfig{
			Host:     os.Getenv("MAILER_SMTP_HOST"),
			Port:     port,
			Username: os.Getenv("MAILER_SMTP_USERNAME"),
			Password: os.Getenv("MAILER_SMTP_PASSWORD"),
		}
	default:
		panic("invalid MAILER_TYPE")
	}
	mailerConfig.FromAddress = *util.Require(mail.ParseAddress(os.Getenv("MAILER_FROM_ADDRESS")))

	if d, ok := os.LookupEnv("INVITE_TOKEN_VALID_DURATION"); ok {
		inviteTokenValidDuration = requirePositiveDuration(d)
	}
	if d, ok := os.LookupEnv("RESET_TOKEN_VALID_DURATION"); ok {
		resetTokenValidDuration = requirePositiveDuration(d)
	}
	if d, ok := os.LookupEnv("AGENT_TOKEN_MAX_VALID_DURATION"); ok {
		agentTokenMaxValidDuration = requirePositiveDuration(d)
	}
	if d, ok := os.LookupEnv("AGENT_INTERVAL"); ok {
		agentInterval = requirePositiveDuration(d)
	}
	if d, ok := os.LookupEnv("STATUS_ENTRIES_MAX_AGE"); ok {
		statusEntriesMaxAge = util.PtrTo(requirePositiveDuration(d))
	}

	sentryDSN = os.Getenv("SENTRY_DSN")
	if value, ok := os.LookupEnv("SENTRY_DEBUG"); ok {
		sentryDebug = util.Require(strconv.ParseBool(value))
	}

	if value, ok := os.LookupEnv("ENABLE_QUERY_LOGGING"); ok {
		enableQueryLogging = util.Require(strconv.ParseBool(value))
	}

	if value, ok := os.LookupEnv("AGENT_DOCKER_CONFIG"); ok {
		agentDockerConfig = []byte(value)
	}

	if value, ok := os.LookupEnv("FRONTEND_SENTRY_DSN"); ok {
		frontendSentryDSN = &value
	}

	if value, ok := os.LookupEnv("FRONTEND_POSTHOG_TOKEN"); ok {
		frontendPosthogToken = &value
	}

	if value, ok := os.LookupEnv("FRONTEND_POSTHOG_API_HOST"); ok {
		frontendPosthogAPIHost = &value
	}

	if value, ok := os.LookupEnv("FRONTEND_POSTHOG_UI_HOST"); ok {
		frontendPosthogUIHost = &value
	}
}

func requirePositiveDuration(val string) time.Duration {
	d := util.Require(time.ParseDuration(val))
	if d.Nanoseconds() <= 0 {
		panic("duration must be positive")
	}
	return d
}

func DatabaseUrl() string {
	return databaseUrl
}

func JWTSecret() []byte {
	return jwtSecret
}

func Host() string { return host }

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
