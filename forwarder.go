package jsonlotelforwarder

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mashiike/go-otlp-helper/otlp"
	logspb "go.opentelemetry.io/proto/otlp/logs/v1"
	metricspb "go.opentelemetry.io/proto/otlp/metrics/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
)

type Forwarder struct {
	options *Options
}

func New(options *Options) (*Forwarder, error) {
	if err := options.Validate(); err != nil {
		return nil, fmt.Errorf("validate options: %w", err)
	}
	return &Forwarder{
		options: options,
	}, nil
}

func (f *Forwarder) Run(ctx context.Context) {
	if strings.HasPrefix(os.Getenv("AWS_EXECUTION_ENV"), "AWS_Lambda") || os.Getenv("AWS_LAMBDA_RUNTIME_API") != "" {
		lambda.Start(f.Invoke)
		return
	}
	dec := json.NewDecoder(os.Stdin)
	for dec.More() {
		var payload json.RawMessage
		if err := dec.Decode(&payload); err != nil {
			slog.Error("failed to decode payload", "error", err)
			os.Exit(1)
		}
		if _, err := f.Invoke(ctx, payload); err != nil {
			slog.Error("failed to invoke", "error", err)
			os.Exit(1)
		}
	}
}

func (f *Forwarder) Invoke(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	results, ok := Parse(payload)
	if !ok {
		return json.RawMessage(`{"skip":true}`), nil
	}
	return f.invokeAsExportTelemetry(ctx, results)
}

func (f *Forwarder) invokeAsExportTelemetry(ctx context.Context, results []*PaseResult) (json.RawMessage, error) {
	opts := f.options.clientOptions
	opts = append(opts, otlp.WithLogger(slog.Default()))
	client, err := otlp.NewClient("http://localhost:4317", opts...)
	if err != nil {
		return nil, fmt.Errorf("create otlp client: %w", err)
	}
	slog.InfoContext(ctx, "start otlp client", "endpoint", "http://localhost:4317")
	if err := client.Start(ctx); err != nil {
		return nil, fmt.Errorf("start otlp client: %w", err)
	}
	for _, result := range results {
		if result.Skip() {
			continue
		}
		if err := f.exportResult(ctx, client, result); err != nil {
			slog.ErrorContext(ctx, "failed to export telemetry", "error", err)
			continue
		}
	}
	slog.InfoContext(ctx, "stop otlp client")
	if err := client.Stop(ctx); err != nil {
		return nil, fmt.Errorf("stop trace client: %w", err)
	}
	return json.RawMessage(`{"success":true}`), nil
}

func (f *Forwarder) exportResult(ctx context.Context, client *otlp.Client, result *PaseResult) error {
	if result.Traces != nil && f.options.EnableTraces() {
		recourceSpans := result.Traces.GetResourceSpans()
		slog.InfoContext(ctx, "upload traces", "resource_spans", len(recourceSpans), "trace_ids", distinctListTraceIDs(recourceSpans))
		if err := client.UploadTraces(ctx, recourceSpans); err != nil {
			return fmt.Errorf("upload traces: %w", err)
		}
		slog.DebugContext(ctx, "uploaded traces", "resource_spans", len(recourceSpans))
		return nil
	}
	if result.Metrics != nil && f.options.EnableMetrics() {
		resourceMetrics := result.Metrics.GetResourceMetrics()
		slog.InfoContext(ctx, "upload metrics", "resource_metrics", len(resourceMetrics))
		if err := client.UploadMetrics(ctx, resourceMetrics); err != nil {
			return fmt.Errorf("upload metrics: %w", err)
		}
		slog.DebugContext(ctx, "uploaded metrics", "resource_metrics", len(resourceMetrics))
		return nil
	}
	if result.Logs != nil && f.options.EnableLogs() {
		resourceLogs := result.Logs.GetResourceLogs()
		slog.InfoContext(ctx, "upload logs", "resource_logs", len(resourceLogs))
		if err := client.UploadLogs(ctx, resourceLogs); err != nil {
			return fmt.Errorf("upload logs: %w", err)
		}
		slog.DebugContext(ctx, "uploaded logs", "resource_logs", len(resourceLogs))
		return nil
	}
	return nil
}

func distinctListTraceIDs(resourceSpans []*tracepb.ResourceSpans) []string {
	traceIDs := make(map[string]struct{})
	for _, resourceSpan := range resourceSpans {
		for _, scopeSpan := range resourceSpan.GetScopeSpans() {
			for _, span := range scopeSpan.GetSpans() {
				traceID := span.GetTraceId()
				traceIDStr := base64.StdEncoding.EncodeToString(traceID)
				if len(traceID) != 16 {
					slog.Warn("invalid trace id length", "trace_id", traceIDStr, "length", len(traceID))
				}
				traceIDs[traceIDStr] = struct{}{}
			}
		}
	}
	keys := make([]string, 0, len(traceIDs))
	for key := range traceIDs {
		keys = append(keys, key)
	}
	return keys
}

type PaseResult struct {
	Traces  *tracepb.TracesData
	Metrics *metricspb.MetricsData
	Logs    *logspb.LogsData
}

func (r *PaseResult) Skip() bool {
	return r == nil || (r.Traces == nil && r.Metrics == nil && r.Logs == nil)
}

func Parse(data []byte) ([]*PaseResult, bool) {
	if !json.Valid(data) {
		return nil, false
	}
	var subscriptionFilter CloudWatchSubscriptionFilterEvent
	if err := json.Unmarshal(data, &subscriptionFilter); err == nil && subscriptionFilter.AWSLogs != nil {
		logEvents := subscriptionFilter.GetLogEvents()
		if len(logEvents) == 0 {
			return nil, false
		}
		var results []*PaseResult
		for _, logEvent := range logEvents {
			tempResults, ok := Parse([]byte(logEvent))
			if !ok {
				continue
			}
			for _, tempResult := range tempResults {
				if tempResult.Skip() {
					continue
				}
				results = append(results, tempResult)
			}
		}
		if len(results) == 0 {
			return nil, false
		}
		return results, true
	}
	var traces tracepb.TracesData
	if err := otlp.UnmarshalJSON(data, &traces); err == nil {
		return []*PaseResult{{Traces: &traces}}, true
	}
	var metrics metricspb.MetricsData
	if err := otlp.UnmarshalJSON(data, &metrics); err == nil {
		return []*PaseResult{{Metrics: &metrics}}, true
	}
	var logs logspb.LogsData
	if err := otlp.UnmarshalJSON(data, &logs); err == nil {
		return []*PaseResult{{Logs: &logs}}, true
	}
	return nil, false
}
