package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/belo4ya/edu-dist-calculate-api/internal/calculator/config"
)

type HTTPServer struct {
	HTTP *http.Server
	conf *config.Config
}

func NewHTTPServer(conf *config.Config) *HTTPServer {
	srv := &HTTPServer{conf: conf}
	srv.HTTP = &http.Server{
		Addr: conf.HTTPAddr,
		//Handler: nil,
	}
	return srv
}

func (s *HTTPServer) Start(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		slog.InfoContext(ctx, fmt.Sprintf("http server start listening on: %s", s.conf.HTTPAddr))
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
