package api

import "net/http"

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// Register Routes
	mux.HandleFunc("/health", s.healthCheck)
	mux.HandleFunc("/api/v1/register", s.userHandler.Register)
	mux.HandleFunc("/api/v1/login", s.userHandler.Login)
	mux.HandleFunc("/api/v1/forgot-password", s.userHandler.ForgotPassword)
	mux.HandleFunc("/api/v1/auth/social", s.userHandler.SocialLogin)

	return mux
}
