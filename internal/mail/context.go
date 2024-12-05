package mail

import "context"

type contextKey int

const key contextKey = iota

func AddToContext(ctx context.Context, mailer Mailer) context.Context {
	return context.WithValue(ctx, key, mailer)
}

func FromContext(ctx context.Context) Mailer {
	if mailer, ok := ctx.Value(key).(Mailer); ok {
		return mailer
	} else {
		panic("context has no instance of Mailer")
	}
}
