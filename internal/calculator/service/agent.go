package service

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"math"
	"time"

	"github.com/belo4ya/edu-dist-calculate-api/internal/calculator/config"
	"github.com/belo4ya/edu-dist-calculate-api/internal/calculator/repository/modelv2"
	calculatorv1 "github.com/belo4ya/edu-dist-calculate-api/pkg/calculator/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AgentRepository interface {
	GetPendingTask(context.Context) (modelv2.Task, error)
	FinishTask(context.Context, modelv2.UpdateTaskCmd) error
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
	task, err := s.repo.GetPendingTask(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "no tasks yet, try later")
		}
		return nil, InternalError(err)
	}

	return &calculatorv1.GetTaskResponse{Task: &calculatorv1.Task{
		Id:            task.ID,
		Arg1:          task.Arg1,
		Arg2:          task.Arg2,
		Operation:     mapTaskOperation(task.Operation),
		OperationTime: mapTaskOperationTime(task.Operation, s.conf),
	}}, nil
}

func (s *AgentService) SubmitTaskResult(ctx context.Context, req *calculatorv1.SubmitTaskResultRequest) (*emptypb.Empty, error) {
	var updateTask modelv2.UpdateTaskCmd
	if math.IsNaN(req.Result) {
		updateTask = modelv2.UpdateTaskCmd{
			ID:     req.Id,
			Status: modelv2.TaskStatusFailed,
			Result: 0,
		}
	} else {
		updateTask = modelv2.UpdateTaskCmd{
			ID:     req.Id,
			Status: modelv2.TaskStatusCompleted,
			Result: req.Result,
		}
	}

	if err := s.repo.FinishTask(ctx, updateTask); err != nil {
		return nil, InternalError(err)
	}
	return &emptypb.Empty{}, nil
}

func mapTaskOperation(s modelv2.TaskOperation) calculatorv1.TaskOperation {
	switch s {
	case modelv2.TaskOperationAddition:
		return calculatorv1.TaskOperation_TASK_OPERATION_ADDITION
	case modelv2.TaskOperationSubtraction:
		return calculatorv1.TaskOperation_TASK_OPERATION_SUBTRACTION
	case modelv2.TaskOperationMultiplication:
		return calculatorv1.TaskOperation_TASK_OPERATION_MULTIPLICATION
	case modelv2.TaskOperationDivision:
		return calculatorv1.TaskOperation_TASK_OPERATION_DIVISION
	default:
		return calculatorv1.TaskOperation_TASK_OPERATION_UNSPECIFIED
	}
}

func mapTaskOperationTime(op modelv2.TaskOperation, conf *config.Config) *durationpb.Duration {
	ms := 0
	switch op {
	case modelv2.TaskOperationAddition:
		ms = conf.TimeAdditionMs
	case modelv2.TaskOperationSubtraction:
		ms = conf.TimeSubtractionMs
	case modelv2.TaskOperationMultiplication:
		ms = conf.TimeMultiplicationMs
	case modelv2.TaskOperationDivision:
		ms = conf.TimeDivisionMs
	}
	return durationpb.New(time.Duration(ms) * time.Millisecond)
}
