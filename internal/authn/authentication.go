package authn

import (
	"context"
	"errors"
	"net/http"
)

type contextKey struct{}

type Authentication[T any] struct {
	authenticators      []RequestAuthenticator[T]
	contextKey          contextKey
	unknownErrorHandler func(w http.ResponseWriter, r *http.Request, err error)
}

func New[T any](authenticators ...RequestAuthenticator[T]) *Authentication[T] {
	return &Authentication[T]{authenticators: authenticators, contextKey: contextKey{}}
}

func (a *Authentication[T]) SetUnknownErrorHandler(handler func(w http.ResponseWriter, r *http.Request, err error)) {
	a.unknownErrorHandler = handler
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
		var err error
		for _, provider := range a.authenticators {
			var result T
			if result, err = provider.Authenticate(r.Context(), r); err != nil {
				if errors.Is(err, ErrNoAuthentication) {
					continue
				} else {
					break
				}
			} else {
				next.ServeHTTP(w, r.WithContext(a.NewContext(r.Context(), result)))
				return
			}
		}

		a.handleError(w, r, err)
	})
}

func (a *Authentication[T]) ValidatorMiddleware(fn func(value T) error) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := fn(a.Require(r.Context())); err != nil {
				a.handleError(w, r, err)
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}

func (a *Authentication[T]) handleError(w http.ResponseWriter, r *http.Request, err error) {
	hhe := &HttpHeaderError{}
	if errors.As(err, &hhe) {
		hhe.WriteTo(w)
	}

	if errors.Is(err, ErrBadAuthentication) || errors.Is(err, ErrNoAuthentication) {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	} else if a.unknownErrorHandler == nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	} else {
		a.unknownErrorHandler(w, r, err)
	}
}
