package jsonlotelforwarder

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/fujiwara/lamblocal"
	logspb "go.opentelemetry.io/proto/otlp/logs/v1"
	metricspb "go.opentelemetry.io/proto/otlp/metrics/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/protobuf/encoding/protojson"
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
	lamblocal.Run(ctx, f.Invoke)
}

func (f *Forwarder) Invoke(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	results, ok := Parse(payload)
	if !ok {
		return json.RawMessage(`{"skip":true}`), nil
	}
	return f.invokeAsExportTelemetry(ctx, results)
}

func (f *Forwarder) invokeAsExportTelemetry(ctx context.Context, results []*PaseResult) (json.RawMessage, error) {
	client := NewClient(f.options)
	if err := client.Start(ctx); err != nil {
		return nil, fmt.Errorf("start trace client: %w", err)
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
	if err := client.Stop(ctx); err != nil {
		return nil, fmt.Errorf("stop trace client: %w", err)
	}
	return json.RawMessage(`{"success":true}`), nil
}

func (f *Forwarder) exportResult(ctx context.Context, client *Client, result *PaseResult) error {
	if result.Traces != nil && f.options.EnableTraces() {
		if err := client.UploadTraces(ctx, result.Traces.ResourceSpans); err != nil {
			return fmt.Errorf("upload traces: %w", err)
		}
		slog.InfoContext(ctx, "upload traces", "resource_spans", len(result.Traces.ResourceSpans))
		return nil
	}
	if result.Metrics != nil && f.options.EnableMetrics() {
		if err := client.UploadMetrics(ctx, result.Metrics.ResourceMetrics); err != nil {
			return fmt.Errorf("upload metrics: %w", err)
		}
		slog.InfoContext(ctx, "upload metrics", "resource_metrics", len(result.Metrics.ResourceMetrics))
		return nil
	}
	if result.Logs != nil && f.options.EnableLogs() {
		if err := client.UploadLogs(ctx, result.Logs.ResourceLogs); err != nil {
			return fmt.Errorf("upload logs: %w", err)
		}
		slog.InfoContext(ctx, "upload logs", "resource_logs", len(result.Logs.ResourceLogs))
		return nil
	}
	return nil
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
	if err := protojson.Unmarshal(data, &traces); err == nil {
		return []*PaseResult{{Traces: &traces}}, true
	}
	var metrics metricspb.MetricsData
	if err := protojson.Unmarshal(data, &metrics); err == nil {
		return []*PaseResult{{Metrics: &metrics}}, true
	}
	var logs logspb.LogsData
	if err := protojson.Unmarshal(data, &logs); err == nil {
		return []*PaseResult{{Logs: &logs}}, true
	}
	return nil, false
}
