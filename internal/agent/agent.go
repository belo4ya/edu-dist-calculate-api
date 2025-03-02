package agent

import (
	"context"

	"github.com/belo4ya/edu-dist-calculate-api/internal/agent/client"
	"github.com/belo4ya/edu-dist-calculate-api/internal/agent/config"
)

type Agent struct {
	conf   *config.Config
	client *client.CalculatorClient
}

func New(conf *config.Config, c *client.CalculatorClient) *Agent {
	return &Agent{conf: conf, client: c}
}

func (m *Agent) Start(ctx context.Context) error {
	return nil
}
