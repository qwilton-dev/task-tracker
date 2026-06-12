package middleware

import (
	"net/http"
	"strings"

	"github.com/rs/cors"
)

func CORS(allowOrigins string) func(http.Handler) http.Handler {
	origins := strings.Split(allowOrigins, ",")
	allowed := make([]string, 0, len(origins))
	for _, o := range origins {
		o = strings.TrimSpace(o)
		if o != "" {
			allowed = append(allowed, o)
		}
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   allowed,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           86400,
	})

	return func(next http.Handler) http.Handler {
		return c.Handler(next)
	}
}
