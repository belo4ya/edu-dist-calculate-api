package agent

import (
	"github.com/belo4ya/edu-dist-calculate-api/internal/agent/config"
)

type Agent struct {
	conf *config.Config
}

func New(conf *config.Config) *Agent {
	return &Agent{conf: conf}
}
