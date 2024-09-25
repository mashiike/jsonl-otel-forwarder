local caller = std.native('caller_identity')();

{
  Description: 'Example of jsonl-otel-forwarder',
  Architectures: ['arm64'],
  Environment: {
    Variables: {
      FORWARDER_LOG_LEVEL: 'debug',
      FORWARDER_EXPORT_ENDPOINT_URL: 'https://otlp.mackerelio.com:4317/',
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
