package api

import "net/http"

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// Public Routes
	mux.HandleFunc("/health", s.healthCheck)
	mux.HandleFunc("/api/v1/register", s.userHandler.Register)
	mux.HandleFunc("/api/v1/login", s.userHandler.Login)
	mux.HandleFunc("/api/v1/refresh", s.userHandler.RefreshToken)

	// Protected Routes
	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("/api/v1/profile", s.userHandler.GetProfile)

	// Wrap protected routes with AuthMiddleware
	authMiddleware := AuthMiddleware(s.authSvc)
	
	// Final handler that routes to either mux
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/profile" {
			authMiddleware(protectedMux).ServeHTTP(w, r)
			return
		}
		mux.ServeHTTP(w, r)
	})
}
