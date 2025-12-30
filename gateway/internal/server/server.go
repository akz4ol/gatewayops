// Package server provides the HTTP server for the gateway.
package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/akz4ol/gatewayops/gateway/internal/config"
	"github.com/rs/zerolog"
)

// Server represents the HTTP server.
type Server struct {
	httpServer *http.Server
	config     *config.Config
	logger     zerolog.Logger
}

// New creates a new HTTP server.
func New(cfg *config.Config, handler http.Handler, logger zerolog.Logger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         ":" + cfg.Server.Port,
			Handler:      handler,
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
			IdleTimeout:  cfg.Server.IdleTimeout,
		},
		config: cfg,
		logger: logger,
	}
}

// Start starts the HTTP server and blocks until shutdown.
func (s *Server) Start() error {
	// Channel to listen for errors from server
	serverErrors := make(chan error, 1)

	// Start the server
	go func() {
		s.logger.Info().
			Str("addr", s.httpServer.Addr).
			Str("env", s.config.Server.Env).
			Msg("Starting HTTP server")

		serverErrors <- s.httpServer.ListenAndServe()
	}()

	// Channel to listen for OS signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal or an error
	select {
	case err := <-serverErrors:
		if err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}

	case sig := <-shutdown:
		s.logger.Info().
			Str("signal", sig.String()).
			Msg("Received shutdown signal")

		// Give outstanding requests time to complete
		ctx, cancel := context.WithTimeout(context.Background(), s.config.Server.ShutdownTimeout)
		defer cancel()

		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.logger.Error().
				Err(err).
				Msg("Graceful shutdown failed, forcing shutdown")

			if err := s.httpServer.Close(); err != nil {
				return fmt.Errorf("force close failed: %w", err)
			}
		}

		s.logger.Info().Msg("Server shutdown complete")
	}

	return nil
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// Ready checks if the server is ready to accept requests.
func (s *Server) Ready() bool {
	// Add any dependency checks here (database, redis, etc.)
	return true
}

// Health checks if the server is healthy.
func (s *Server) Health() bool {
	return true
}

// Addr returns the server address.
func (s *Server) Addr() string {
	return s.httpServer.Addr
}

// StartedAt returns when the server was started (for uptime calculation).
var startedAt = time.Now()

// Uptime returns how long the server has been running.
func Uptime() time.Duration {
	return time.Since(startedAt)
}
