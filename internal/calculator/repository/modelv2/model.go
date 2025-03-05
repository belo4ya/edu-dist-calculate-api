package modelv2

import "time"

type Expression struct {
	ID         string           `json:"id"`
	Expression string           `json:"expression"`
	Status     ExpressionStatus `json:"status"`
	Result     float64          `json:"result"`
	Error      string           `json:"error"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ExpressionStatus string

const (
	ExpressionStatusPending    ExpressionStatus = "Pending"
	ExpressionStatusInProgress ExpressionStatus = "InProgress"
	ExpressionStatusCompleted  ExpressionStatus = "Completed"
	ExpressionStatusFailed     ExpressionStatus = "Failed"
)

type Task struct {
	ID            string `json:"id"`
	ExpressionID  string `json:"expression_id"`
	ParentTask1ID string `json:"parent_task_1_id"`
	ParentTask2ID string `json:"parent_task_2_id"`

	Arg1      float64       `json:"arg_1"`
	Arg2      float64       `json:"arg_2"`
	Operation TaskOperation `json:"operation"`
	Status    TaskStatus    `json:"status"`
	Result    float64       `json:"result"`
	ExpireAt  time.Time     `json:"expire_at"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TaskOperation string

const (
	TaskOperationAddition       TaskOperation = "+"
	TaskOperationSubtraction    TaskOperation = "-"
	TaskOperationMultiplication TaskOperation = "*"
	TaskOperationDivision       TaskOperation = "/"
)

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "Pending"
	TaskStatusInProgress TaskStatus = "InProgress"
	TaskStatusCompleted  TaskStatus = "Completed"
	TaskStatusFailed     TaskStatus = "Failed"
)
