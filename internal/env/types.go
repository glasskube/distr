package env

import (
	"fmt"
	"net/mail"
)

type RegistrationMode string

const (
	RegistrationEnabled  RegistrationMode = "enabled"
	RegistrationHidden   RegistrationMode = "hidden"
	RegistrationDisabled RegistrationMode = "disabled"
)

func parseRegistrationMode(value string) (RegistrationMode, error) {
	switch value {
	case string(RegistrationEnabled), string(RegistrationHidden), string(RegistrationDisabled):
		return RegistrationMode(value), nil
	default:
		return "", fmt.Errorf("invalid value for environment variable REGISTRATION: %v", value)
	}
}

type MailerTypeString string

const (
	MailerTypeSMTP        MailerTypeString = "smtp"
	MailerTypeSES         MailerTypeString = "ses"
	MailerTypeUnspecified MailerTypeString = ""
)

func parseMailerType(value string) (MailerTypeString, error) {
	switch value {
	case string(MailerTypeSES), string(MailerTypeSMTP), string(MailerTypeUnspecified):
		return MailerTypeString(value), nil
	default:
		return "", fmt.Errorf("invalid value for environment variable MAILER_TYPE: %v", value)
	}
}

type MailerConfig struct {
	Type        MailerTypeString
	FromAddress mail.Address
	SmtpConfig  *MailerSMTPConfig
}

type MailerSMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

type S3Config struct {
	Bucket          string
	Region          string
	Endpoint        *string
	AccessKeyID     *string
	SecretAccessKey *string
	UsePathStyle    bool
	AllowRedirect   bool
}
