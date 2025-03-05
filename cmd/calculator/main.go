package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/belo4ya/edu-dist-calculate-api/internal/calculator/calc"
	"github.com/belo4ya/edu-dist-calculate-api/internal/calculator/config"
	"github.com/belo4ya/edu-dist-calculate-api/internal/calculator/repository"
	"github.com/belo4ya/edu-dist-calculate-api/internal/calculator/server"
	"github.com/belo4ya/edu-dist-calculate-api/internal/calculator/service"
	"github.com/belo4ya/edu-dist-calculate-api/internal/logging"
	"github.com/belo4ya/edu-dist-calculate-api/internal/mgmtserver"
	"github.com/belo4ya/runy"
	"github.com/dgraph-io/badger/v4"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func init() {
	_ = godotenv.Load(".env.calculator")
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

	log := slog.Default()
	log.InfoContext(ctx, "logger is configured")
	log.InfoContext(ctx, "config initialized", "config", conf)

	mgmtSrv := mgmtserver.New(&mgmtserver.Config{Addr: conf.MgmtAddr})
	grpcSrv := server.NewGRPCServer(conf)
	httpSrv := server.NewHTTPServer(conf)

	db, err := badger.Open(badger.DefaultOptions(conf.DBBadgerPath))
	if err != nil {
		return fmt.Errorf("open badger: %w", err)
	}
	defer func() {
		_ = db.Close()
	}()

	repo := repository.New(db)

	calcSvc := service.NewCalculatorService(conf, log, calc.NewCalculator(), repo)
	agentSvc := service.NewAgentService(conf, log, repo)
	internalSvc := service.NewInternalService(conf, log, repo)

	for i, svc := range []interface {
		Register(*grpc.Server)
		RegisterGRPCGateway(context.Context, *runtime.ServeMux, []grpc.DialOption) error
	}{calcSvc, agentSvc, internalSvc} {
		svc.Register(grpcSrv.GRPC)
		if err := svc.RegisterGRPCGateway(ctx, httpSrv.GWMux, []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}); err != nil {
			return fmt.Errorf("register grpc gateway %d: %w", i, err)
		}
	}

	runy.Add(mgmtSrv, grpcSrv, httpSrv)
	if err := runy.Start(ctx); err != nil {
		return fmt.Errorf("problem with running app: %w", err)
	}
	return nil
}
