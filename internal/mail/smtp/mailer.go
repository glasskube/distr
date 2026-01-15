package smtp

import (
	"context"

	"github.com/distr-sh/distr/internal/mail"
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
	ImplicitTLS              bool
}

var _ mail.Mailer = &smtpMailer{}

func New(config Config) (*smtpMailer, error) {
	options := []gomail.Option{
		gomail.WithTLSPortPolicy(config.TLSPolicy),
	}
	if config.Port != 0 {
		options = append(options, gomail.WithPort(config.Port))
	}
	if config.Username != "" {
		options = append(options,
			gomail.WithSMTPAuth(gomail.SMTPAuthLogin),
			gomail.WithUsername(config.Username),
			gomail.WithPassword(config.Password),
			gomail.WithoutNoop(),
		)
	}
	if config.ImplicitTLS {
		options = append(options, gomail.WithSSL())
	}
	client, err := gomail.NewClient(config.Host, options...)

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
	if err := message.To(mail.To...); err != nil {
		return err
	}
	if err := message.Bcc(mail.Bcc...); err != nil {
		return err
	}
	if err := message.From(s.config.GetActualFromAddress(ctx, mail)); err != nil {
		return err
	}
	if mail.ReplyTo != "" {
		if err := message.ReplyTo(mail.ReplyTo); err != nil {
			return err
		}
	}
	if mail.HtmlBodyFunc != nil {
		if body, err := mail.HtmlBodyFunc(); err != nil {
			return err
		} else {
			message.SetBodyString(gomail.TypeTextHTML, body)
		}
	}
	if mail.TextBodyFunc != nil {
		if body, err := mail.TextBodyFunc(); err != nil {
			return err
		} else {
			message.SetBodyString(gomail.TypeTextPlain, body)
		}
	}
	return s.client.DialAndSendWithContext(ctx, message)
}
