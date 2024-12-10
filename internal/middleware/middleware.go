package middleware

import (
	"net/http"
	"time"

	"github.com/glasskube/cloud/internal/auth"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/mail"
	"github.com/glasskube/cloud/internal/types"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func ContextInjectorMiddelware(db *pgxpool.Pool, mailer mail.Mailer) func(next http.Handler) http.Handler {
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
			if currentRole, err := auth.CurrentUserRole(ctx); err != nil {
				http.Error(w, err.Error(), http.StatusForbidden)
			} else if currentRole != userRole {
				http.Error(w, "insufficient permissions", http.StatusForbidden)
			} else {
				handler.ServeHTTP(w, r)
			}
		}
		return http.HandlerFunc(fn)
	}
}
