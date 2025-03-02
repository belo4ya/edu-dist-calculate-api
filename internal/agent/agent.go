package agent

import (
	"context"
	"log/slog"
	"math"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/belo4ya/edu-dist-calculate-api/internal/agent/client"
	"github.com/belo4ya/edu-dist-calculate-api/internal/agent/config"
	calculatorv1 "github.com/belo4ya/edu-dist-calculate-api/pkg/calculator/v1"
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

	backoff := initialBackoff
	jitter := func() time.Duration {
		return time.Duration(rand.Int63n(int64(100 * time.Millisecond)))
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Получение задачи от оркестратора
			resp, err := a.client.GetTask(ctx, nil)
			if err != nil {
				log.WarnContext(ctx, "ошибка при запросе задачи", "error", err)
				time.Sleep(backoff + jitter())
				backoff = minDuration(backoff*2, maxBackoff)
				continue
			}

			// Сбрасываем backoff при наличии задачи
			backoff = initialBackoff

			// Обработка полученной задачи
			task := resp.GetTask()

			log = log.With("task_id", task.Id)
			log.InfoContext(ctx, "получена задача")

			// Вычисление результата
			result := a.executeTask(task)

			// Отправка результата оркестратору
			_, err = a.client.SubmitTaskResult(ctx, &calculatorv1.SubmitTaskResultRequest{
				Id:     task.Id,
				Result: result,
			})

			if err != nil {
				log.ErrorContext(ctx, "ошибка при отправке результата задачи", "error", err)
				continue
			}

			log.InfoContext(ctx, "задача выполнена успешно", "result", result)
		}
	}
}

const (
	initialBackoff = 100 * time.Millisecond
	maxBackoff     = 5 * time.Second
)

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
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
