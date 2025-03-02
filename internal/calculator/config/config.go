package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	GRPCAddr int `env:"GRPC_ADDR"`
	HTTPAddr int `env:"HTTP_ADDR"`
	MGMTAddr int `env:"MGMT_ADDR"`

	TimeAdditionMs       int `env:"TIME_ADDITION_MS"`
	TimeSubtractionMs    int `env:"TIME_SUBTRACTION_MS"`
	TimeMultiplicationMs int `env:"TIME_MULTIPLICATIONS_MS"`
	TimeDivisionMs       int `env:"TIME_DIVISIONS_MS"`
}

func Load() (*Config, error) {
	conf, err := env.ParseAs[Config]()
	if err != nil {
		return nil, fmt.Errorf("env parse: %w", err)
	}
	return &conf, nil
}
