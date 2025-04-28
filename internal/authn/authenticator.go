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

func Chain3[IN any, MID1 any, MID2 any, OUT any](
	a Authenticator[IN, MID1],
	b Authenticator[MID1, MID2],
	c Authenticator[MID2, OUT],
) Authenticator[IN, OUT] {
	return AuthenticatorFunc[IN, OUT](func(ctx context.Context, in IN) (out OUT, err error) {
		if mid1, err := a.Authenticate(ctx, in); err != nil {
			return out, err
		} else if mid2, err := b.Authenticate(ctx, mid1); err != nil {
			return out, err
		} else {
			return c.Authenticate(ctx, mid2)
		}
	})
}

func Chain4[IN any, MID1 any, MID2 any, MID3 any, OUT any](
	a Authenticator[IN, MID1],
	b Authenticator[MID1, MID2],
	c Authenticator[MID2, MID3],
	d Authenticator[MID3, OUT],
) Authenticator[IN, OUT] {
	return AuthenticatorFunc[IN, OUT](func(ctx context.Context, in IN) (out OUT, err error) {
		if mid1, err := a.Authenticate(ctx, in); err != nil {
			return out, err
		} else if mid2, err := b.Authenticate(ctx, mid1); err != nil {
			return out, err
		} else if mid3, err := c.Authenticate(ctx, mid2); err != nil {
			return out, err
		} else {
			return d.Authenticate(ctx, mid3)
		}
	})
}

func Alternative[A any, B any](authenticators ...Authenticator[A, B]) Authenticator[A, B] {
	return AuthenticatorFunc[A, B](func(ctx context.Context, in A) (result B, err error) {
		for _, authenticator := range authenticators {
			if out, err1 := authenticator.Authenticate(ctx, in); err1 != nil {
				multierr.AppendInto(&err, err1)
				if errors.Is(err1, ErrNoAuthentication) || errors.Is(err1, ErrBadAuthentication) {
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
