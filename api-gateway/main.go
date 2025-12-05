package main

import (
	"os"
	"time"
	"flag"
	"syscall"
	"context"
	"net/http"
	"os/signal"
	_ "net/http/pprof"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"gateway/internal/routers"
	gc "gateway/internal/graceful"
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

	logFile, err := os.OpenFile("logs/all.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic("Failed to open all.log: " + err.Error())
	}
	errFile, err := os.OpenFile("logs/error.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
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

	go func(){
		log.Debug("Starting pprof server")
		http.Handle("/debug/pprof", http.DefaultServeMux)
		log.Error("Pprof server failed",
			zap.Error(http.ListenAndServe(":"+os.Getenv("PPROF_PORT"), nil)))
	}()

	srv := routers.NewServer(log)
	go func(){
		log.Debug("Server starting on " + srv.Srv.Addr)
		if err := srv.Srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("HTTP server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-quit
	log.Warn("Shutdown signal received")
	gracefulShutdown(srv, log)
}

func gracefulShutdown(srv *routers.Server, log *zap.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Info("Shutting down HTTP server")
	if err := gc.Shutdown(srv.Srv.Close, ctx); err != nil {
		log.Error("HTTP server shutdown error", zap.Error(err))
	}

	log.Info("Shutting down services")
	if err := srv.ShutdownServices(ctx); err != nil {
		log.Error("Services shutdown error", zap.Error(err))
	}
}
