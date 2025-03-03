package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/belo4ya/edu-dist-calculate-api/internal/calculator/config"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

type HTTPServer struct {
	HTTP *http.Server
	Mux  *runtime.ServeMux
	conf *config.Config
}

func NewHTTPServer(conf *config.Config) *HTTPServer {
	mux := runtime.NewServeMux()
	return &HTTPServer{
		HTTP: &http.Server{
			Addr:    conf.HTTPAddr,
			Handler: mux,
		},
		Mux:  mux,
		conf: conf,
	}
}

func (s *HTTPServer) Start(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		slog.InfoContext(ctx, "http server start listening on "+s.conf.HTTPAddr)
		if err := s.HTTP.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("start http server: %w", err)
		}
		close(errCh)
	}()
	select {
	case <-ctx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		slog.InfoContext(ctx, "shutting down http server")
		if err := s.HTTP.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutdown http server: %w", err)
		}
		return nil
	case err := <-errCh:
		return err
	}
}
