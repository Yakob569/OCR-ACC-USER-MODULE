package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/cashflow/auth-service/internal/adapters/handlers"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	port        string
	userHandler *handlers.UserHandler
	db          *pgxpool.Pool
}

func NewServer(port string, userHandler *handlers.UserHandler, db *pgxpool.Pool) *Server {
	return &Server{
		port:        port,
		userHandler: userHandler,
		db:          db,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Register Routes
	mux.HandleFunc("/health", s.healthCheck)
	mux.HandleFunc("/api/v1/register", s.userHandler.Register)

	log.Printf("🚀 Auth Service running on :%s (Hexagonal Architecture)", s.port)
	return http.ListenAndServe(":"+s.port, mux)
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	if err := s.db.Ping(r.Context()); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"status":"unhealthy","database":"disconnected"}`)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"healthy","database":"connected"}`)
}
