package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	CalculatorAddr int `env:"CALCULATOR_ADDR"`
	ComputingPower int `env:"COMPUTING_POWER"`
}

func Load() (*Config, error) {
	conf, err := env.ParseAs[Config]()
	if err != nil {
		return nil, fmt.Errorf("env parse: %w", err)
	}
	return &conf, nil
}
