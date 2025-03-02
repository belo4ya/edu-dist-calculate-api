package logging

import (
	"log/slog"
	"os"
)

type Config struct {
	Level string
}

func Configure(conf *Config) error {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(conf.Level)); err != nil {
		return err
	}
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     lvl,
	}))
	slog.SetDefault(log)
	return nil
}
