package api

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/cashflow/auth-service/internal/adapters/handlers"
	"github.com/cashflow/auth-service/internal/core/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	port        string
	userHandler *handlers.UserHandler
	authSvc     ports.AuthService
	db          *pgxpool.Pool
	httpServer  *http.Server
}

func NewServer(port string, userHandler *handlers.UserHandler, authSvc ports.AuthService, db *pgxpool.Pool) *Server {
	return &Server{
		port:        port,
		userHandler: userHandler,
		authSvc:     authSvc,
		db:          db,
	}
}

func (s *Server) Start() error {
	handler := s.RegisterRoutes()

	s.httpServer = &http.Server{
		Addr:    ":" + s.port,
		Handler: handler,
	}

	log.Printf("🚀 Auth Service running on :%s (Hexagonal Architecture)", s.port)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Stopping HTTP server...")
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.db == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"status":"unhealthy","database":"unavailable"}`)
		return
	}

	if err := s.db.Ping(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"status":"unhealthy","database":"disconnected"}`)
		return
	}
	fmt.Fprintf(w, `{"status":"healthy","database":"connected"}`)
}
