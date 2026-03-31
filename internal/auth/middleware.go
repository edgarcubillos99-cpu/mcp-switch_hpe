package auth

import (
	"net/http"
	"strings"
)

// APIKeyMiddleware intercepta las peticiones y valida el token
func APIKeyMiddleware(validKey string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		// Esperamos el formato "Bearer <api_key>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[1] != validKey {
			http.Error(w, "Invalid API Key", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
