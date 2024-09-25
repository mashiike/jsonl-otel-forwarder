package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"

	jsonlotelforwarder "github.com/mashiike/jsonl-otel-forwarder"
)

func main() {
	opts := jsonlotelforwarder.DefaultOptions()
	opts.SetFlags(flag.CommandLine)
	var (
		logLevel string = "info"
	)
	if envLogLevel := os.Getenv(jsonlotelforwarder.EnvPrefix + "LOG_LEVEL"); envLogLevel != "" {
		logLevel = envLogLevel
	}
	flag.StringVar(&logLevel, "log-level", logLevel, "log level ($FORWARDER_LOG_LEVEL)")
	flag.Parse()

	var minLevel slog.Level
	var logLevelErr error
	if err := minLevel.UnmarshalText([]byte(logLevel)); err != nil {
		logLevelErr = err
		minLevel = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: minLevel,
	})))
	if logLevelErr != nil {
		slog.Warn("invalid log level, fallback to info level", "error", logLevelErr, "log-level", logLevel)
	}

	forwarder, err := jsonlotelforwarder.New(opts)
	if err != nil {
		slog.Error("failed to create forwarder", "error", err)
		os.Exit(1)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	forwarder.Run(ctx)
}
