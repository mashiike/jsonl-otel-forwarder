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
$ cat trace.json | jsonl-otel-forwarder --export-endpoint-url http://localhost:4317
```

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


### Usage on AWS Lambda with AWS CloudWatch Logs Subscription Filter

see [examples](./_examples/) directory.

CloudWatch Logs to OpenTelemetry Collector

## License

MIT License
