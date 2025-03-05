package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/belo4ya/edu-dist-calculate-api/internal/calculator/calc"
	"github.com/belo4ya/edu-dist-calculate-api/internal/calculator/config"
	"github.com/belo4ya/edu-dist-calculate-api/internal/calculator/repository/modelv2"
	calculatorv1 "github.com/belo4ya/edu-dist-calculate-api/pkg/calculator/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Calculator interface {
	Parse(string) ([]calc.Token, error)
	Schedule([]calc.Token) []calc.Task
}

type CalculatorRepository interface {
	CreateExpression(context.Context, modelv2.CreateExpressionCmd, []modelv2.CreateExpressionTaskCmd) (string, error)
	ListExpressions(context.Context) ([]modelv2.Expression, error)
	GetExpression(context.Context, string) (modelv2.Expression, error)
}

type CalculatorService struct {
	calculatorv1.UnimplementedCalculatorServiceServer
	conf *config.Config
	log  *slog.Logger
	calc Calculator
	repo CalculatorRepository
}

func NewCalculatorService(conf *config.Config, calc Calculator, repo CalculatorRepository) *CalculatorService {
	return &CalculatorService{
		conf: conf,
		log:  slog.With("logger", "calculator-service"),
		calc: calc,
		repo: repo,
	}
}

func (s *CalculatorService) Register(srv *grpc.Server) {
	calculatorv1.RegisterCalculatorServiceServer(srv, s)
}

func (s *CalculatorService) RegisterGRPCGateway(ctx context.Context, mux *runtime.ServeMux, addr string, clientOpts []grpc.DialOption) error {
	return calculatorv1.RegisterCalculatorServiceHandlerFromEndpoint(ctx, mux, "localhost"+addr, clientOpts)
}

func (s *CalculatorService) Calculate(ctx context.Context, req *calculatorv1.CalculateRequest) (*calculatorv1.CalculateResponse, error) {
	parsed, err := s.calc.Parse(req.Expression)
	if err != nil {
		if errors.Is(err, calc.ErrInvalidExpr) {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid expression %q: %s", req.Expression, err.Error()))
		}
		return nil, InternalError(err)
	}

	tasks := s.calc.Schedule(parsed)

	createExpr := modelv2.CreateExpressionCmd{Expression: req.Expression}
	createTasks := make([]modelv2.CreateExpressionTaskCmd, 0, len(tasks))
	for _, t := range tasks {
		createTasks = append(createTasks, modelv2.CreateExpressionTaskCmd{
			ID:            t.ID,
			ParentTask1ID: t.ParentTask1ID,
			ParentTask2ID: t.ParentTask2ID,
			Arg1:          t.Arg1,
			Arg2:          t.Arg2,
			Operation:     modelv2.TaskOperation(t.Operation),
		})
	}

	id, err := s.repo.CreateExpression(ctx, createExpr, createTasks)
	if err != nil {
		return nil, InternalError(err)
	}

	return &calculatorv1.CalculateResponse{Id: id}, nil
}

func (s *CalculatorService) ListExpressions(ctx context.Context, _ *emptypb.Empty) (*calculatorv1.ListExpressionsResponse, error) {
	exprs, err := s.repo.ListExpressions(ctx)
	if err != nil {
		return nil, InternalError(err)
	}

	resp := &calculatorv1.ListExpressionsResponse{Expressions: make([]*calculatorv1.Expression, 0, len(exprs))}
	for _, expr := range exprs {
		resp.Expressions = append(resp.Expressions, &calculatorv1.Expression{
			Id:     expr.ID,
			Status: mapExprStatus(expr.Status),
			Result: expr.Result,
		})
	}
	return resp, nil
}

func (s *CalculatorService) GetExpression(ctx context.Context, req *calculatorv1.GetExpressionRequest) (*calculatorv1.GetExpressionResponse, error) {
	expr, err := s.repo.GetExpression(ctx, req.Id)
	if err != nil {
		if errors.Is(err, nil) {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("expression %q not found", req.Id))
		}
		return nil, InternalError(err)
	}

	return &calculatorv1.GetExpressionResponse{
		Expression: &calculatorv1.Expression{
			Id:     expr.ID,
			Status: mapExprStatus(expr.Status),
			Result: expr.Result,
		},
	}, nil
}

func mapExprStatus(s modelv2.ExpressionStatus) calculatorv1.ExpressionStatus {
	if s == modelv2.ExpressionStatusPending {
		return calculatorv1.ExpressionStatus_EXPRESSION_STATUS_PENDING
	}
	if s == modelv2.ExpressionStatusInProgress {
		return calculatorv1.ExpressionStatus_EXPRESSION_STATUS_IN_PROGRESS
	}
	if s == modelv2.ExpressionStatusCompleted {
		return calculatorv1.ExpressionStatus_EXPRESSION_STATUS_COMPLETED
	}
	if s == modelv2.ExpressionStatusFailed {
		return calculatorv1.ExpressionStatus_EXPRESSION_STATUS_FAILED
	}
	return calculatorv1.ExpressionStatus_EXPRESSION_STATUS_UNSPECIFIED
}
