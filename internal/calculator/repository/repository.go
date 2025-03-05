package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/belo4ya/edu-dist-calculate-api/internal/calculator/repository/modelv2"
	"github.com/dgraph-io/badger/v4"
	"github.com/rs/xid"
)

type Repository struct {
	db *badger.DB
}

func New(db *badger.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateExpression(_ context.Context, cmd modelv2.CreateExpressionCmd, tasksCmd []modelv2.CreateExpressionTaskCmd) (string, error) {
	now := time.Now().UTC()

	expr := modelv2.Expression{
		ID:         xid.New().String(),
		Expression: cmd.Expression,
		Status:     modelv2.ExpressionStatusPending,
		Result:     0,
		Error:      "",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	tasks := make([]modelv2.Task, 0, len(tasksCmd))
	for _, t := range tasksCmd {
		tasks = append(tasks, modelv2.Task{
			ID:            t.ID,
			ExpressionID:  expr.ID,
			ParentTask1ID: t.ParentTask1ID,
			ParentTask2ID: t.ParentTask2ID,
			Arg1:          t.Arg1,
			Arg2:          t.Arg2,
			Operation:     t.Operation,
			Status:        modelv2.TaskStatusPending,
			Result:        0,
			ExpireAt:      time.Time{},
			CreatedAt:     now,
			UpdatedAt:     now,
		})
	}

	err := r.db.Update(func(txn *badger.Txn) error {
		// Store expression data
		exprData, err := json.Marshal(expr)
		if err != nil {
			return err
		}
		err = txn.Set([]byte("expr:"+expr.ID), exprData)
		if err != nil {
			return err
		}

		// Add to expression list
		listData, err := json.Marshal(expr.ID)
		if err != nil {
			return err
		}
		err = txn.Set([]byte("expr:list:"+expr.ID), listData)
		if err != nil {
			return err
		}

		// Store tasks
		for _, task := range tasks {
			// Store individual task
			taskData, err := json.Marshal(task)
			if err != nil {
				return err
			}
			err = txn.Set([]byte("task:"+task.ID), taskData)
			if err != nil {
				return err
			}

			// Add to expression's task list
			exprTaskKey := []byte("expr:" + expr.ID + ":tasks:" + task.ID)
			err = txn.Set(exprTaskKey, []byte{1})
			if err != nil {
				return err
			}

			// Add to pending task queue if it's ready to be executed
			if task.ParentTask1ID == "" && task.ParentTask2ID == "" {
				taskQueueKey := []byte("task:queue:pending")
				err = txn.Set(append(taskQueueKey, []byte(":"+task.ID)...), []byte{1})
				if err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("db update: %w", err)
	}

	return expr.ID, nil
}

func (r *Repository) ListExpressions(_ context.Context) ([]modelv2.Expression, error) {
	exprs := make([]modelv2.Expression, 0)

	err := r.db.View(func(txn *badger.Txn) error {
		// Create an iterator with prefix for the expression list
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := []byte("expr:list:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			// Get expression ID from the list
			var exprID string
			err := it.Item().Value(func(val []byte) error {
				return json.Unmarshal(val, &exprID)
			})
			if err != nil {
				return err
			}

			// Fetch the actual expression data using its ID
			item, err := txn.Get([]byte("expr:" + exprID))
			if err != nil {
				return err
			}

			// Deserialize the expression
			var expr modelv2.Expression
			err = item.Value(func(val []byte) error {
				return json.Unmarshal(val, &expr)
			})
			if err != nil {
				return err
			}

			exprs = append(exprs, expr)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("db view: %w", err)
	}

	return exprs, nil
}

func (r *Repository) GetExpression(_ context.Context, id string) (modelv2.Expression, error) {
	var expr modelv2.Expression

	err := r.db.View(func(txn *badger.Txn) error {
		// Construct the key for the expression
		key := []byte("expr:" + id)

		// Try to get the item from the database
		item, err := txn.Get(key)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return fmt.Errorf("expression with id %s not found", id)
			}
			return err
		}

		// Deserialize the expression data
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &expr)
		})
	})

	if err != nil {
		return modelv2.Expression{}, fmt.Errorf("db view: %w", err)
	}

	return expr, nil
}

func (r *Repository) GetTask(ctx context.Context) (modelv2.Task, error) {
	//TODO implement me
	panic("implement me")
}

func (r *Repository) UpdateTask(ctx context.Context, cmd modelv2.UpdateTaskCmd) error {
	//TODO implement me
	panic("implement me")
}

//func mapTaskOperation(s string) modelv2.TaskOperation {
//	switch s {
//	case "+":
//		return modelv2.TaskOperationAddition
//	case "-":
//		return modelv2.TaskOperationSubtraction
//	case "*":
//		return modelv2.TaskOperationMultiplication
//	case "/":
//		return modelv2.TaskOperationDivision
//	default:
//		return ""
//	}
//}
