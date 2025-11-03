package main

import (
	"flag"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"gateway/internal/routers"
)

func setupLog() *zap.Logger {
	var logLevel string
	flag.StringVar(&logLevel, "level", "debug", "set log level")
	flag.Parse()

	cfg := zap.NewDevelopmentConfig()
	switch logLevel {
	case "debug":
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "warn":
		cfg.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		cfg.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	}
	log, _ := cfg.Build()
	return log
}

func main() {
	log := setupLog()

	srv := routers.NewServer(log)
	log.Fatal("Fatal server failure", zap.Any("Error", srv.ListenAndServe()))
}
