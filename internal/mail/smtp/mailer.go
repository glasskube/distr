package smtp

import (
	"context"

	"github.com/glasskube/cloud/internal/mail"
	gomail "github.com/wneessen/go-mail"
)

type smtpMailer struct {
	client *gomail.Client
	config mail.MailerConfig
}

type Config struct {
	mail.MailerConfig
	Host, Username, Password string
	Port                     int
	TLSPolicy                gomail.TLSPolicy
}

var _ mail.Mailer = &smtpMailer{}

func New(config Config) (*smtpMailer, error) {
	client, err := gomail.NewClient(config.Host,
		gomail.WithPort(config.Port),
		gomail.WithSMTPAuth(gomail.SMTPAuthLogin),
		gomail.WithUsername(config.Username),
		gomail.WithPassword(config.Password),
		gomail.WithTLSPortPolicy(config.TLSPolicy),
	)

	if err != nil {
		return nil, err
	} else {
		return &smtpMailer{client: client, config: config.MailerConfig}, nil
	}
}

// Send implements mail.Mailer.
func (s *smtpMailer) Send(ctx context.Context, mail mail.Mail) error {
	message := gomail.NewMsg()
	message.Subject(mail.Subject)
	if err := message.From(s.config.FromAddress); err != nil {
		return err
	}
	for _, rcpt := range mail.To {
		if err := message.AddTo(rcpt); err != nil {
			return err
		}
	}
	if mail.HtmlBody != "" {
		message.SetBodyString(gomail.TypeTextHTML, mail.HtmlBody)
	}
	if mail.TextBody != "" {
		message.SetBodyString(gomail.TypeTextPlain, mail.TextBody)
	}
	return s.client.DialAndSendWithContext(ctx, message)
}
