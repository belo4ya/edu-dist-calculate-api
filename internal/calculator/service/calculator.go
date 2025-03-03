package service

import (
	"context"
	"log/slog"

	"github.com/belo4ya/edu-dist-calculate-api/internal/calculator/config"
	calculatorv1 "github.com/belo4ya/edu-dist-calculate-api/pkg/calculator/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type CalculatorService struct {
	calculatorv1.UnimplementedCalculatorServiceServer
	conf *config.Config
	log  *slog.Logger
}

func NewCalculatorService(conf *config.Config) *CalculatorService {
	return &CalculatorService{
		conf: conf,
		log:  slog.With("logger", "calculator-svc"),
	}
}

func (s *CalculatorService) Register(srv *grpc.Server) {
	calculatorv1.RegisterCalculatorServiceServer(srv, s)
}

func (s *CalculatorService) RegisterGRPCGateway(ctx context.Context, mux *runtime.ServeMux, addr string, clientOpts []grpc.DialOption) error {
	return calculatorv1.RegisterCalculatorServiceHandlerFromEndpoint(ctx, mux, "localhost"+addr, clientOpts)
}

func (s *CalculatorService) Calculate(context.Context, *calculatorv1.CalculateRequest) (*calculatorv1.CalculateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Calculate not implemented")
}

func (s *CalculatorService) ListExpressions(context.Context, *emptypb.Empty) (*calculatorv1.ListExpressionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListExpressions not implemented")
}

func (s *CalculatorService) GetExpression(context.Context, *calculatorv1.GetExpressionRequest) (*calculatorv1.GetExpressionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetExpression not implemented")
}
