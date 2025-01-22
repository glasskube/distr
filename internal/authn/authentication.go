package authn

import (
	"context"
	"errors"
	"net/http"
)

type contextKey struct{}

type Authentication[T any] struct {
	authenticators []RequestAuthenticator[T]
	contextKey     contextKey
}

func New[T any](authenticators ...RequestAuthenticator[T]) *Authentication[T] {
	return &Authentication[T]{authenticators: authenticators, contextKey: contextKey{}}
}

func (a *Authentication[T]) NewContext(ctx context.Context, auth T) context.Context {
	return context.WithValue(ctx, a.contextKey, auth)
}

func (a *Authentication[T]) Get(ctx context.Context) (result T, err error) {
	if auth, ok := ctx.Value(a.contextKey).(T); ok {
		return auth, nil
	} else {
		return result, ErrNoAuthentication
	}
}

func (a *Authentication[T]) Require(ctx context.Context) T {
	if auth, err := a.Get(ctx); err != nil {
		panic(err)
	} else {
		return auth
	}
}

func (a *Authentication[T]) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, provider := range a.authenticators {
			if result, err := provider.Authenticate(r.Context(), r); err != nil {
				if errors.Is(err, ErrBadAuthentication) {
					break
				} else if errors.Is(err, ErrNoAuthentication) {
					continue
				} else {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			} else {
				next.ServeHTTP(w, r.WithContext(a.NewContext(r.Context(), result)))
				return
			}
		}
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	})
}

func (a *Authentication[T]) ValidatorMiddleware(
	fn func(ctx context.Context, value T) error,
) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := fn(r.Context(), a.Require(r.Context())); err != nil {
				if errors.Is(err, ErrBadAuthentication) {
					http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				} else {
					next.ServeHTTP(w, r)
				}
			}
		})
	}
}
