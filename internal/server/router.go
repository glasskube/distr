package server

import (
	"net/http"

	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/handlers"
	"github.com/go-chi/chi/v5"
)

func ApiRouter() chi.Router {
	// TODO for all (most) routes auth middleware
	router := chi.NewRouter()
	router.Use(dbCtxMiddleware)
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
