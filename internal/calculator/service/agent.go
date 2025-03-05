package service

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/belo4ya/edu-dist-calculate-api/internal/calculator/config"
	"github.com/belo4ya/edu-dist-calculate-api/internal/calculator/model"
	calculatorv1 "github.com/belo4ya/edu-dist-calculate-api/pkg/calculator/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AgentRepository interface {
	GetTask(context.Context) (model.Task, error)
	UpdateTask(context.Context, *calculatorv1.SubmitTaskResultRequest) error
}

type AgentService struct {
	calculatorv1.UnimplementedAgentServiceServer
	conf *config.Config
	log  *slog.Logger
	repo AgentRepository
}

func NewAgentService(conf *config.Config, repo AgentRepository) *AgentService {
	return &AgentService{
		conf: conf,
		log:  slog.With("logger", "agent-service"),
		repo: repo,
	}
}

func (s *AgentService) Register(srv *grpc.Server) {
	calculatorv1.RegisterAgentServiceServer(srv, s)
}

func (s *AgentService) RegisterGRPCGateway(ctx context.Context, mux *runtime.ServeMux, addr string, clientOpts []grpc.DialOption) error {
	return calculatorv1.RegisterAgentServiceHandlerFromEndpoint(ctx, mux, "localhost"+addr, clientOpts)
}

func (s *AgentService) GetTask(ctx context.Context, _ *emptypb.Empty) (*calculatorv1.GetTaskResponse, error) {
	task, err := s.repo.GetTask(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "no tasks yet, try later")
		}
		return nil, InternalError(err)
	}

	_ = task
	return &calculatorv1.GetTaskResponse{Task: &calculatorv1.Task{
		Id:            "",
		Arg1:          0,
		Arg2:          0,
		Operation:     0,
		OperationTime: nil,
	}}, nil
}

func (s *AgentService) SubmitTaskResult(ctx context.Context, req *calculatorv1.SubmitTaskResultRequest) (*emptypb.Empty, error) {
	if err := s.repo.UpdateTask(ctx, req); err != nil {
		return nil, InternalError(err)
	}
	return &emptypb.Empty{}, nil
}
