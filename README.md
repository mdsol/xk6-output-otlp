# K6 OTLP Output Extension

This repository is for [K6 output extension](https://k6.io/docs/extensions/). The extension exports metrics of K6 tests using OTLP/HTTP protocol.

## Usage

Get the Extended K6 Binary

- Find the [latest release}(https://github.com/mdsol/xk6-output-otlp/releases).
- Download `k6.tar.gz` archive, and extract the `k6` binary with

  ```sh
  tar -xvzf k6.tar.gz
  ```

- Run K6 tests like

  ```sh
  ./k6 run --out otlp --config <config-file> <test-file>
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
      "rate_conversion": "counters",
      "trend_conversion": "gauges",
      "add_id_attributes": true
    }
  }
}
```

Environment variables:

| Environment Variable       | Default Value | Description |
|----------------------------|---------------|-------------|
| `K6_OTLP_ADD_ID_ATTRS`     | `false`       | If `true`, attributes `provider_id` and `run_id` added to metrics. |
| `K6_OTLP_GZIP`             | `false`       | `true` or `false`. Use GZIP encoding or not.  |
| `K6_OTLP_HTTP_HEADERS`     | empty         | Optional HTTP headers |
| `K6_OTLP_INSECURE`         | `true`        | `true` or `false`. Validate SSL certificate or not. |
| `K6_OTLP_PUSH_INTERVAL`    | `5s`          | Metric push interval in Go duration format for intermediate metrics. At the end on the test metrics exported regardless of this value. |
| `K6_OTLP_TIMEOUT`          | `5s`          | HTTP request timeout  in Go duration format |
| `K6_OTLP_RATE_CONVERSION`  | `counters`    | `counters` or `gauge`. Conversion type for metrics of type `rate`. |
| `K6_OTLP_TREND_CONVERSION` | `gauges`      | `gauges` or `histogram`. Conversion type for metrics of type `trend`. |
| `K6_OTLP_SERVER_URL`       | `http://localhost:8080/v1/metrics`| OTLP metrics endpoint url. Usually ends with `/v1/metrics` |

## K6 Metrics Conversion

The Grafana K6 testing utility uses a metric model that requires some metrics conversion before sending to Opentelemetry Collector or other OTLP receiver.

### Rate

A metric of type "rate" could be converted to a collection of counters (default) or a gauge, depending on the test configuration.

If the rate type conversion is `counter`, the converted metric is a pair of counters withe labels "value=0|1".
To calculate actual rate value, in this case we need to make an equation:

```text
rate = rate(counter{value==1} / (counter{value==1} + counter{value==0}))
```

If the rate type conversion is `gauge`, the result metric is a float gauge which value is pre-calculated.

### Trend

A metric of type "trend" could be converted to a collection of gauges (default) or a histogram, depending on the test configuration.

It is expected that using conversion to Gauges we get the same results we have in the test output.
For each statistic type of the trend we add `stat` label with appropriate value (`min`, `max`, `avg`, `p90`, etc.).

Conversion a trend to OpenTelemetry histogram is an experimental feature.

### Attributes

Some time it is required to distinguish metrics received from different test runs. For this scenario we have
a configuration option that adds the following attributes/labels to the metrics:

| Attribute     | Meaning |
|---------------|---------|
| `provider_id` | Initially, a random generated string. Expected to be the same for the same running location (host, user) |
| `run_id`      | A cyclic counter 0..255 indicates the test run. We should limit the range for avoiding high-cardinality. |

To insert attributes, set environment variable `K6_OTLP_ADD_ID_ATTRS=true`
or set `add_id_attributes: true` in the config file.

#### Using of meaningful provider_id

You can update your local file `~/.xk6-output-otlp/provider_id` by changing the randomly generated id to
something meaningful (your hostname or username, etc.).
The new id should be unique and match pattern `^[0-9a-zA-Z_]{3,16}$`.

## Contributing

See [CONTRIBUTING](./CONTRIBUTING.md)

## Contact

See the [factbook](factbook.yaml).
