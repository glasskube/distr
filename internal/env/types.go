package env

type MailerTypeString string
type TLSMode string

const (
	MailerTypeSMTP MailerTypeString = "smtp"
	MailerTypeSES  MailerTypeString = "ses"
)

type MailerConfig struct {
	Type        MailerTypeString
	FromAddress string
	SmtpConfig  *MailerSMTPConfig
}

type MailerSMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}
