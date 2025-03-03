package service

import (
	"context"
	"log/slog"

	"github.com/belo4ya/edu-dist-calculate-api/internal/calculator/config"
	calculatorv1 "github.com/belo4ya/edu-dist-calculate-api/pkg/calculator/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AgentService struct {
	calculatorv1.UnimplementedAgentServiceServer
	conf *config.Config
	log  *slog.Logger
}

func NewAgentService(conf *config.Config) *AgentService {
	return &AgentService{
		conf: conf,
		log:  slog.With("logger", "agent-svc"),
	}
}

func (s *AgentService) Register(srv *grpc.Server) {
	calculatorv1.RegisterAgentServiceServer(srv, s)
}

func (s *AgentService) GetTask(context.Context, *emptypb.Empty) (*calculatorv1.GetTaskResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTask not implemented")
}

func (s *AgentService) SubmitTaskResult(context.Context, *calculatorv1.SubmitTaskResultRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitTaskResult not implemented")
}
