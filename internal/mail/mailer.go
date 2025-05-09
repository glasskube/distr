package mail

import (
	"context"
)

type Mailer interface {
	Send(ctx context.Context, mail Mail) error
}

type FromAddressSrcFn = func(ctx context.Context, mail Mail) string

type MailerConfig struct {
	FromAddressSrc []FromAddressSrcFn
}

func StaticFromAddress(address string) FromAddressSrcFn {
	return func(ctx context.Context, mail Mail) string {
		return address
	}
}

func MailOverrideFromAddress() FromAddressSrcFn {
	return func(ctx context.Context, mail Mail) string {
		if mail.From != nil {
			return mail.From.String()
		}
		return ""
	}
}

func (mc *MailerConfig) GetActualFromAddress(ctx context.Context, mail Mail) string {
	for _, fn := range mc.FromAddressSrc {
		if a := fn(ctx, mail); a != "" {
			return a
		}
	}
	return ""
}
