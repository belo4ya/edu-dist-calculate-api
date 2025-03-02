package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	CalculatorAddr           string        `env:"CALCULATOR_ADDR"`
	GrpcClientConnectTimeout time.Duration `env:"GRPC_CLIENT_CONNECT_TIMEOUT"`
	ComputingPower           int           `env:"COMPUTING_POWER"`
}

func Load() (*Config, error) {
	conf := &Config{
		CalculatorAddr:           ":50051",
		GrpcClientConnectTimeout: 10 * time.Second,
		ComputingPower:           10,
	}
	if err := env.Parse(&conf); err != nil {
		return nil, fmt.Errorf("env parse: %w", err)
	}
	return conf, nil
}
