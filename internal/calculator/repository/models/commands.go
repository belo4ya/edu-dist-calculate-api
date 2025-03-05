package models

import "time"

type CreateExpressionCmd struct {
	Expression string
}

type CreateExpressionTaskCmd struct {
	ID            string
	ParentTask1ID string
	ParentTask2ID string

	Arg1          float64
	Arg2          float64
	Operation     TaskOperation
	OperationTime time.Duration
}

type FinishTaskCmd struct {
	ID     string
	Status TaskStatus
	Result float64
}
