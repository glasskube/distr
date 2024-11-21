package server

import (
	"github.com/go-chi/cors"
	"net/http"

	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/handlers"
	"github.com/go-chi/chi/v5"
)

func ApiRouter() chi.Router {
	// TODO for all (most) routes auth middleware
	router := chi.NewRouter()
	router.Use(loggerCtxMiddleware, dbCtxMiddleware)
	router.Use(corsMiddleware())
	router.Route("/applications", handlers.ApplicationsRouter)
	return router
}

func dbCtxMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db := getDbPool()
		ctx := internalctx.WithDb(r.Context(), db)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func loggerCtxMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := getLogger()
		ctx := internalctx.WithLogger(r.Context(), logger)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func corsMiddleware() func(next http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://localhost:4200"}, // TODO allow localhost only during dev
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	})
}
