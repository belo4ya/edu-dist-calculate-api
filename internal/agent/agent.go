package agent

import (
	"context"
	"log/slog"
	"math"
	"sync"
	"time"

	"github.com/belo4ya/edu-dist-calculate-api/internal/agent/config"
	calculatorv1 "github.com/belo4ya/edu-dist-calculate-api/pkg/calculator/v1"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
)

type CalculatorAgentAPIClient interface {
	GetTask(ctx context.Context) (*calculatorv1.Task, error)
	SubmitTaskResult(ctx context.Context, res *calculatorv1.SubmitTaskResultRequest) error
}

type Agent struct {
	conf   *config.Config
	client CalculatorAgentAPIClient
	log    *slog.Logger
}

func New(conf *config.Config, c CalculatorAgentAPIClient) *Agent {
	return &Agent{
		conf:   conf,
		client: c,
		log:    slog.With("logger", "agent"),
	}
}

func (a *Agent) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	for i := 0; i < a.conf.ComputingPower; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			a.worker(ctx, i)
		}()
	}

	wg.Wait()
	return nil
}

func (a *Agent) worker(ctx context.Context, workerID int) {
	log := a.log.With("worker_id", workerID)
	log.InfoContext(ctx, "worker started")

	for {
		select {
		case <-ctx.Done():
			log.InfoContext(ctx, "worker stopped")
			return
		default:
			task, err := a.fetchTask(ctx, log)
			if err != nil {
				continue
			}

			log = log.With("task_id", task.Id)
			log.DebugContext(ctx, "executing task")

			result := a.executeTask(task)

			if err := a.submitTaskResult(ctx, log, task.Id, result); err != nil {
				continue
			}

			log.DebugContext(ctx, "task completed", "result", result)
		}
	}
}

func (a *Agent) executeTask(task *calculatorv1.Task) float64 {
	time.Sleep(task.OperationTime.AsDuration())

	switch task.Operation {
	case calculatorv1.TaskOperation_TASK_OPERATION_ADDITION:
		return task.Arg1 + task.Arg2
	case calculatorv1.TaskOperation_TASK_OPERATION_SUBTRACTION:
		return task.Arg1 - task.Arg2
	case calculatorv1.TaskOperation_TASK_OPERATION_MULTIPLICATION:
		return task.Arg1 * task.Arg2
	case calculatorv1.TaskOperation_TASK_OPERATION_DIVISION:
		if task.Arg2 == 0 {
			return math.NaN()
		}
		return task.Arg1 / task.Arg2
	default:
		return math.NaN()
	}
}

func (a *Agent) fetchTask(ctx context.Context, log *slog.Logger) (*calculatorv1.Task, error) {
	backoff := backoffExponentialWithJitter(100*time.Millisecond, 5*time.Second, 0.2)

	for attempt := 1; ; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			task, err := a.client.GetTask(ctx)
			if err != nil {
				log.ErrorContext(ctx, "failed to fetch task", "error", err, "attempt", attempt)
				time.Sleep(backoff(attempt))
				continue
			}
			if task == nil {
				log.DebugContext(ctx, "no tasks available")
				time.Sleep(5 * time.Second)
				continue
			}

			return task, nil
		}
	}
}

func (a *Agent) submitTaskResult(ctx context.Context, log *slog.Logger, taskID string, result float64) error {
	backoff := backoffExponentialWithJitter(100*time.Millisecond, 5*time.Second, 0.2)

	req := &calculatorv1.SubmitTaskResultRequest{
		Id:     taskID,
		Result: result,
	}

	for attempt := 1; ; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := a.client.SubmitTaskResult(ctx, req); err != nil {
				log.ErrorContext(ctx, "failed to submit result", "error", err, "attempt", attempt)
				time.Sleep(backoff(attempt))
				continue
			}
			return nil
		}
	}
}

func backoffExponentialWithJitter(dur time.Duration, cap time.Duration, jitter float64) func(int) time.Duration {
	f := retry.BackoffExponentialWithJitter(dur, jitter)
	return func(attempt int) time.Duration {
		return min(f(context.Background(), uint(attempt)), cap)
	}
}
