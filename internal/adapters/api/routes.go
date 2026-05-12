package api

import (
	"net/http"
	"strings"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// Public Routes
	mux.HandleFunc("/health", s.healthCheck)
	mux.HandleFunc("/api/v1/register", s.userHandler.Register)
	mux.HandleFunc("/api/v1/login", s.userHandler.Login)
	mux.HandleFunc("/api/v1/refresh", s.userHandler.RefreshToken)
	mux.HandleFunc("/api/v1/logout", s.userHandler.Logout)

	// Protected Routes
	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("/api/v1/profile", s.userHandler.GetProfile)
	protectedMux.HandleFunc("/api/v1/dashboard/summary", s.dashboardHandler.GetSummary)
	protectedMux.HandleFunc("/api/v1/groups", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.groupHandler.CreateGroup(w, r)
			return
		}
		if r.Method == http.MethodGet {
			s.groupHandler.ListGroups(w, r)
			return
		}

		w.WriteHeader(http.StatusMethodNotAllowed)
	})
	protectedMux.HandleFunc("/api/v1/groups/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/images") {
			if r.Method == http.MethodPost {
				s.groupHandler.UploadGroupImages(w, r)
				return
			}
			if r.Method == http.MethodGet {
				s.groupHandler.ListGroupImages(w, r)
				return
			}
		}
		if strings.HasSuffix(r.URL.Path, "/results") {
			s.groupHandler.ListGroupResults(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/exports/csv") {
			s.groupHandler.CreateCSVExport(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/exports") {
			s.groupHandler.ListGroupExports(w, r)
			return
		}
		s.groupHandler.GetGroup(w, r)
	})
	protectedMux.HandleFunc("/api/v1/images/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/result") {
			s.groupHandler.GetImageResult(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/review") {
			if r.Method == http.MethodGet {
				s.groupHandler.GetImageReview(w, r)
				return
			}
			s.groupHandler.SubmitImageReview(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/retry") {
			s.groupHandler.RetryImage(w, r)
			return
		}
		s.groupHandler.GetImage(w, r)
	})

	// Wrap protected routes with AuthMiddleware
	authMiddleware := AuthMiddleware(s.authSvc)

	// Final handler that routes to either mux
	var finalHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/profile" || r.URL.Path == "/api/v1/dashboard/summary" || r.URL.Path == "/api/v1/groups" || strings.HasPrefix(r.URL.Path, "/api/v1/groups/") || strings.HasPrefix(r.URL.Path, "/api/v1/images/") {
			authMiddleware(protectedMux).ServeHTTP(w, r)
			return
		}
		mux.ServeHTTP(w, r)
	})

	return corsMiddleware(finalHandler)
}
