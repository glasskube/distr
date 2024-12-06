package mail

import "context"

type Mailer interface {
	Send(ctx context.Context, mail Mail) error
}

type MailerConfig struct {
	FromAddress string
}
