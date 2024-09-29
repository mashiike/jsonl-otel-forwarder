# jsonl-otel-forwarder

## Overview

`jsonl-otel-forwarder` is a tool designed to forward JSON Lines (JSONL) formatted logs to OpenTelemetry (OTel) collectors. This tool helps in integrating log data with observability platforms that support OpenTelemetry.

## Features

- Forward JSONL logs to OTel collectors
- Easy configuration and deployment

## Installation

To install `jsonl-otel-forwarder`, use the following command:

```sh
go get github.com/mashiike/jsonl-otel-forwarder
```

## Usage on Local
To use jsonl-otel-forwarder, run the following command:

```sh
$ cat trace.json | jsonl-otel-forwarder --otlp-endpoint http://localhost:4317
```

trace.json:
```json
{
  "resourceSpans": [
    {
      "resource": {
        "attributes": [
          {
            "key": "service.name",
            "value": {
              "stringValue": "my.service"
            }
          }
        ]
      },
      "scopeSpans": [
        {
          "scope": {
            "name": "my.library",
            "version": "1.0.0",
            "attributes": [
              {
                "key": "my.scope.attribute",
                "value": {
                  "stringValue": "some scope attribute"
                }
              }
            ]
          },
          "spans": [
            {
              "traceId": "5B8EFFF798038103D269B633813FC60C",
              "spanId": "EEE19B7EC3C1B174",
              "parentSpanId": "EEE19B7EC3C1B173",
              "name": "I'm a server span",
              "startTimeUnixNano": "1544712660000000000",
              "endTimeUnixNano": "1544712661000000000",
              "kind": 2,
              "attributes": [
                {
                  "key": "my.span.attr",
                  "value": {
                    "stringValue": "some value"
                  }
                }
              ]
            }
          ]
        }
      ]
    }
  ]
}
```

flag options are as follows:

```sh
$ jsonl-otel-forwarder --help
Usage: jsonl-otel-forwarder [options]

         Forward JSON Lines OTLP logs to OpenTelemetry Collector/Server

Options:

  -log-level string
        log level ($FORWARDER_LOG_LEVEL) (default "info")
  -otlp-endpoint string
        OTLP endpoint to use, e.g. http://localhost:4317 ($FORWARDER_OTLP_ENDPOINT,$OTEL_EXPORTER_OTLP_ENDPOINT)
  -otlp-headers string
        OTLP headers to use, e.g. key1=value1,key2=value2 ($FORWARDER_OTLP_HEADERS,$OTEL_EXPORTER_OTLP_HEADERS)
  -otlp-logs-endpoint string
        OTLP logs endpoint to use, overrides --otlp-endpoint ($FORWARDER_OTLP_LOGS_ENDPOINT,$OTEL_EXPORTER_OTLP_LOGS_ENDPOINT)
  -otlp-logs-headers string
        OTLP logs headers to use, append or override --otlp-headers ($FORWARDER_OTLP_LOGS_HEADERS,$OTEL_EXPORTER_OTLP_LOGS_HEADERS)
  -otlp-logs-protocol string
        OTLP logs protocol to use, overrides --otlp-protocol ($FORWARDER_OTLP_LOGS_PROTOCOL,$OTEL_EXPORTER_OTLP_LOGS_PROTOCOL)
  -otlp-logs-timeout string
        OTLP logs export timeout to use, overrides --otlp-timeout ($FORWARDER_OTLP_LOGS_TIMEOUT,$OTEL_EXPORTER_OTLP_LOGS_TIMEOUT)
  -otlp-metrics-endpoint string
        OTLP metrics endpoint to use, overrides --otlp-endpoint ($FORWARDER_OTLP_METRICS_ENDPOINT,$OTEL_EXPORTER_OTLP_METRICS_ENDPOINT)
  -otlp-metrics-headers string
        OTLP metrics headers to use, append or override --otlp-headers ($FORWARDER_OTLP_METRICS_HEADERS,$OTEL_EXPORTER_OTLP_METRICS_HEADERS)
  -otlp-metrics-protocol string
        OTLP metrics protocol to use, overrides --otlp-protocol ($FORWARDER_OTLP_METRICS_PROTOCOL,$OTEL_EXPORTER_OTLP_METRICS_PROTOCOL)
  -otlp-metrics-timeout string
        OTLP metrics export timeout to use, overrides --otlp-timeout ($FORWARDER_OTLP_METRICS_TIMEOUT,$OTEL_EXPORTER_OTLP_METRICS_TIMEOUT)
  -otlp-protocol string
        OTLP protocol to use e.g. grpc, http/json, http/protobuf ($FORWARDER_OTLP_PROTOCOL,$OTEL_EXPORTER_OTLP_PROTOCOL)
  -otlp-timeout string
        OTLP export timeout to use, e.g. 5s ($FORWARDER_OTLP_TIMEOUT,$OTEL_EXPORTER_OTLP_TIMEOUT)
  -otlp-traces-endpoint string
        OTLP traces endpoint to use, overrides --otlp-endpoint ($FORWARDER_OTLP_TRACES_ENDPOINT,$OTEL_EXPORTER_OTLP_TRACES_ENDPOINT)
  -otlp-traces-headers string
        OTLP traces headers to use, append or override --otlp-headers ($FORWARDER_OTLP_TRACES_HEADERS,$OTEL_EXPORTER_OTLP_TRACES_HEADERS)
  -otlp-traces-protocol string
        OTLP traces protocol to use, overrides --otlp-protocol ($FORWARDER_OTLP_TRACES_PROTOCOL,$OTEL_EXPORTER_OTLP_TRACES_PROTOCOL)
  -otlp-traces-timeout string
        OTLP traces export timeout to use, overrides --otlp-timeout ($FORWARDER_OTLP_TRACES_TIMEOUT,$OTEL_EXPORTER_OTLP_TRACES_TIMEOUT)
  -signals string
        comma separated list of signals to forward [traces,metrics,logs] ($FORWARDER_SIGNALS) (default "traces,metrics,logs")
```

options priority is as follows:

1. command line options
2. `FORWARDER_` prefixed environment variables
3. `OTEL_EXPORTER_` prefixed environment variables

### Usage on AWS Lambda with AWS CloudWatch Logs Subscription Filter

see [examples](./_examples/) directory.

CloudWatch Logs to OpenTelemetry Collector

## License

MIT License
