package authn

import (
	"context"
	"net/http"
)

type Authenticator[IN any, OUT any] interface {
	Authenticate(ctx context.Context, r IN) (OUT, error)
}

type RequestAuthenticator[T any] Authenticator[*http.Request, T]

type AuthenticatorFunc[IN any, OUT any] func(ctx context.Context, r IN) (OUT, error)

func (af AuthenticatorFunc[IN, OUT]) Authenticate(ctx context.Context, r IN) (OUT, error) {
	return af(ctx, r)
}

func Chain[IN any, MID any, OUT any](a Authenticator[IN, MID], b Authenticator[MID, OUT]) Authenticator[IN, OUT] {
	return AuthenticatorFunc[IN, OUT](func(ctx context.Context, in IN) (out OUT, err error) {
		if tmp, err := a.Authenticate(ctx, in); err != nil {
			return out, err
		} else {
			return b.Authenticate(ctx, tmp)
		}
	})
}
