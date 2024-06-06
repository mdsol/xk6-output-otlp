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

## Tagging

If the deployment creates AWS resources, please use the [Convention](./doc/resource-tagging-convention.md).
