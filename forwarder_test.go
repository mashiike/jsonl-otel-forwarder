package jsonlotelforwarder_test

import (
	"context"
	"os"
	"testing"

	otlpmux "github.com/mashiike/go-otlp-helper/otlp"
	"github.com/mashiike/go-otlp-helper/otlp/otlptest"
	jsonlotelforwarder "github.com/mashiike/jsonl-otel-forwarder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestForwarder__Trace(t *testing.T) {
	mux := otlpmux.NewServerMux()
	var actual []byte
	mux.Trace().HandleFunc(func(ctx context.Context, request *otlpmux.TraceRequest) (*otlpmux.TraceResponse, error) {
		var err error
		actual, err = protojson.Marshal(request)
		assert.NoError(t, err)
		headers, ok := otlpmux.HeadersFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, "dummy", headers.Get("Api-Key"))
		return &otlpmux.TraceResponse{}, nil
	})
	server := otlptest.NewServer(mux)
	defer server.Close()
	expected, err := os.ReadFile("testdata/trace.json")
	require.NoError(t, err)

	opts := jsonlotelforwarder.DefaultOptions()
	t.Setenv("FORWARDER_OTLP_ENDPOINT", server.URL)
	t.Setenv("FORWARDER_OTLP_PROTOCOL", "grpc")
	t.Setenv("FORWARDER_OTLP_HEADERS", "Api-Key=dummy")

	forwarder, err := jsonlotelforwarder.New(opts)
	require.NoError(t, err)
	ctx := context.Background()
	_, err = forwarder.Invoke(ctx, expected)
	require.NoError(t, err)

	var data otlpmux.TraceRequest
	err = protojson.Unmarshal(expected, &data)
	require.NoError(t, err)
	expectedRemarshal, err := protojson.Marshal(&data)
	require.NoError(t, err)

	t.Log("actual:", string(actual))
	t.Log("expected:", string(expectedRemarshal))
	require.JSONEq(t, string(expectedRemarshal), string(actual))
}

func TestForwarder__Metrics(t *testing.T) {
	mux := otlpmux.NewServerMux()
	var actual []byte
	mux.Metrics().HandleFunc(func(ctx context.Context, request *otlpmux.MetricsRequest) (*otlpmux.MetricsResponse, error) {
		var err error
		actual, err = protojson.Marshal(request)
		assert.NoError(t, err)
		headers, ok := otlpmux.HeadersFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, "dummy", headers.Get("Api-Key"))
		return &otlpmux.MetricsResponse{}, nil
	})
	server := otlptest.NewServer(mux)
	defer server.Close()
	expected, err := os.ReadFile("testdata/metrics.json")
	require.NoError(t, err)

	opts := jsonlotelforwarder.DefaultOptions()
	t.Setenv("FORWARDER_OTLP_ENDPOINT", server.URL)
	t.Setenv("FORWARDER_OTLP_PROTOCOL", "grpc")
	t.Setenv("FORWARDER_OTLP_HEADERS", "Api-Key=dummy")

	forwarder, err := jsonlotelforwarder.New(opts)
	require.NoError(t, err)
	ctx := context.Background()
	_, err = forwarder.Invoke(ctx, expected)
	require.NoError(t, err)

	var data otlpmux.MetricsRequest
	err = protojson.Unmarshal(expected, &data)
	require.NoError(t, err)
	expectedRemarshal, err := protojson.Marshal(&data)
	require.NoError(t, err)

	t.Log("actual:", string(actual))
	t.Log("expected:", string(expectedRemarshal))
	require.JSONEq(t, string(expectedRemarshal), string(actual))
}

func TestForwarder__Logs(t *testing.T) {
	mux := otlpmux.NewServerMux()
	var actual []byte
	mux.Logs().HandleFunc(func(ctx context.Context, request *otlpmux.LogsRequest) (*otlpmux.LogsResponse, error) {
		var err error
		actual, err = protojson.Marshal(request)
		assert.NoError(t, err)
		headers, ok := otlpmux.HeadersFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, "dummy", headers.Get("Api-Key"))
		return &otlpmux.LogsResponse{}, nil
	})
	server := otlptest.NewServer(mux)
	defer server.Close()
	expected, err := os.ReadFile("testdata/logs.json")
	require.NoError(t, err)

	opts := jsonlotelforwarder.DefaultOptions()
	t.Setenv("FORWARDER_OTLP_ENDPOINT", server.URL)
	t.Setenv("FORWARDER_OTLP_PROTOCOL", "grpc")
	t.Setenv("FORWARDER_OTLP_HEADERS", "Api-Key=dummy")

	forwarder, err := jsonlotelforwarder.New(opts)
	require.NoError(t, err)
	ctx := context.Background()
	_, err = forwarder.Invoke(ctx, expected)
	require.NoError(t, err)

	var data otlpmux.LogsRequest
	err = protojson.Unmarshal(expected, &data)
	require.NoError(t, err)
	expectedRemarshal, err := protojson.Marshal(&data)
	require.NoError(t, err)

	t.Log("actual:", string(actual))
	t.Log("expected:", string(expectedRemarshal))
	require.JSONEq(t, string(expectedRemarshal), string(actual))
}

func TestForworder__SubscriptionFilter__Trace(t *testing.T) {
	mux := otlpmux.NewServerMux()
	var actual []byte
	mux.Trace().HandleFunc(func(ctx context.Context, request *otlpmux.TraceRequest) (*otlpmux.TraceResponse, error) {
		var err error
		actual, err = protojson.Marshal(request)
		assert.NoError(t, err)
		headers, ok := otlpmux.HeadersFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, "dummy", headers.Get("Api-Key"))
		return &otlpmux.TraceResponse{}, nil
	})
	server := otlptest.NewServer(mux)
	defer server.Close()
	trace, err := os.ReadFile("testdata/trace.json")
	require.NoError(t, err)
	payload := EncodeSubscriptionFilterEvent(t, [][]byte{
		[]byte("test log event 1"),
		[]byte(`{"eventVersion":"1.03","userIdentity":{"type":"Root"}`),
		[]byte("test log event 2"),
		trace,
		[]byte("test log event 3"),
	})
	t.Setenv("FORWARDER_OTLP_ENDPOINT", server.URL)
	t.Setenv("FORWARDER_OTLP_PROTOCOL", "grpc")
	t.Setenv("FORWARDER_OTLP_HEADERS", "Api-Key=dummy")
	opts := jsonlotelforwarder.DefaultOptions()
	forwarder, err := jsonlotelforwarder.New(opts)
	require.NoError(t, err)
	ctx := context.Background()
	resp, err := forwarder.Invoke(ctx, payload)
	require.NoError(t, err)
	require.JSONEq(t, `{"success":true}`, string(resp))

	var data otlpmux.TraceRequest
	err = protojson.Unmarshal(trace, &data)
	require.NoError(t, err)
	expectedRemarshal, err := protojson.Marshal(&data)
	require.NoError(t, err)

	t.Log("actual:", string(actual))
	t.Log("expected:", string(expectedRemarshal))
	require.JSONEq(t, string(expectedRemarshal), string(actual))
}
