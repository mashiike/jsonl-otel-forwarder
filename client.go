package jsonlotelforwarder

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	colmetricpb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	logspb "go.opentelemetry.io/proto/otlp/logs/v1"
	metricspb "go.opentelemetry.io/proto/otlp/metrics/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Client struct {
	mu sync.RWMutex

	endpoint      string
	metadata      metadata.MD
	conn          *grpc.ClientConn
	dialOpts      []grpc.DialOption
	cnn           *grpc.ClientConn
	stopCtx       context.Context
	stopFunc      context.CancelFunc
	exportTimeout time.Duration
}

func NewClient(opts *Options) *Client {
	return &Client{
		endpoint:      opts.Endpoint(),
		metadata:      opts.headers(),
		dialOpts:      opts.DialOptions(),
		exportTimeout: opts.exportTimeoutDuration(),
	}
}

func (c *Client) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	conn, err := grpc.NewClient(c.endpoint, c.dialOpts...)
	if err != nil {
		return err
	}
	c.stopCtx, c.stopFunc = context.WithCancel(ctx)
	c.conn = conn
	return nil
}

func (c *Client) newContext(parent context.Context) (context.Context, context.CancelFunc) {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	if c.exportTimeout > 0 {
		ctx, cancel = context.WithTimeout(parent, c.exportTimeout)
	} else {
		ctx, cancel = context.WithCancel(parent)
	}

	if c.metadata.Len() > 0 {
		ctx = metadata.NewOutgoingContext(ctx, c.metadata)
	}

	go func() {
		select {
		case <-ctx.Done():
		case <-c.stopCtx.Done():
			cancel()
		}
	}()
	return ctx, cancel
}

var (
	ErrAlreadyClosed = errors.New("already closed")
	ErrNotStarted    = errors.New("not started")
)

func (c *Client) UploadTraces(ctx context.Context, protoSpans []*tracepb.ResourceSpans) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.conn == nil {
		return ErrNotStarted
	}

	sericeClient := coltracepb.NewTraceServiceClient(c.conn)
	ctx, cancel := c.newContext(ctx)
	defer cancel()

	resp, err := sericeClient.Export(ctx, &coltracepb.ExportTraceServiceRequest{
		ResourceSpans: protoSpans,
	})
	if resp != nil && resp.PartialSuccess != nil {
		msg := resp.PartialSuccess.GetErrorMessage()
		n := resp.PartialSuccess.GetRejectedSpans()
		if n != 0 || msg != "" {
			return fmt.Errorf("failed to export %d spans: %s", n, msg)
		}
	}
	if status.Code(err) == codes.OK {
		return nil
	}
	return err
}

func (c *Client) UploadMetrics(ctx context.Context, protoMetrics []*metricspb.ResourceMetrics) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.conn == nil {
		return ErrNotStarted
	}

	serviceClient := colmetricpb.NewMetricsServiceClient(c.conn)
	ctx, cancel := c.newContext(ctx)
	defer cancel()

	resp, err := serviceClient.Export(ctx, &colmetricpb.ExportMetricsServiceRequest{
		ResourceMetrics: protoMetrics,
	})
	if resp != nil && resp.PartialSuccess != nil {
		msg := resp.PartialSuccess.GetErrorMessage()
		n := resp.PartialSuccess.GetRejectedDataPoints()
		if n != 0 || msg != "" {
			return fmt.Errorf("failed to export %d metrics: %s", n, msg)
		}
	}
	if status.Code(err) == codes.OK {
		return nil
	}
	return err
}

func (c *Client) UploadLogs(ctx context.Context, protoLogs []*logspb.ResourceLogs) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.conn == nil {
		return ErrNotStarted
	}

	serviceClient := collogspb.NewLogsServiceClient(c.conn)
	ctx, cancel := c.newContext(ctx)
	defer cancel()

	resp, err := serviceClient.Export(ctx, &collogspb.ExportLogsServiceRequest{
		ResourceLogs: protoLogs,
	})
	if resp != nil && resp.PartialSuccess != nil {
		msg := resp.PartialSuccess.GetErrorMessage()
		n := resp.PartialSuccess.GetRejectedLogRecords()
		if n != 0 || msg != "" {
			return fmt.Errorf("failed to export %d logs: %s", n, msg)
		}
	}
	if status.Code(err) == codes.OK {
		return nil
	}
	return err
}

func (c *Client) Stop(ctx context.Context) error {
	err := ctx.Err()
	// wait trace uploads to finish
	acquired := make(chan struct{})
	go func() {
		c.mu.Lock()
		close(acquired)
	}()

	select {
	case <-ctx.Done():
		c.stopFunc()
		err = ctx.Err()

		<-acquired
	case <-acquired:
	}
	defer c.mu.Unlock()
	if c.conn == nil {
		return ErrAlreadyClosed
	}
	closeErr := c.conn.Close()
	if err == nil && closeErr != nil {
		err = closeErr
	}
	c.conn = nil
	return err
}
