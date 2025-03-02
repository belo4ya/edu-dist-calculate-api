package agent

import (
	"context"

	"github.com/belo4ya/edu-dist-calculate-api/internal/agent/config"
)

type Manager struct {
	conf *config.Config
}

func NewManager(conf *config.Config) *Manager {
	return &Manager{conf: conf}
}

func (a *Manager) Start(ctx context.Context) error {
	return nil
}
