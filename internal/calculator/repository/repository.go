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

var emptyVal = []byte{1}

func (r *Repository) CreateExpression(
	_ context.Context,
	cmd modelv2.CreateExpressionCmd,
	tasksCmd []modelv2.CreateExpressionTaskCmd,
) (string, error) {
	now := time.Now().UTC()

	expr := modelv2.Expression{
		ID:         xid.New().String(),
		Expression: cmd.Expression,
		Status:     modelv2.ExpressionStatusPending,
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
		if err = txn.Set(exprKey(expr.ID), exprData); err != nil {
			return err
		}

		// Add to expression list
		if err = txn.Set(exprListKey(expr.ID), []byte(expr.ID)); err != nil {
			return err
		}

		// Store tasks
		for _, task := range tasks {
			// Store individual task
			taskData, err := json.Marshal(task)
			if err != nil {
				return err
			}
			if err := txn.Set(taskKey(task.ID), taskData); err != nil {
				return err
			}

			// Add to expression's task list
			if err = txn.Set(exprTaskKey(expr.ID, task.ID), emptyVal); err != nil {
				return err
			}

			// Add to pending task queue if it's ready to be executed
			if task.ParentTask1ID == "" && task.ParentTask2ID == "" {
				if err = txn.Set(taskQueuePendingKey(task.ID), emptyVal); err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	return expr.ID, nil
}

func (r *Repository) ListExpressions(_ context.Context) ([]modelv2.Expression, error) {
	var exprs []modelv2.Expression

	err := r.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := exprListPrefix()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			// Get expression ID from the list
			var exprID string
			_ = it.Item().Value(func(val []byte) error {
				exprID = string(val)
				return nil
			})

			// Fetch the actual expression data using its ID
			item, err := txn.Get(exprKey(exprID))
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
		return nil, err
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

func (r *Repository) GetPendingTask(_ context.Context) (modelv2.Task, error) {
	var task modelv2.Task
	var taskID string

	err := r.db.Update(func(txn *badger.Txn) error {
		// Look for a task in the pending queue
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := []byte("task:queue:pending:")
		it.Seek(prefix)

		// Check if we have any pending task
		if !it.ValidForPrefix(prefix) {
			return fmt.Errorf("no pending tasks available")
		}

		// Get the first pending task ID
		taskID = string(it.Item().Key())[len(prefix):]

		// Remove from pending queue
		if err := txn.Delete(it.Item().Key()); err != nil {
			return fmt.Errorf("failed to remove task from pending queue: %w", err)
		}

		// Get the task data
		taskItem, err := txn.Get([]byte("task:" + taskID))
		if err != nil {
			return fmt.Errorf("task with id %s not found: %w", taskID, err)
		}

		// Deserialize the task
		err = taskItem.Value(func(val []byte) error {
			return json.Unmarshal(val, &task)
		})
		if err != nil {
			return fmt.Errorf("failed to unmarshal task: %w", err)
		}

		// Set the task status to in-progress and update timestamp
		task.Status = modelv2.TaskStatusInProgress
		task.UpdatedAt = time.Now().UTC()
		task.ExpireAt = time.Now().UTC().Add(1 * time.Hour) // Set expiration time for the task

		// Update the task in the store
		updatedTaskData, err := json.Marshal(task)
		if err != nil {
			return fmt.Errorf("failed to marshal updated task: %w", err)
		}

		if err := txn.Set([]byte("task:"+taskID), updatedTaskData); err != nil {
			return fmt.Errorf("failed to update task status: %w", err)
		}

		// Update expression status to in-progress if it was pending
		exprKey := []byte("expr:" + task.ExpressionID)
		exprItem, err := txn.Get(exprKey)
		if err != nil {
			return fmt.Errorf("failed to get expression: %w", err)
		}

		var expr modelv2.Expression
		err = exprItem.Value(func(val []byte) error {
			return json.Unmarshal(val, &expr)
		})
		if err != nil {
			return fmt.Errorf("failed to unmarshal expression: %w", err)
		}

		if expr.Status == modelv2.ExpressionStatusPending {
			expr.Status = modelv2.ExpressionStatusInProgress
			expr.UpdatedAt = time.Now().UTC()

			exprData, err := json.Marshal(expr)
			if err != nil {
				return fmt.Errorf("failed to marshal updated expression: %w", err)
			}

			if err := txn.Set(exprKey, exprData); err != nil {
				return fmt.Errorf("failed to update expression status: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return modelv2.Task{}, err
	}

	return task, nil
}

func (r *Repository) FinishTask(_ context.Context, cmd modelv2.UpdateTaskCmd) error {
	return r.db.Update(func(txn *badger.Txn) error {
		// First, retrieve the existing task
		taskKey := []byte("task:" + cmd.ID)
		item, err := txn.Get(taskKey)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return fmt.Errorf("task with id %s not found", cmd.ID)
			}
			return err
		}

		var task modelv2.Task
		err = item.Value(func(val []byte) error {
			return json.Unmarshal(val, &task)
		})
		if err != nil {
			return err
		}

		// Update task fields
		task.Status = cmd.Status
		task.Result = cmd.Result
		task.UpdatedAt = time.Now().UTC()

		// If the task is completed, check if we need to mark it as completed in the queue
		if cmd.Status == modelv2.TaskStatusCompleted || cmd.Status == modelv2.TaskStatusFailed {
			// Remove from pending queue if it was there
			pendingQueueKey := []byte("task:queue:pending:" + cmd.ID)
			_ = txn.Delete(pendingQueueKey) // Ignore error if key doesn't exist

			if cmd.Status == modelv2.TaskStatusCompleted {
				// If this task is completed, check if any dependent tasks can now be queued
				if err := r.queueDependentTasks(txn, task); err != nil {
					return err
				}
			} else if cmd.Status == modelv2.TaskStatusFailed {
				// If this task is failed, fail fast all expression tasks
				if err := r.failDependentTasks(txn, task); err != nil {
					return err
				}
			}

			// Check if all tasks for this expression are completed
			// and update expression status if necessary
			if err := r.checkExpressionCompletion(txn, task.ExpressionID); err != nil {
				return err
			}
		}

		// Save updated task
		updatedTaskData, err := json.Marshal(task)
		if err != nil {
			return err
		}
		return txn.Set(taskKey, updatedTaskData)
	})
}

// Helper function to check if all tasks for an expression are completed
func (r *Repository) checkExpressionCompletion(txn *badger.Txn, exprID string) error {
	// Get the expression
	exprKey := []byte("expr:" + exprID)
	item, err := txn.Get(exprKey)
	if err != nil {
		return err
	}

	var expr modelv2.Expression
	err = item.Value(func(val []byte) error {
		return json.Unmarshal(val, &expr)
	})
	if err != nil {
		return err
	}

	// Check if all tasks are completed
	allCompleted := true
	anyFailed := false

	// Iterate through all tasks for this expression
	prefix := []byte("expr:" + exprID + ":tasks:")
	it := txn.NewIterator(badger.DefaultIteratorOptions)
	defer it.Close()

	for it.Seek(prefix); it.ValidForPrefix(prefix) && allCompleted; it.Next() {
		taskID := string(it.Item().Key())[len(prefix):]
		taskItem, err := txn.Get([]byte("task:" + taskID))
		if err != nil {
			return err
		}

		var task modelv2.Task
		err = taskItem.Value(func(val []byte) error {
			return json.Unmarshal(val, &task)
		})
		if err != nil {
			return err
		}

		if task.Status == modelv2.TaskStatusFailed {
			anyFailed = true
			allCompleted = false
			break
		}

		if task.Status != modelv2.TaskStatusCompleted {
			allCompleted = false
		}
	}

	// Update expression status if all tasks are completed or any failed
	if anyFailed {
		expr.Status = modelv2.ExpressionStatusFailed
		expr.UpdatedAt = time.Now().UTC()
	} else if allCompleted {
		// Find the final task (the one with no dependent tasks)
		var finalTask modelv2.Task
		found := false

		it.Rewind()
		for it.Seek(prefix); it.ValidForPrefix(prefix) && !found; it.Next() {
			taskID := string(it.Item().Key())[len(prefix):]
			taskItem, err := txn.Get([]byte("task:" + taskID))
			if err != nil {
				return err
			}

			var task modelv2.Task
			err = taskItem.Value(func(val []byte) error {
				return json.Unmarshal(val, &task)
			})
			if err != nil {
				return err
			}

			// Check if this task is the final one (not used as a parent by any other task)
			isDependency := false
			it2 := txn.NewIterator(badger.DefaultIteratorOptions)
			defer it2.Close()

			for it2.Seek(prefix); it2.ValidForPrefix(prefix) && !isDependency; it2.Next() {
				var checkTask modelv2.Task
				err = it2.Item().Value(func(val []byte) error {
					return json.Unmarshal(val, &checkTask)
				})
				if err != nil {
					return err
				}

				if checkTask.ParentTask1ID == task.ID || checkTask.ParentTask2ID == task.ID {
					isDependency = true
				}
			}

			if !isDependency {
				finalTask = task
				found = true
			}
		}

		if found {
			expr.Status = modelv2.ExpressionStatusCompleted
			expr.Result = finalTask.Result
			expr.UpdatedAt = time.Now().UTC()
		}
	}

	// Save updated expression
	exprData, err := json.Marshal(expr)
	if err != nil {
		return err
	}

	return txn.Set(exprKey, exprData)
}

// Helper function to queue dependent tasks that are now ready to execute
func (r *Repository) queueDependentTasks(txn *badger.Txn, completedTask modelv2.Task) error {
	// Get all tasks for this expression
	prefix := []byte("expr:" + completedTask.ExpressionID + ":tasks:")
	it := txn.NewIterator(badger.DefaultIteratorOptions)
	defer it.Close()

	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		taskID := string(it.Item().Key())[len(prefix):]
		taskItem, err := txn.Get([]byte("task:" + taskID))
		if err != nil {
			continue // Skip if we can't get this task
		}

		var task modelv2.Task
		err = taskItem.Value(func(val []byte) error {
			return json.Unmarshal(val, &task)
		})
		if err != nil {
			continue // Skip if we can't unmarshal
		}

		// Check if this task depends on the completed task
		if task.ParentTask1ID == completedTask.ID || task.ParentTask2ID == completedTask.ID {
			// Check if all dependencies are now satisfied
			dependenciesMet := true

			if task.ParentTask1ID != "" {
				parent1Item, err := txn.Get([]byte("task:" + task.ParentTask1ID))
				if err != nil {
					dependenciesMet = false
				} else {
					var parent1 modelv2.Task
					err = parent1Item.Value(func(val []byte) error {
						return json.Unmarshal(val, &parent1)
					})
					if err != nil || parent1.Status != modelv2.TaskStatusCompleted {
						dependenciesMet = false
					}
				}
			}

			if dependenciesMet && task.ParentTask2ID != "" {
				parent2Item, err := txn.Get([]byte("task:" + task.ParentTask2ID))
				if err != nil {
					dependenciesMet = false
				} else {
					var parent2 modelv2.Task
					err = parent2Item.Value(func(val []byte) error {
						return json.Unmarshal(val, &parent2)
					})
					if err != nil || parent2.Status != modelv2.TaskStatusCompleted {
						dependenciesMet = false
					}
				}
			}

			// If all dependencies are satisfied, update the task's arguments with parent results
			// and queue it for execution
			if dependenciesMet {
				// Set arguments from parent tasks if they exist
				if task.ParentTask1ID != "" {
					parent1Item, _ := txn.Get([]byte("task:" + task.ParentTask1ID))
					var parent1 modelv2.Task
					_ = parent1Item.Value(func(val []byte) error {
						return json.Unmarshal(val, &parent1)
					})
					task.Arg1 = parent1.Result
				}

				if task.ParentTask2ID != "" {
					parent2Item, _ := txn.Get([]byte("task:" + task.ParentTask2ID))
					var parent2 modelv2.Task
					_ = parent2Item.Value(func(val []byte) error {
						return json.Unmarshal(val, &parent2)
					})
					task.Arg2 = parent2.Result
				}

				// Update the task
				updatedTaskData, err := json.Marshal(task)
				if err != nil {
					continue
				}
				if err := txn.Set([]byte("task:"+task.ID), updatedTaskData); err != nil {
					continue
				}

				// Add to pending queue
				if err := txn.Set([]byte("task:queue:pending:"+task.ID), emptyVal); err != nil {
					continue
				}
			}
		}
	}

	return nil
}

// TODO: If this task is failed, fail fast all expression tasks
func (r *Repository) failDependentTasks(txn *badger.Txn, failedTask modelv2.Task) error {
	// Get all tasks for this expression
	prefix := []byte("expr:" + failedTask.ExpressionID + ":tasks:")
	it := txn.NewIterator(badger.DefaultIteratorOptions)
	defer it.Close()

	// Store all tasks that depend on the failed task (directly or indirectly)
	// using a map for quick lookups
	failedTaskIDs := make(map[string]struct{})
	failedTaskIDs[failedTask.ID] = struct{}{}

	// Find all dependent tasks (recursively)
	// We need multiple passes to handle multi-level dependencies
	changed := true
	for changed {
		changed = false

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			taskID := string(it.Item().Key())[len(prefix):]

			// Skip tasks that are already marked as failed
			if _, exists := failedTaskIDs[taskID]; exists {
				continue
			}

			taskItem, err := txn.Get([]byte("task:" + taskID))
			if err != nil {
				continue // Skip if we can't get this task
			}

			var task modelv2.Task
			err = taskItem.Value(func(val []byte) error {
				return json.Unmarshal(val, &task)
			})
			if err != nil {
				continue // Skip if we can't unmarshal
			}

			// Check if this task depends on any failed task
			_, parent1Failed := failedTaskIDs[task.ParentTask1ID]
			_, parent2Failed := failedTaskIDs[task.ParentTask2ID]

			if parent1Failed || parent2Failed {
				failedTaskIDs[task.ID] = struct{}{}
				changed = true
			}
		}
	}

	// Mark all dependent tasks as failed and remove them from the pending queue
	for taskID := range failedTaskIDs {
		// Skip the original failed task as it's already marked as failed
		if taskID == failedTask.ID {
			continue
		}

		taskItem, err := txn.Get([]byte("task:" + taskID))
		if err != nil {
			continue // Skip if we can't get this task
		}

		var task modelv2.Task
		err = taskItem.Value(func(val []byte) error {
			return json.Unmarshal(val, &task)
		})
		if err != nil {
			continue
		}

		// Update task status to failed
		task.Status = modelv2.TaskStatusFailed
		task.UpdatedAt = time.Now().UTC()

		// Remove from pending queue if it was there
		pendingQueueKey := []byte("task:queue:pending:" + task.ID)
		_ = txn.Delete(pendingQueueKey) // Ignore error if key doesn't exist

		// Save updated task
		updatedTaskData, err := json.Marshal(task)
		if err != nil {
			continue
		}

		if err := txn.Set([]byte("task:"+task.ID), updatedTaskData); err != nil {
			continue
		}
	}

	// Update expression status to failed
	exprKey := []byte("expr:" + failedTask.ExpressionID)
	exprItem, err := txn.Get(exprKey)
	if err != nil {
		return err
	}

	var expr modelv2.Expression
	err = exprItem.Value(func(val []byte) error {
		return json.Unmarshal(val, &expr)
	})
	if err != nil {
		return err
	}

	expr.Status = modelv2.ExpressionStatusFailed
	expr.Error = "Task execution failed"
	expr.UpdatedAt = time.Now().UTC()

	exprData, err := json.Marshal(expr)
	if err != nil {
		return err
	}

	return txn.Set(exprKey, exprData)
}

func (r *Repository) ListExpressionTasks(_ context.Context, id string) ([]modelv2.Task, error) {
	var tasks []modelv2.Task

	err := r.db.View(func(txn *badger.Txn) error {
		// First check if the expression exists
		if _, err := txn.Get(exprKey(id)); err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return fmt.Errorf("expression with id %s not found", id)
			}
			return err
		}

		// Expression exists, now get all its tasks
		prefix := exprTasksPrefix(id)
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			// Get task ID from the key
			taskID := taskIDFromExprTaskKey(it.Item().Key(), id)

			// Get the actual task data
			taskItem, err := txn.Get(taskKey(taskID))
			if err != nil {
				// Log error but continue with other tasks
				continue
			}

			// Deserialize the task
			var task modelv2.Task
			err = taskItem.Value(func(val []byte) error {
				return json.Unmarshal(val, &task)
			})
			if err != nil {
				// Log error but continue with other tasks
				continue
			}

			tasks = append(tasks, task)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("list expression tasks: %w", err)
	}

	return tasks, nil
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
