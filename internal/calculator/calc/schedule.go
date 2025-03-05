package calc

import (
	"strconv"

	"github.com/belo4ya/edu-dist-calculate-api/internal/calculator/calc/stackx"
	"github.com/belo4ya/edu-dist-calculate-api/internal/calculator/model"
	"github.com/rs/xid"
)

type Task struct {
	Arg1      float64
	Arg2      float64
	Operation string

	ID            string
	ParentTask1ID string
	ParentTask2ID string
}

func (c *Calculator) Schedule(rpn []model.Token) []Task {
	plan := make([]Task, 0, len(rpn))

	type stackItem struct {
		IsTask bool
		TaskID string
		Value  float64
	}
	stack := stackx.New[stackItem]()

	for _, token := range rpn {
		if token.IsNumber {
			value, _ := strconv.ParseFloat(token.Value, 64)
			stack.Push(stackItem{IsTask: false, Value: value})
			continue
		}

		task := Task{ID: xid.New().String(), Operation: token.Value}

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
