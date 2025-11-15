package httpserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"avito-tech/configs"
)

type HTTPServer struct {
	httpServer *http.Server
	config     *configs.Config
}

func New(cfg *configs.Config, handler http.Handler) *HTTPServer {
	return &HTTPServer{
		config: cfg,
		httpServer: &http.Server{
			Addr:         cfg.Address(),
			Handler:      handler,
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
			IdleTimeout:  cfg.Server.IdleTimeout,
		},
	}
}

func (s *HTTPServer) Start(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("Starting server on %s", s.httpServer.Addr)

		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Failed to start server: %s", err.Error())
			stop()
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down server...")

	return s.shutdown()
}

func (s *HTTPServer) shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.Server.ShutdownTimeout)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %s", err.Error())
	}

	log.Println("Server exited gracefully")

	return nil
}
