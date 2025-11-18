package main

import (
	"os"
	"flag"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"gateway/internal/routers"
)

func setupLog() *zap.Logger {
	var logLevel string
	flag.StringVar(&logLevel, "level", "debug", "set log level")
	flag.Parse()

	var level zapcore.Level
	switch logLevel {
	case "debug": level = zapcore.DebugLevel
	case "warn": level = zapcore.WarnLevel
	case "error": level = zapcore.ErrorLevel
	default: level = zapcore.DebugLevel
	}
	if err := os.MkdirAll("logs", 0755); err != nil {
		panic("Failed to create logs directory: " + err.Error())
	}

	cfgEn := zap.NewDevelopmentEncoderConfig()
	fileEn := zapcore.NewJSONEncoder(cfgEn)
	consEn := zapcore.NewConsoleEncoder(cfgEn)

	logFile, err := os.OpenFile("logs/all.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic("Failed to open all.log: " + err.Error())
	}
	errFile, err := os.OpenFile("logs/error.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic("Failed to open error.log: " + err.Error())
	}

	allCore := zapcore.NewCore(fileEn, zapcore.AddSync(logFile), level)
	errCore := zapcore.NewCore(fileEn, zapcore.AddSync(errFile), zapcore.ErrorLevel)
	stdoutCore := zapcore.NewCore(consEn, zapcore.AddSync(os.Stdout), level)
	core := zapcore.NewTee(allCore, errCore, stdoutCore)

	log := zap.New(core, zap.AddCaller())
	return log
}

func main() {
	log := setupLog()
	defer log.Sync()

	srv := routers.NewServer(log)
	log.Debug("Server starting...")
	log.Fatal("Fatal server failure", zap.Any("Error", srv.ListenAndServe()))
}
