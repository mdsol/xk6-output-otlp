# Contributing

## Branches and Pull Requests

The `main` branch is considered the source of stable releases.
Please open a PR with your proposed change.

## Building Extended K6 Binary

1. [Download and install Go](https://go.dev/doc/install) if required.
2. [Install XK6 tool](https://github.com/grafana/xk6/?tab=readme-ov-file#install-xk6)
3. Clone repository into a new folder.
4. Go to the new folder.
5. Build the extension with `make build` command. Find new K6 binary in `./bin` subfolder.
6. Run K6 tests with it using `--out otlp` flag, like

   ```sh
   ./bin/k6 run --out otlp --config <./samples/config.json> ./samples/test.js
   ```

## Upload metrics to Vision sandbox

Example of local configuration file (VPN Required):

```yaml
{
	"collectors": {
		"otlp": {
			"metrics_url": "https://motel-collector-sandbox-metrics.telemetry.nonprod-medidata.net/v1/metrics",
			"headers": {
				"job": "tests"
			},
			"push_interval": "10s",
			"timeout": "3s",
			"gzip": true,
			"insecure": false,
			"rate_conversion": "gauge",
			"trend_conversion": "histogram",
			"add_id_attributes": true
		}
	}
}
```

## Tagging

If the deployment creates AWS resources, please use the [Convention](./doc/resource-tagging-convention.md).
