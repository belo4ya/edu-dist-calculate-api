package main

import (
	"fmt"
	"log/slog"

	"github.com/belo4ya/edu-dist-calculate-api/internal/calculator/calc"
	"github.com/belo4ya/edu-dist-calculate-api/internal/calculator/model"
)

type Repo struct {
}

func (r *Repo) InsertTask(task *model.Task) (int, error) {
	slog.Info("insert task", "task", task)
	return 1, nil
}

func main() {
	repo := &Repo{}

	s := "2 + 2 * 2 + (9 + 3)"

	err := calc.Calc(s, 1, repo)
	if err != nil {
		slog.Error("error", "error", err)
	}

	c := &calc.Calculator{}
	rpn, err := c.Parse(s)
	if err != nil {
		slog.Error("error", "error", err)
	}
	fmt.Printf("%+v\n", rpn)
}
