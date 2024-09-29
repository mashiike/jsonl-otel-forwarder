package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	"github.com/handlename/ssmwrap/v2"
	"github.com/ken39arg/go-flagx"
	jsonlotelforwarder "github.com/mashiike/jsonl-otel-forwarder"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if paths := os.Getenv("SSMWRAP_NAMES"); paths != "" {
		rules := make([]ssmwrap.ExportRule, 0)
		for _, path := range strings.Split(paths, ",") {
			rules = append(rules, ssmwrap.ExportRule{
				Path: path,
			})
		}
		if err := ssmwrap.Export(ctx, rules, ssmwrap.ExportOptions{}); err != nil {
			fmt.Fprintf(os.Stderr, "failed to export parameters: %v", err)
			os.Exit(1)
		}
	}
	logLevel := "info"
	flag.StringVar(&logLevel, "log-level", logLevel, "log level ($FORWARDER_LOG_LEVEL)")
	opts := jsonlotelforwarder.DefaultOptions()
	opts.SetFlags(flag.CommandLine)
	flag.VisitAll(flagx.EnvToFlagWithPrefix("FORWARDER_"))
	flag.CommandLine.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Usage: jsonl-otel-forwarder [options]")
		fmt.Fprintln(flag.CommandLine.Output(), "")
		fmt.Fprintln(flag.CommandLine.Output(), "\t Forward JSON Lines OTLP logs to OpenTelemetry Collector/Server")
		fmt.Fprintln(flag.CommandLine.Output(), "")
		fmt.Fprintln(flag.CommandLine.Output(), "Options:")
		fmt.Fprintln(flag.CommandLine.Output(), "")
		flag.PrintDefaults()
	}
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
	slog.Info("start forwarder", "version", jsonlotelforwarder.Version, "log-level", minLevel)

	forwarder, err := jsonlotelforwarder.New(opts)
	if err != nil {
		slog.Error("failed to create forwarder", "error", err)
		os.Exit(1)
	}
	forwarder.Run(ctx)
}
