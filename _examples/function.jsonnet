local caller = std.native('caller_identity')();

{
  Description: 'Example of jsonl-otel-forwarder',
  Architectures: ['arm64'],
  Environment: {
    Variables: {
      FORWARDER_LOG_LEVEL: 'debug',
      FORWARDER_OTLP_ENDPOINT: 'https://otlp.mackerelio.com:4317/',
      FORWARDER_OTLP_PROTOCL: 'grpc',
      FORWARDER_OTLP_TRACES_PROTOCOL: 'http/protobuf',
      FORWARDER_SIGNALS: 'traces,metrics',
      SSMWRAP_NAMES: '/jsonl-otel-forwarder/*',
      TZ: 'Asia/Tokyo',
    },
  },
  FunctionName: 'jsonl-otel-forwarder',
  Handler: 'bootstrap',
  MemorySize: 128,
  Role: 'arn:aws:iam::%s:role/jsonl-otel-forwarder' % caller.Account,
  Runtime: 'provided.al2023',
  Tags: {},
  Timeout: 300,
  TracingConfig: {
    Mode: 'PassThrough',
  },
}
