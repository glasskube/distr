package svc

import (
	"context"
	"errors"

	"github.com/distr-sh/distr/internal/auth"
	"github.com/distr-sh/distr/internal/env"
	"github.com/distr-sh/distr/internal/mail"
	"github.com/distr-sh/distr/internal/mail/noop"
	"github.com/distr-sh/distr/internal/mail/ses"
	"github.com/distr-sh/distr/internal/mail/smtp"
	gomail "github.com/wneessen/go-mail"
)

func (r *Registry) GetMailer() mail.Mailer {
	return r.mailer
}

func createMailer(ctx context.Context) (mail.Mailer, error) {
	config := env.GetMailerConfig()
	authOrgOverrideFromAddress := func(ctx context.Context, mail mail.Mail) string {
		if auth, err := auth.Authentication.Get(ctx); err == nil {
			if org := auth.CurrentOrg(); org != nil && org.EmailFromAddress != nil {
				return *org.EmailFromAddress
			}
		}
		return ""
	}
	switch config.Type {
	case env.MailerTypeSMTP:
		smtpConfig := smtp.Config{
			MailerConfig: mail.MailerConfig{
				FromAddressSrc: []mail.FromAddressSrcFn{
					mail.MailOverrideFromAddress(),
					authOrgOverrideFromAddress,
					mail.StaticFromAddress(config.FromAddress.String()),
				},
			},
			Host:        config.SmtpConfig.Host,
			Port:        config.SmtpConfig.Port,
			Username:    config.SmtpConfig.Username,
			Password:    config.SmtpConfig.Password,
			ImplicitTLS: config.SmtpConfig.ImplicitTLS,
			TLSPolicy:   gomail.TLSOpportunistic,
		}
		return smtp.New(smtpConfig)
	case env.MailerTypeSES:
		sesConfig := ses.Config{
			MailerConfig: mail.MailerConfig{
				FromAddressSrc: []mail.FromAddressSrcFn{
					mail.MailOverrideFromAddress(),
					authOrgOverrideFromAddress,
					mail.StaticFromAddress(config.FromAddress.String()),
				},
			},
		}
		return ses.NewFromContext(ctx, sesConfig)
	case env.MailerTypeUnspecified:
		return noop.New(), nil
	default:
		return nil, errors.New("invalid mailer type")
	}
}
