package client

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/belo4ya/edu-dist-calculate-api/internal/agent/config"
	calculatorv1 "github.com/belo4ya/edu-dist-calculate-api/pkg/calculator/v1"
	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/timeout"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

type CalculatorClient struct {
	calculatorv1.CalculatorServiceClient
}

func NewCalculatorClient(ctx context.Context, conf *config.Config) (*CalculatorClient, func(), error) {
	conn, err := grpc.NewClient(
		conf.CalculatorAddr,
		WithCommonGRPCDialOptions(grpc.WithTransportCredentials(insecure.NewCredentials()))...,
	)
	if err != nil {
		return nil, func() {}, fmt.Errorf("init grpc client: %w", err)
	}

	slog.InfoContext(ctx, "connecting to calculator", "addr", conf.CalculatorAddr)
	if err := WaitForReadyGRPCConnection(ctx, conf.GrpcClientConnectTimeout, conn); err != nil {
		return nil, func() {}, fmt.Errorf("wait for ready grpc conn: %w", err)
	}

	cleanup := func() {
		if err := conn.Close(); err != nil {
			slog.ErrorContext(ctx, "failed to close client connection", "error", err)
		}
	}

	return &CalculatorClient{CalculatorServiceClient: calculatorv1.NewCalculatorServiceClient(conn)}, cleanup, nil
}

func WithCommonGRPCDialOptions(opts ...grpc.DialOption) []grpc.DialOption {
	return append(CommonGRPCDialOptions(), opts...)
}

func CommonGRPCDialOptions() []grpc.DialOption {
	clientMetrics := grpcprom.NewClientMetrics(grpcprom.WithClientHandlingTimeHistogram())
	prometheus.MustRegister(clientMetrics)

	return []grpc.DialOption{
		grpc.WithChainUnaryInterceptor(
			timeout.UnaryClientInterceptor(10*time.Second),
			retry.UnaryClientInterceptor(
				retry.WithMax(3),
				retry.WithBackoff(retry.BackoffExponentialWithJitter(200*time.Millisecond, 0.1)),
			),
			clientMetrics.UnaryClientInterceptor(),
		),
	}
}

func WaitForReadyGRPCConnection(ctx context.Context, timeout time.Duration, conn *grpc.ClientConn) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	for {
		state := conn.GetState()
		if state == connectivity.Ready {
			return nil
		}
		if state == connectivity.Idle {
			conn.Connect()
		}
		if !conn.WaitForStateChange(ctx, state) {
			return fmt.Errorf("connect to %s (%s): %w", conn.Target(), state.String(), ctx.Err())
		}
	}
}
