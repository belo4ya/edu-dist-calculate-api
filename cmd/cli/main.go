package main

import (
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

	err := calc.Calc("2 + 2 * 2 + 1", 1, repo)
	if err != nil {
		slog.Error("error", "error", err)
	}
}
