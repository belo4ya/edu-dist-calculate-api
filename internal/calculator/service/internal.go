package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/belo4ya/edu-dist-calculate-api/internal/calculator/config"
	"github.com/belo4ya/edu-dist-calculate-api/internal/calculator/repository/models"
	"github.com/belo4ya/edu-dist-calculate-api/internal/logging"
	calculatorv1 "github.com/belo4ya/edu-dist-calculate-api/pkg/calculator/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InternalRepository interface {
	ListExpressionTasks(context.Context, string) ([]models.Task, error)
}

type InternalService struct {
	calculatorv1.UnimplementedInternalServiceServer
	conf *config.Config
	log  *slog.Logger
	repo InternalRepository
}

func NewInternalService(conf *config.Config, log *slog.Logger, repo InternalRepository) *InternalService {
	return &InternalService{
		conf: conf,
		log:  logging.WithName(log, "internal-service"),
		repo: repo,
	}
}

func (s *InternalService) Register(srv *grpc.Server) {
	calculatorv1.RegisterInternalServiceServer(srv, s)
}

func (s *InternalService) RegisterGRPCGateway(ctx context.Context, mux *runtime.ServeMux, clientOpts []grpc.DialOption) error {
	return calculatorv1.RegisterInternalServiceHandlerFromEndpoint(ctx, mux, "localhost"+s.conf.GRPCAddr, clientOpts)
}

func (s *InternalService) ListExpressionTasks(
	ctx context.Context,
	req *calculatorv1.ListExpressionTasksRequest,
) (*calculatorv1.ListExpressionTasksResponse, error) {
	tasks, err := s.repo.ListExpressionTasks(ctx, req.Id)
	if err != nil {
		if errors.Is(err, models.ErrExpressionNotFound) {
			return nil, status.Error(codes.NotFound, "expression not found")
		}
		return nil, InternalError(fmt.Errorf("list expression tasks: %w", err))
	}

	resp := &calculatorv1.ListExpressionTasksResponse{
		Tasks: make([]*calculatorv1.ListExpressionTasksResponse_Task, 0, len(tasks)),
	}
	for _, task := range tasks {
		resp.Tasks = append(resp.Tasks, mapTaskToInternalTaskResponse(task))
	}
	return resp, nil
}
