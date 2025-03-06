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
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AgentRepository interface {
	GetPendingTask(context.Context) (modelv2.Task, error)
	FinishTask(context.Context, modelv2.UpdateTaskCmd) error
	ListExpressionTasks(context.Context, string) ([]modelv2.Task, error)
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

func (s *AgentService) ListExpressionTasks(
	ctx context.Context,
	req *calculatorv1.ListExpressionTasksRequest,
) (*calculatorv1.ListExpressionTasksResponse, error) {
	tasks, err := s.repo.ListExpressionTasks(ctx, req.Id)
	if err != nil {
		return nil, InternalError(err)
	}

	resp := &calculatorv1.ListExpressionTasksResponse{
		Tasks: make([]*calculatorv1.ListExpressionTasksResponse_Task, 0, len(tasks)),
	}
	for _, task := range tasks {
		resp.Tasks = append(resp.Tasks, &calculatorv1.ListExpressionTasksResponse_Task{
			Id:             task.ID,
			ExpressionId:   task.ExpressionID,
			ParentTask_1Id: task.ParentTask1ID,
			ParentTask_2Id: task.ParentTask2ID,
			Arg_1:          task.Arg1,
			Arg_2:          task.Arg2,
			Operation:      mapTaskOperation(task.Operation),
			OperationTime:  mapTaskOperationTime(task.Operation, s.conf),
			Status:         mapTaskStatus(task.Status),
			Result:         task.Result,
			ExpireAt:       timestamppb.New(task.ExpireAt),
			CreatedAt:      timestamppb.New(task.CreatedAt),
			UpdatedAt:      timestamppb.New(task.UpdatedAt),
		})
	}
	return resp, nil
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

func mapTaskStatus(s modelv2.TaskStatus) calculatorv1.ListExpressionTasksResponse_TaskStatus {
	switch s {
	case modelv2.TaskStatusPending:
		return calculatorv1.ListExpressionTasksResponse_TASK_STATUS_PENDING
	case modelv2.TaskStatusInProgress:
		return calculatorv1.ListExpressionTasksResponse_TASK_STATUS_IN_PROGRESS
	case modelv2.TaskStatusCompleted:
		return calculatorv1.ListExpressionTasksResponse_TASK_STATUS_COMPLETED
	case modelv2.TaskStatusFailed:
		return calculatorv1.ListExpressionTasksResponse_TASK_STATUS_FAILED
	default:
		return calculatorv1.ListExpressionTasksResponse_TASK_STATUS_UNSPECIFIED
	}
}
