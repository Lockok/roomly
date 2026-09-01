package middleware

import (
	"log"
	"net/http"

	"github.com/Lockok/roomly/internal/platform/httputil"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf(
					"panic recovered: request_id=%s method=%s path=%s error=%v",
					GetRequestID(r.Context()),
					r.Method,
					r.URL.Path,
					recovered,
				)

				httputil.WriteError(
					w,
					http.StatusInternalServerError,
					"INTERNAL ERROR",
					"internal server error",
				)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
