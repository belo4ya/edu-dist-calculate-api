package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	LogLevel     string `env:"LOG_LEVEL"`
	MgmtAddr     string `env:"MGMT_ADDR"`
	GRPCAddr     string `env:"GRPC_ADDR"`
	HTTPAddr     string `env:"HTTP_ADDR"`
	DBBadgerPath string `env:"DB_BADGER_PATH"`

	TimeAdditionMs       int `env:"TIME_ADDITION_MS"`
	TimeSubtractionMs    int `env:"TIME_SUBTRACTION_MS"`
	TimeMultiplicationMs int `env:"TIME_MULTIPLICATIONS_MS"`
	TimeDivisionMs       int `env:"TIME_DIVISIONS_MS"`
}

func Load() (*Config, error) {
	conf := &Config{
		LogLevel:             "info",
		MgmtAddr:             ":8081",
		GRPCAddr:             ":50051",
		HTTPAddr:             ":8080",
		DBBadgerPath:         ".data/badger",
		TimeAdditionMs:       1000,
		TimeSubtractionMs:    1000,
		TimeMultiplicationMs: 1000,
		TimeDivisionMs:       1000,
	}
	if err := env.Parse(conf); err != nil {
		return nil, fmt.Errorf("env parse: %w", err)
	}
	return conf, nil
}
