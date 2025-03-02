package agent

import (
	"context"
	"log/slog"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/belo4ya/edu-dist-calculate-api/internal/agent/client"
	"github.com/belo4ya/edu-dist-calculate-api/internal/agent/config"
	calculatorv1 "github.com/belo4ya/edu-dist-calculate-api/pkg/calculator/v1"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
)

type Agent struct {
	conf   *config.Config
	client *client.CalculatorClient
	log    *slog.Logger
}

func New(conf *config.Config, c *client.CalculatorClient) *Agent {
	return &Agent{
		conf:   conf,
		client: c,
		log:    slog.Default().With("logger", "agent"),
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
	log.InfoContext(ctx, "запуск воркера")

	for {
		select {
		case <-ctx.Done():
			log.InfoContext(ctx, "завершение работы воркера")
			return
		default:
			task, err := a.fetchTask(ctx, log)
			if err != nil {
				continue
			}

			log = log.With("task_id", task.Id)
			log.InfoContext(ctx, "выполнение задачи")

			result := a.executeTask(task)

			if err := a.submitTaskResult(ctx, log, task.Id, result); err != nil {
				continue
			}

			log.InfoContext(ctx, "задача выполнена успешно", "result", result)
		}
	}
}

// fetchTask получает задачу от оркестратора с повторами при ошибках
func (a *Agent) fetchTask(ctx context.Context, log *slog.Logger) (*calculatorv1.Task, error) {
	backoff := backoffExponentialWithJitter(100*time.Millisecond, 5*time.Second, 0.2)

	for attempt := 1; ; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			task, err := a.client.GetTask(ctx)
			if err != nil || task == nil {
				log.WarnContext(ctx, "ошибка при запросе задачи", "error", err, "attempt", attempt)
				time.Sleep(backoff(attempt))
				continue
			}
			if task == nil {
				log.DebugContext(ctx, "нет доступных задач")
				time.Sleep(backoff(attempt))
				continue
			}

			return task, nil
		}
	}
}

// submitTaskResult отправляет результат задачи с повторами при ошибках
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
			if err := a.client.SubmitResult(ctx, req); err != nil {
				log.WarnContext(ctx, "ошибка при отправке результата задачи", "error", err, "attempt", attempt)
				time.Sleep(backoff(attempt))
			}
		}
	}
}

func backoffExponentialWithJitter(dur time.Duration, cap time.Duration, jitter float64) func(int) time.Duration {
	f := retry.BackoffExponentialWithJitter(dur, jitter)
	return func(attempt int) time.Duration {
		return min(f(context.Background(), uint(attempt)), cap)
	}
}

func (a *Agent) executeTask(task *calculatorv1.Task) float64 {
	arg1, err1 := strconv.ParseFloat(task.Arg1, 64)
	arg2, err2 := strconv.ParseFloat(task.Arg2, 64)
	if err1 != nil || err2 != nil {
		return math.NaN()
	}
	time.Sleep(task.OperationTime.AsDuration())
	return ops[task.Operation](arg1, arg2)
}

var ops = map[calculatorv1.Operation]func(float64, float64) float64{
	calculatorv1.Operation_OPERATION_ADDITION: func(a, b float64) float64 {
		return a + b
	},
	calculatorv1.Operation_OPERATION_SUBTRACTION: func(a, b float64) float64 {
		return a - b
	},
	calculatorv1.Operation_OPERATION_MULTIPLICATION: func(a, b float64) float64 {
		return a * b
	},
	calculatorv1.Operation_OPERATION_DIVISION: func(a, b float64) float64 {
		return a / b
	},
}
