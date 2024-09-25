package jsonlotelforwarder_test

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	jsonlotelforwarder "github.com/mashiike/jsonl-otel-forwarder"
	"github.com/stretchr/testify/require"
)

func EncodeLogEvents(t *testing.T, payload jsonlotelforwarder.CloudWatchLogsData) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	enc := json.NewEncoder(w)
	err := enc.Encode(payload)
	require.NoError(t, err)
	err = w.Close()
	require.NoError(t, err)
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	event := jsonlotelforwarder.CloudWatchSubscriptionFilterEvent{
		AWSLogs: &jsonlotelforwarder.CloudWatchSubscriptionFilterEventAWSLogs{
			Data: encoded,
		},
	}
	bs, err := json.Marshal(event)
	require.NoError(t, err)
	return bs
}

func EncodeSubscriptionFilterEvent(t *testing.T, logRecords [][]byte) []byte {
	t.Helper()
	payload := jsonlotelforwarder.CloudWatchLogsData{
		Owner:       "123456789012",
		LogGroup:    "test-log-group",
		LogStream:   "test-log-stream",
		MessageType: "DATA_MESSAGE",
		LogEvents:   make([]jsonlotelforwarder.CloudWatchLogsLogEvent, 0, len(logRecords)),
	}
	for i, record := range logRecords {
		payload.LogEvents = append(payload.LogEvents, jsonlotelforwarder.CloudWatchLogsLogEvent{
			ID:        fmt.Sprintf("eventId-%d", i),
			Timestamp: 1440442987000,
			Message:   string(record),
		})
	}
	return EncodeLogEvents(t, payload)
}

func TestCloudWatchSubscriptionFilterEvent(t *testing.T) {
	bs, err := os.ReadFile("testdata/subscription_filter.json")
	require.NoError(t, err)
	var event jsonlotelforwarder.CloudWatchSubscriptionFilterEvent
	err = json.Unmarshal(bs, &event)
	require.NoError(t, err)
	acutal := event.GetLogEvents()
	require.ElementsMatch(t, []string{
		"{\"eventVersion\":\"1.03\",\"userIdentity\":{\"type\":\"Root\"}",
		"{\"eventVersion\":\"1.03\",\"userIdentity\":{\"type\":\"Root\"}",
		"{\"eventVersion\":\"1.03\",\"userIdentity\":{\"type\":\"Root\"}",
	}, acutal)
}
