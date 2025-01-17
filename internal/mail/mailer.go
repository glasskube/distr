package mail

import (
	"context"
	"net/mail"
)

type Mailer interface {
	Send(ctx context.Context, mail Mail) error
}

type MailerConfig struct {
	DefaultFromAddress mail.Address
}
