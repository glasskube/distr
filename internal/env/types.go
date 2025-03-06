package env

import "net/mail"

type MailerTypeString string
type TLSMode string
type RegistrationMode string

const (
	MailerTypeSMTP        MailerTypeString = "smtp"
	MailerTypeSES         MailerTypeString = "ses"
	MailerTypeUnspecified MailerTypeString = ""

	RegistrationEnabled  RegistrationMode = "enabled"
	RegistrationHidden   RegistrationMode = "hidden"
	RegistrationDisabled RegistrationMode = "disabled"
)

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
