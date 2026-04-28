package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/cashflow/auth-service/internal/core/ports"
)

const UserIDKey = "user_id"

func AuthMiddleware(authSvc ports.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Authorization header format must be Bearer {token}", http.StatusUnauthorized)
				return
			}

			token := parts[1]
			userID, err := authSvc.ValidateToken(token)
			if err != nil {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from localhost:5173 and Vercel app
		origin := r.Header.Get("Origin")
		if origin == "http://localhost:5173" || origin == "https://ocr-web-weld.vercel.app" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			// Fallback to localhost if no origin or other
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
