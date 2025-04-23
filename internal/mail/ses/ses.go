package ses

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/glasskube/distr/internal/mail"
	"github.com/glasskube/distr/internal/util"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
)

type sesMailer struct {
	client *ses.Client
	config mail.MailerConfig
}

type Config struct {
	mail.MailerConfig
	Aws *aws.Config
}

var _ mail.Mailer = (&sesMailer{})

func New(config Config) *sesMailer {
	return &sesMailer{
		client: ses.NewFromConfig(*config.Aws),
		config: config.MailerConfig,
	}
}

func NewFromContext(ctx context.Context, config Config) (*sesMailer, error) {
	if cfg, err := awsconfig.LoadDefaultConfig(ctx); err != nil {
		return nil, err
	} else {
		otelaws.AppendMiddlewares(&cfg.APIOptions)
		config.Aws = &cfg
		return New(config), nil
	}
}

// Send implements Mailer.
func (s *sesMailer) Send(ctx context.Context, mail mail.Mail) error {
	message := ses.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses:  mail.To,
			BccAddresses: mail.Bcc,
		},
		Message: &types.Message{
			Subject: &types.Content{Data: &mail.Subject},
			Body:    &types.Body{},
		},
	}
	if mail.From != nil {
		message.Source = util.PtrTo(mail.From.String())
	} else {
		message.Source = util.PtrTo(s.config.DefaultFromAddress.String())
	}
	if mail.ReplyTo != "" {
		message.ReplyToAddresses = []string{mail.ReplyTo}
	}
	if mail.TextBodyFunc != nil {
		if body, err := mail.TextBodyFunc(); err != nil {
			return err
		} else {
			message.Message.Body.Text = &types.Content{Data: &body}
		}
	}
	if mail.HtmlBodyFunc != nil {
		if body, err := mail.HtmlBodyFunc(); err != nil {
			return err
		} else {
			message.Message.Body.Html = &types.Content{Data: &body}
		}
	}
	if _, err := s.client.SendEmail(ctx, &message); err != nil {
		return err
	} else {
		return nil
	}
}
