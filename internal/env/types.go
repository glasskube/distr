package env

import "net/mail"

type MailerTypeString string
type TLSMode string

const (
	MailerTypeSMTP MailerTypeString = "smtp"
	MailerTypeSES  MailerTypeString = "ses"
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
