## gatling-server

[![Build](https://github.com/jecklgamis/gatling-server/actions/workflows/build.yaml/badge.svg)](https://github.com/jecklgamis/gatling-server/actions/workflows/build.yaml)

An API server for running [Gatling](https://gatling.io/) OSS load test simulations.

This README covers **developing** gatling-server. For running, deploying, using the API, and AI integration, see
the docs site: **[jecklgamis.github.io/gatling-server](https://jecklgamis.github.io/gatling-server/)**
(or browse [`docs/`](docs) directly).

## Features

* Runs simulations packaged as a self-contained jar (simulation classes and resources bundled together)
* Task submission via HTTP upload, S3 download, or a generic http(s)/s3 URL
* Standalone file upload endpoint with a browsable uploads directory
* Artifact upload to S3 (metadata, console log, results, etc.)
* Endpoints for task metadata, console log, simulation log, and results
* HTTP and SNS event notifiers for heartbeat and task lifecycle events
* Docker image on Docker Hub, plus prebuilt binaries and a Helm chart
* AI integration via [gatling-mcp-server](https://github.com/jecklgamis/gatling-mcp-server)

## Getting Started

### From Source (for development)

```bash
./run-server.sh
```

Generates a self-signed TLS cert if one isn't already present, then runs `cmd/server/gatling-server.go` directly
with `APP_ENVIRONMENT=dev` (loads `configs/config-dev.yaml`) and `SCRIPTS_DIR=scripts` - no build step, no `bin/`
layout. This is distinct from `scripts/dist/run-server.sh`, which is bundled into release archives and expects the
packaged `bin/` layout (`APP_ENVIRONMENT=prod`, `SCRIPTS_DIR=bin`).

### Verify it's Up

```bash
curl http://localhost:58080/buildInfo
```

HTTP uploads require an API token, sent as a bearer token in the `Authorization` header. It defaults to `default`
unless the server was started with its own `API_TOKEN`; requests with a missing or invalid token get `401 Unauthorized`.

Log verbosity is set via `logLevel` in `configs/config-<env>.yaml` (`debug`, `info`, `warn`, or `error`; defaults to
`info` if unset or unrecognized).

Per-request access logging is disabled by default. Set `accessLog.enabled: true` in `configs/config-<env>.yaml` to
turn it on; set `accessLog.file` to a file path to route those entries there (as JSON lines), separate from the
application log stream, or leave it empty to interleave them with the rest of the application's logs.

### Testing

```bash
go test -short ./...   # short tests
go test ./...          # all tests, including S3 integration tests requiring AWS_REGION/*_S3_URL env vars
```

## Documentation

Running it with Docker or a prebuilt binary, deploying it to Kubernetes, the full HTTP API reference, and AI
integration via gatling-mcp-server all live in the docs site:
**[jecklgamis.github.io/gatling-server](https://jecklgamis.github.io/gatling-server/)** (source in [`docs/`](docs)).
