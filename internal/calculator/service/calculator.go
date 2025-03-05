package service

import (
	"context"
	"errors"
	"fmt"
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

type Calculator interface {
	Parse(string) ([]model.Token, error)
	Schedule([]model.Token) []model.Task
}

type CalculatorRepository interface {
	CreateExpression(context.Context, model.Expression, []model.Task) (string, error)
	ListExpressions(context.Context) ([]model.Expression, error)
	GetExpression(context.Context, string) (model.Expression, error)
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
		if errors.Is(err, nil) {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid expression %q: %s", req.Expression, err.Error()))
		}
		return nil, InternalError(err)
	}

	tasks := s.calc.Schedule(parsed)

	expr := model.Expression{
		ID:           0,
		Expression:   "",
		Status:       "",
		Result:       0,
		ErrorMessage: "",
	}

	id, err := s.repo.CreateExpression(ctx, expr, tasks)
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

	resp := &calculatorv1.ListExpressionsResponse{Expressions: make([]*calculatorv1.Expression, len(exprs))}
	for i, expr := range exprs {
		_ = expr
		resp.Expressions[i] = &calculatorv1.Expression{
			Id:     "",
			Status: 0,
			Result: nil,
		}
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

	_ = expr
	return &calculatorv1.GetExpressionResponse{
		Expression: &calculatorv1.Expression{
			Id:     "",
			Status: 0,
			Result: nil,
		},
	}, nil
}
