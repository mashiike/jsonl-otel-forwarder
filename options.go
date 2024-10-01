package jsonlotelforwarder

import (
	"flag"
	"os"
	"strconv"
	"strings"

	"github.com/mashiike/go-otlp-helper/otlp"
)

func DefaultOptions() *Options {
	signals := os.Getenv("FORWARDER_SIGNALS")
	if signals == "" {
		signals = "traces,metrics,logs"
	}
	return &Options{
		Signals: signals,
		clientOptions: []otlp.ClientOption{
			otlp.DefaultClientOptions("FORWARDER_", "OTEL_EXPORTER_"),
			otlp.WithUserAgent("jsonl-otel-forwarder/" + Version),
		},
	}
}

type Options struct {
	Signals       string
	Batch         bool
	clientOptions []otlp.ClientOption
}

func (o *Options) SetFlags(fs *flag.FlagSet) {
	o.clientOptions = append(
		o.clientOptions,
		otlp.ClientOptionsWithFlagSet(fs, "", "FORWARDER_", "OTEL_EXPORTER_"),
	)
	fs.StringVar(&o.Signals, "signals", o.Signals, "comma separated list of signals to forward [traces,metrics,logs] ($FORWARDER_SIGNALS)")
	fs.BoolVar(&o.Batch, "batch", toBool(os.Getenv("FOWARDER_BATCH")), "batch forward to export endpoint ($FORWARDER_BATCH)")
}

func toBool(s string) bool {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return false
	}
	return b
}

func (o *Options) SignalsList() []string {
	return strings.Split(o.Signals, ",")
}

func (o *Options) EnableTraces() bool {
	for _, signal := range o.SignalsList() {
		if strings.EqualFold(signal, "traces") {
			return true
		}
		if strings.EqualFold(signal, "trace") {
			return true
		}
	}
	return false
}

func (o *Options) EnableMetrics() bool {
	for _, signal := range o.SignalsList() {
		if strings.EqualFold(signal, "metrics") {
			return true
		}
		if strings.EqualFold(signal, "metric") {
			return true
		}
	}
	return false
}

func (o *Options) EnableLogs() bool {
	for _, signal := range o.SignalsList() {
		if strings.EqualFold(signal, "logs") {
			return true
		}
		if strings.EqualFold(signal, "log") {
			return true
		}
	}
	return false
}

func (o *Options) Validate() error {
	var err error
	_, err = otlp.NewClient("http://localhost:4317", o.clientOptions...)
	return err
}
