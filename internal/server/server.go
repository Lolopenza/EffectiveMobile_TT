package server

import (
	"context"
	"fmt"
	"net/http"

	"em_tz_anvar/internal/config"
	"em_tz_anvar/internal/handler"
)

type Server struct {
	httpServer *http.Server
}

// NewServer
func NewServer(cfg *config.Config, handlers *handler.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
			Handler:      handlers.InitRoutes(),
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
		},
	}
}

func (s *Server) Run() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
