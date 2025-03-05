package modelv2

type CreateExpressionCmd struct {
	Expression string
}

type CreateExpressionTaskCmd struct {
	ID            string
	ParentTask1ID string
	ParentTask2ID string

	Arg1      float64
	Arg2      float64
	Operation TaskOperation
}

type UpdateTaskCmd struct {
	ID     string     `json:"id"`
	Status TaskStatus `json:"status"`
	Result float64    `json:"result"`
}
