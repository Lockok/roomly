package health

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Lockok/roomly/internal/platform/httputil"
)

func Live(w http.ResponseWriter, _ *http.Request) {
	httputil.WriteJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func Ready(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			httputil.WriteError(
				w,
				http.StatusServiceUnavailable,
				"DATABASE_UNAVAILABLE",
				"database is unavailable",
			)
			return
		}

		httputil.WriteJSON(w, http.StatusOK, map[string]string{
			"status": "ok",
		})
	}
}
