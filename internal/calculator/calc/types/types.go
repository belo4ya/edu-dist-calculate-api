package types

import "errors"

var ErrInvalidExpr = errors.New("invalid expression")

type Token struct {
	IsNumber bool
	Number   float64
	Symbol   string
}

type Task struct {
	ID            string
	ParentTask1ID string
	ParentTask2ID string

	Arg1      float64
	Arg2      float64
	Operation string
}
