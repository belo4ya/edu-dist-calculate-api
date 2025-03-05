package repository

import (
	"context"
	"time"

	"github.com/belo4ya/edu-dist-calculate-api/internal/calculator/repository/modelv2"
	"github.com/cockroachdb/pebble"
	"github.com/rs/xid"
)

type Repository struct {
	db *pebble.DB
}

func New(db *pebble.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateExpression(ctx context.Context, expr modelv2.CreateExpressionCmd, tasks []modelv2.CreateExpressionTaskCmd) (string, error) {
	_ = modelv2.Expression{
		ID:         xid.New().String(),
		Expression: expr.Expression,
		Status:     modelv2.ExpressionStatusPending,
		Result:     0,
		Error:      "",
		CreatedAt:  time.Time{},
		UpdatedAt:  time.Time{},
	}
	//TODO implement me
	panic("implement me")

}

func (r *Repository) ListExpressions(ctx context.Context) ([]modelv2.Expression, error) {
	//TODO implement me
	panic("implement me")
}

func (r *Repository) GetExpression(ctx context.Context, s string) (modelv2.Expression, error) {
	//TODO implement me
	panic("implement me")
}
