# K6 OTLP Output Extension

This repository is for [K6 output extension](https://k6.io/docs/extensions/). The extension exports metrics of K6 tests using OTLP/HTTP protocol.

## Usage

The released binary is expected to run in containers or linux-like hosts of `amd64` architecture.
For Windows 10+ environments you can run it in  WSL2.
For other environments, please build your version from sources.

### Get the pre-build extended K6 binary

- Find the [latest release}(https://github.com/mdsol/xk6-output-otlp/releases).
- Download `k6.tar.gz` archive, and extract the `k6` binary with

  ```sh
  tar -xvzf k6.tar.gz
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
| `K6_OTLP_SERVER_URL`       | `http://localhost:8080/v1/metrics`| OTLP metrics endpoint url. Usually ends with `/v1/metrics` |

### Run K6 tests like

  ```sh
  ./k6 run --out otlp --config <config-file> <test-file>
  ```

## K6 Metrics Conversion

The Grafana K6 testing utility uses a metric model that requires some metrics conversion before sending to Opentelemetry Collector or other OTLP receiver.

### Rate

Each metric of type "rate" is converted to three OTLP metrics.

If `<name>` is a name of the original rate metric, it produces:

- `k6_<name>_total` OTLP counter which contains the total number of occurrences;
- `k6_<name>_success_total` OTLP counter which contains the total number of successful occurrences;
- `k6_<name>_success_rate` OTLP Gauge, which contains the pre-computed rate. 

The latest pre-computed rate value should match appropriate K6 output.

### Trend

Each metric of type "trend" is converted to one OTLP histogram and
a gauge which label `stat` will define pre-computed statistics.

_Example: original K6 trend metric: `http_req_duration` produces `k6_http_req_duration{stat="ag|min|max|p50|p90|p95"}` metric family._

The latest pre-computed stat values should match appropriate K6 output.

### Attributes

Sometimes it's required to distinguish metrics received from different test runs. For this scenario we have
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
