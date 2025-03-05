package calc

import (
	"github.com/belo4ya/edu-dist-calculate-api/internal/calculator/calc/stackx"
	"github.com/rs/xid"
)

type Task struct {
	ID            string
	ParentTask1ID string
	ParentTask2ID string

	Arg1      float64
	Arg2      float64
	Operation string
}

func (c *Calculator) Schedule(rpn []Token) []Task {
	plan := make([]Task, 0, len(rpn))

	type stackItem struct {
		IsTask bool
		TaskID string
		Value  float64
	}
	stack := stackx.New[stackItem]()

	for _, token := range rpn {
		if token.IsNumber {
			stack.Push(stackItem{IsTask: false, Value: token.Number})
			continue
		}

		task := Task{ID: xid.New().String(), Operation: token.Symbol}

		right, left := stack.SafePop(), stack.SafePop()
		if left.IsTask {
			task.ParentTask1ID = left.TaskID
		} else {
			task.Arg1 = left.Value
		}
		if right.IsTask {
			task.ParentTask2ID = right.TaskID
		} else {
			task.Arg2 = right.Value
		}

		plan = append(plan, task)
		stack.Push(stackItem{IsTask: true, TaskID: task.ID})
	}

	return plan
}
