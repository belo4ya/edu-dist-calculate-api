package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/belo4ya/edu-dist-calculate-api/internal/calculator/config"
	"github.com/belo4ya/edu-dist-calculate-api/internal/calculator/server"
	"github.com/belo4ya/edu-dist-calculate-api/internal/calculator/service"
	"github.com/belo4ya/edu-dist-calculate-api/internal/logging"
	"github.com/belo4ya/edu-dist-calculate-api/internal/mgmtserver"
	"github.com/belo4ya/runy"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func init() {
	_ = godotenv.Load(".env")
}

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func run() error {
	ctx := runy.SetupSignalHandler(context.Background())

	conf, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if err := logging.Configure(&logging.Config{Level: conf.LogLevel}); err != nil {
		return fmt.Errorf("configure logging: %w", err)
	}

	slog.InfoContext(ctx, "logger is configured")
	slog.InfoContext(ctx, "config initialized", "config", conf)

	mgmtSrv := mgmtserver.New(&mgmtserver.Config{Addr: conf.MgmtAddr})
	grpcSrv := server.NewGRPCServer(conf)
	httpSrv := server.NewHTTPServer(conf)

	clientOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	calcSvc := service.NewCalculatorService(conf)
	calcSvc.Register(grpcSrv.GRPC)
	if err := calcSvc.RegisterGRPCGateway(ctx, httpSrv.Mux, conf.GRPCAddr, clientOpts); err != nil {
		return fmt.Errorf("calculator service: register grpc gateway: %w", err)
	}

	agentSvc := service.NewAgentService(conf)
	agentSvc.Register(grpcSrv.GRPC)
	if err := agentSvc.RegisterGRPCGateway(ctx, httpSrv.Mux, conf.GRPCAddr, clientOpts); err != nil {
		return fmt.Errorf("agent service: register grpc gateway: %w", err)
	}

	runy.Add(mgmtSrv, grpcSrv, httpSrv)
	if err := runy.Start(ctx); err != nil {
		return fmt.Errorf("problem with running app: %w", err)
	}
	return nil
}
