package authn

import (
	"context"
	"errors"
	"net/http"

	"go.uber.org/multierr"
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

func Chain3[A any, B any, C any, D any](
	a Authenticator[A, B],
	b Authenticator[B, C],
	c Authenticator[C, D],
) Authenticator[A, D] {
	return AuthenticatorFunc[A, D](func(ctx context.Context, aa A) (out D, err error) {
		if bb, err1 := a.Authenticate(ctx, aa); err1 != nil {
			err = err1
		} else if cc, err1 := b.Authenticate(ctx, bb); err1 != nil {
			err = err1
		} else {
			out, err = c.Authenticate(ctx, cc)
		}
		return
	})
}

func Alternative[A any, B any](authenticators ...Authenticator[A, B]) Authenticator[A, B] {
	return AuthenticatorFunc[A, B](func(ctx context.Context, in A) (result B, err error) {
		for _, authenticator := range authenticators {
			if out, err1 := authenticator.Authenticate(ctx, in); err1 != nil {
				multierr.AppendInto(&err, err1)
				if errors.Is(err, ErrNoAuthentication) || errors.Is(err, ErrBadAuthentication) {
					continue
				} else {
					break
				}
			} else {
				result = out
				err = nil
				break
			}
		}
		return
	})
}
