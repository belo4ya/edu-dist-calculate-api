package main

import (
	"fmt"
	"log/slog"

	"github.com/belo4ya/edu-dist-calculate-api/internal/calculator/calc"
)

func main() {
	s := "2 + 2 * 2 + (9 + 3)"

	c := &calc.Calculator{}
	rpn, err := c.Parse(s)
	if err != nil {
		slog.Error("error", "error", err)
	}
	fmt.Printf("%+v\n", rpn)
}
