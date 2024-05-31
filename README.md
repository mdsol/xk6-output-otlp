# K6 OTLP Output Extension

This repository is for [K6 output extension](https://k6.io/docs/extensions/). The extension exports metrics of K6 tests using OTLP/HTTP protocol.

## Usage

### Building extended K6 binary

1. [Download and install Go](https://go.dev/doc/install) if required.
2. [Install XK6 tool](https://github.com/grafana/xk6/?tab=readme-ov-file#install-xk6)
3. Clone repository into a new folder.
4. Go to the new folder.
5. Build the extension with `make build` command. Find new K6 binary in `./bin` subfolder.
6. Run tests with it using `--out otlp` flag, like

   ```sh
   make build && ./bin/k6 run --out otlp --config ./samples/config.json ./samples/test.js
   ```

### Configuration

Configuration parameters can be set in a JSON configuration file or using environment variables:

Example of configuration file:

```json
{
  "collectors": {
    "otlp": {
      "metrics_url": "http://localhost:8080/v1/metrics",
      "headers": {
        "job": "tests"
      },
      "push_interval": "1m",
      "timeout": "3s",
      "gzip": true,
      "insecure": true,
      "trend_conversion": "gauges"
    }
  }
}
```

Environment variables:

| Environment Variable       | Description |
|-------------------------   |-------------|
| `K6_OTLP_GZIP`             | `true` or `false`. Use GZIP encoding or not.  |
| `K6_OTLP_HTTP_HEADERS`     | Optional HTTP headers |
| `K6_OTLP_INSECURE`         | `true` or `false`. Validate SSL certificate or not. |
| `K6_OTLP_PUSH_INTERVAL`    | Metric push interval in Go duration format for intermediate metrics. At the end on the test metrics exported regardless of this value. |
| `K6_OTLP_SERVER_URL`       | OTLP metrics endpoint url. Usually ends with `/v1/metrics` |
| `K6_OTLP_TIMEOUT`          | HTTP request timeout  in Go duration format |
| `K6_OTLP_TREND_CONVERSION` | `gauges` or `histogram`. Conversion type for metrics of type `trend`. |

### Run Tests

Example:

```sh
# Build for the first time
make build
# From 
./bin/k6 run --out otlp --config ./samples/config.json  ./samples/test.js
```

### K6 Metrics Conversion

The Grafana K6 testing utility uses a metric model that requires some metrics conversion before sending to Opentelemetry Collector or other OTLP receiver.

#### Rate

A metric of type "rate", which internally is a sequence of 0 and 1 values, is converted to float gauge.
The values is `sum/count` of measures.

#### Trend

A metric of type "trend" could be converted to a collection of gauges (default) or a histogram, depending on the test configuration.

It is expected that using conversion to Gauges we get the same results we have in the test output.
For each statistic type of the trend we add `stat` label with appropriate value (`min`, `max`, `avg`, `p90`, etc.).

Conversion a trend to OpenTelemetry histogram is an experimental feature.
