package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/belo4ya/edu-dist-calculate-api/api"
	"github.com/belo4ya/edu-dist-calculate-api/internal/calculator/config"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	httpSwagger "github.com/swaggo/http-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

type HTTPServer struct {
	HTTP  *http.Server
	GWMux *runtime.ServeMux
	conf  *config.Config
}

func NewHTTPServer(conf *config.Config) *HTTPServer {
	mux := http.NewServeMux()
	gwmux := runtime.NewServeMux(runtime.WithForwardResponseOption(httpResponseModifier))
	mux.Handle("/", gwmux)
	handleDocs(mux)
	return &HTTPServer{
		HTTP: &http.Server{
			Addr:    conf.HTTPAddr,
			Handler: mux,
		},
		GWMux: gwmux,
		conf:  conf,
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

const MetadataHeaderHTTPCode = "x-http-code"

func WithHTTPResponseCode(ctx context.Context, code int) {
	_ = grpc.SetHeader(ctx, metadata.Pairs(MetadataHeaderHTTPCode, strconv.Itoa(code)))
}

func httpResponseModifier(ctx context.Context, w http.ResponseWriter, _ proto.Message) error {
	md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		return nil
	}

	if vals := md.HeaderMD.Get(MetadataHeaderHTTPCode); len(vals) > 0 {
		code, err := strconv.Atoi(vals[0])
		if err != nil {
			return err
		}
		delete(md.HeaderMD, MetadataHeaderHTTPCode)
		delete(w.Header(), "Grpc-Metadata-X-Http-Code")
		w.WriteHeader(code)
	}

	return nil
}

func handleDocs(mux *http.ServeMux) {
	mux.HandleFunc("/docs/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(api.OpenAPISpec)
	})
	mux.Handle("/docs/", httpSwagger.Handler(httpSwagger.URL("/docs/openapi.json")))
}
