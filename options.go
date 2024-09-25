package jsonlotelforwarder

import (
	"flag"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type HeaderSlice []string

func NewHeaderSlice(m map[string]string) HeaderSlice {
	headers := make(HeaderSlice, 0, len(m))
	for k, v := range m {
		headers = append(headers, fmt.Sprintf("%s: %s", k, v))
	}
	return headers
}

func (h *HeaderSlice) String() string {
	return fmt.Sprint(*h)
}

func (h *HeaderSlice) Set(value string) error {
	*h = append(*h, value)
	return nil
}

func (h *HeaderSlice) Build() metadata.MD {
	if h == nil {
		return nil
	}
	md := metadata.New(nil)
	for _, header := range *h {
		parts := strings.SplitN(header, ":", 2)
		if len(parts) != 2 {
			slog.Warn("invalid header format, must be key=value, skip this header", "header", header)
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])
		md.Append(key, value)
	}
	slog.Debug("builded headers", "metadata", md)
	return md
}

type Options struct {
	EndpointURL string
	Headers     HeaderSlice
	GZip        bool
	Signals     string

	u                   *url.URL
	mu                  sync.Mutex
	parsedEndpointURL   string
	urlParseErr         error
	exportTimeout       time.Duration
	ExportTimeout       string
	parsedExportTimeout time.Duration
	timtoutParseErr     error
	md                  metadata.MD
}

func DefaultOptions() *Options {
	return &Options{
		EndpointURL:   "http://localhost:4317",
		Headers:       make(HeaderSlice, 0),
		Signals:       "traces,metrics,logs",
		ExportTimeout: "10s",
	}
}

const EnvPrefix = "FORWARDER_"

func (o *Options) SetFlags(fs *flag.FlagSet) {
	fs.StringVar(&o.EndpointURL, "export-endpoint-url", o.EndpointURL, "OTLP gRPC endpoint URL e.g. http://localhost:4317 ($FORWARDER_EXPORT_ENDPOINT_URL)")
	fs.Var(&o.Headers, "header", "Add a header to the request e.g. 'key:value' ($FORWARDER_HEADER,$FORWARDER_HEADER_*)")
	fs.BoolVar(&o.GZip, "gzip", o.GZip, "Enable gzip compression ($FORWARDER_GZIP)")
	fs.StringVar(&o.Signals, "signals", o.Signals, "Comma separated list of signals to forward e.g. traces,metrics,logs ($FORWARDER_SIGNALS)")
	fs.StringVar(&o.ExportTimeout, "export-timeout", o.ExportTimeout, "Timeout for export request e.g. 10s ($FORWARDER_EXPORT_TIMEOUT)")
	fs.VisitAll(func(f *flag.Flag) {
		names := []string{
			strings.ToUpper(EnvPrefix + strings.ReplaceAll(f.Name, "-", "_")),
			strings.ToLower(EnvPrefix + strings.ReplaceAll(f.Name, "-", "_")),
			strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_")),
			strings.ToLower(strings.ReplaceAll(f.Name, "-", "_")),
		}
		for _, name := range names {
			if s, ok := os.LookupEnv(name); ok {
				f.Value.Set(s)
				break
			}
		}
	})
	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, EnvPrefix+"HEADER_") && !strings.HasPrefix(env, "HEADER_") {
			continue
		}
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		o.Headers.Set(parts[1])
	}
}

func (o *Options) DialOptions() []grpc.DialOption {
	opts := []grpc.DialOption{
		grpc.WithUserAgent("jsonl-otel-forwarder/" + Version),
	}
	u := o.endpointURL()
	if u.Scheme != "https" {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		cred := credentials.NewTLS(nil)
		opts = append(opts, grpc.WithTransportCredentials(cred))
	}
	if o.GZip {
		opts = append(opts, grpc.WithDefaultCallOptions(grpc.UseCompressor("gzip")))
	}
	return opts
}

func (o *Options) Endpoint() string {
	return o.endpointURL().Host
}

func (o *Options) endpointURL() *url.URL {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.u == nil || o.parsedEndpointURL != o.EndpointURL {
		u, err := url.Parse(o.EndpointURL)
		o.u = u
		o.urlParseErr = err
		o.parsedEndpointURL = o.EndpointURL
		slog.Debug("parsed endpoint URL", "endpoint", u.Host, "insecure", u.Scheme != "https")
	}
	return o.u
}

func (o *Options) headers() metadata.MD {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.md == nil {
		o.md = o.Headers.Build()
	}
	return o.md
}

func (o *Options) exportTimeoutDuration() time.Duration {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.parsedExportTimeout != o.exportTimeout {
		d, err := time.ParseDuration(o.ExportTimeout)
		o.exportTimeout = d
		o.timtoutParseErr = err
	}
	return o.exportTimeout
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
	o.endpointURL()
	if o.urlParseErr != nil {
		return o.urlParseErr
	}
	o.exportTimeoutDuration()
	if o.timtoutParseErr != nil {
		return o.timtoutParseErr
	}
	o.headers()
	return nil
}
