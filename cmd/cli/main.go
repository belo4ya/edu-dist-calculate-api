package main

import (
	"fmt"
	"time"

	"github.com/belo4ya/edu-dist-calculate-api/internal/calculator/calc"
	calculatorv1 "github.com/belo4ya/edu-dist-calculate-api/pkg/calculator/v1"
)

func main() {
	parser := calc.NewExpressionParser(nil, map[calculatorv1.Operation]time.Duration{})
	expr, err := parser.ParseExpression("1 + 2 + 3")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", expr)
}
