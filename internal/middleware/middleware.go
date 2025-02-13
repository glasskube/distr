package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/glasskube/distr/internal/auth"
	"github.com/glasskube/distr/internal/authkey"
	"github.com/glasskube/distr/internal/authn"
	"github.com/glasskube/distr/internal/authn/authinfo"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/mail"
	"github.com/glasskube/distr/internal/types"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"go.uber.org/zap"
)

func ContextInjectorMiddleware(db *pgxpool.Pool, mailer mail.Mailer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = internalctx.WithDb(ctx, db)
			ctx = internalctx.WithMailer(ctx, mailer)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func LoggerCtxMiddleware(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := logger.With(zap.String("requestId", middleware.GetReqID(r.Context())))
			ctx := internalctx.WithLogger(r.Context(), logger)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func LoggingMiddleware(handler http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		now := time.Now()
		handler.ServeHTTP(ww, r)
		elapsed := time.Since(now)
		logger := internalctx.GetLogger(r.Context())
		logger.Info("handling request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", ww.Status()),
			zap.String("time", elapsed.String()))
	}
	return http.HandlerFunc(fn)
}

func UserRoleMiddleware(userRole types.UserRole) func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			if auth, err := auth.Authentication.Get(ctx); err != nil {
				http.Error(w, err.Error(), http.StatusForbidden)
			} else if auth.CurrentUserRole() == nil || *auth.CurrentUserRole() != userRole {
				http.Error(w, "insufficient permissions", http.StatusForbidden)
			} else {
				handler.ServeHTTP(w, r)
			}
		}
		return http.HandlerFunc(fn)
	}
}

var Sentry = sentryhttp.New(sentryhttp.Options{Repanic: true}).Handle

func SentryUser(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if hub := sentry.GetHubFromContext(ctx); hub != nil {
			if auth, err := auth.Authentication.Get(ctx); err == nil {
				hub.Scope().SetUser(sentry.User{
					ID:    auth.CurrentUserID().String(),
					Email: auth.CurrentUserEmail(),
				})
			}
		}
		h.ServeHTTP(w, r)
	})
}

func RateLimitCurrentUserIdKeyFunc(r *http.Request) (string, error) {
	if auth, err := auth.Authentication.Get(r.Context()); err != nil {
		return "", err
	} else {
		prefix := ""
		switch auth.Token().(type) {
		case jwt.Token:
			prefix = "jwt"
		case authkey.Key:
			prefix = "authkey"
		}
		return fmt.Sprintf("%v-%v", prefix, auth.CurrentUserID()), nil
	}
}

var RequireOrgID = auth.Authentication.ValidatorMiddleware(func(value authinfo.AuthInfo) error {
	if value.CurrentOrgID() == nil {
		return authn.ErrBadAuthentication
	} else {
		return nil
	}
})

var RequireUserRole = auth.Authentication.ValidatorMiddleware(func(value authinfo.AuthInfo) error {
	if value.CurrentUserRole() == nil {
		return authn.ErrBadAuthentication
	} else {
		return nil
	}
})
