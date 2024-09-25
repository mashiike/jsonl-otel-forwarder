package jsonlotelforwarder

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"log/slog"
)

type CloudWatchSubscriptionFilterEvent struct {
	AWSLogs *CloudWatchSubscriptionFilterEventAWSLogs `json:"awslogs,omitempty"`
}

type CloudWatchSubscriptionFilterEventAWSLogs struct {
	Data string `json:"data,omitempty"`
}

type CloudWatchLogsData struct {
	Owner               string                   `json:"owner,omitempty"`
	LogGroup            string                   `json:"logGroup,omitempty"`
	LogStream           string                   `json:"logStream,omitempty"`
	SubscriptionFilters []string                 `json:"subscriptionFilters,omitempty"`
	MessageType         string                   `json:"messageType,omitempty"`
	LogEvents           []CloudWatchLogsLogEvent `json:"logEvents,omitempty"`
}

type CloudWatchLogsLogEvent struct {
	ID        string `json:"id,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
	Message   string `json:"message,omitempty"`
}

func (e *CloudWatchSubscriptionFilterEvent) GetLogEvents() []string {
	if e.AWSLogs == nil {
		return []string{}
	}
	data, err := base64.StdEncoding.DecodeString(e.AWSLogs.Data)
	if err != nil {
		slog.Warn("failed to decode base64", "error", err)
		return []string{}
	}
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		slog.Warn("failed to create gzip reader", "error", err)
		return []string{}
	}
	dec := json.NewDecoder(reader)
	var parsed CloudWatchLogsData
	if err := dec.Decode(&parsed); err != nil {
		slog.Warn("failed to decode json", "error", err)
		return []string{}
	}
	slog.Info("parsed log events", "owner", parsed.Owner, "logGroup", parsed.LogGroup, "logStream", parsed.LogStream, "messageType", parsed.MessageType, "subscriptionFilters", parsed.SubscriptionFilters, "logEvents", len(parsed.LogEvents))
	events := make([]string, 0, len(parsed.LogEvents))
	for _, event := range parsed.LogEvents {
		events = append(events, event.Message)
	}
	return events
}
