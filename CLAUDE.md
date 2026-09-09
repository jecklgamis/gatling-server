# CLAUDE.md

## Project Overview

An API server for running [Gatling](https://gatling.io/) OSS load test simulations. Accepts a self-contained
simulation jar (via HTTP upload, S3 download, or a generic http(s)/s3 URL), runs it as a subprocess, and exposes
endpoints for task status, console/simulation logs, results, and abort.

[gatling-mcp-server](https://github.com/jecklgamis/gatling-mcp-server) wraps this API as an MCP server, so
simulations can be submitted and monitored in plain English from an MCP-capable AI client (Claude Code, Claude.ai,
Cursor, etc.) instead of hand-writing `curl` calls - see the README's "AI Integration" section.

Sibling projects:
- `../gatling-mcp-server` - the MCP server wrapping this project's task API for AI clients.
- `../gatling-scala-example`, `../gatling-java-example`, `../gatling-kotlin-example` - example simulation projects
  in each language, useful for producing a jar to test with.

## Tech Stack

- Go
- gorilla/mux (routing)
- viper (YAML config)
- log/slog (logging)
- AWS SDK (S3 upload/download)

## Project Structure

- `cmd/server/` - main entry point
- `pkg/server/` - config loading, route wiring, server startup
- `pkg/handler/` - HTTP handlers (task submit/upload, status, logs, results, abort, browse)
- `pkg/taskmanager/` - task lifecycle/state, runs simulations as subprocesses
- `pkg/gatling/` - invokes the Gatling jar itself
- `pkg/uploader/` - artifact uploaders (S3)
- `pkg/s3/` - S3 client wrapper
- `pkg/event/` - event bus and lifecycle/heartbeat events
- `pkg/accesslog/` - per-request access-log middleware
- `pkg/workspace/` - per-task workspace directory management
- `pkg/integrationtest/` - integration tests exercising the real handler stack
- `configs/config-{dev,prod,test}.yaml` - environment configs, selected via `APP_ENVIRONMENT` (kept byte-identical
  across environments; differences are made via env var overrides, not by diverging the files)
- `scripts/dist/run-server.sh` - bundled into release archives, expects the packaged `bin/` layout
  (`APP_ENVIRONMENT=prod`, `SCRIPTS_DIR=bin`)
- `run-server.sh` (repo root) - for running from source during development (`APP_ENVIRONMENT=dev`,
  `SCRIPTS_DIR=scripts`, no build step)
- `deployment/k8s/helm/` - Helm chart for Kubernetes deployment
- `.github/workflows/` - CI/CD (test, build, Docker push, release)

## Configuration

- `APP_ENVIRONMENT` (default `prod`) - selects `configs/config-<env>.yaml`.
- `SCRIPTS_DIR` - overrides `scriptsDir` from config; used to point at `scripts/` when running from source vs.
  `bin/` in a packaged release.
- `API_TOKEN` (default `default`) - bearer token for the `/task/*` API; overrides config if set.
- `BROWSE_USERNAME` / `BROWSE_PASSWORD` (default `default`) - HTTP Basic Auth for `/workspace/` and `/uploads/`;
  overrides config if set.
- `logLevel` in config (`debug`/`info`/`warn`/`error`, default `info`).
- `accessLog.enabled` (default `false`) / `accessLog.file` in config - per-request access logging, off by default;
  routes to a separate file if `file` is set, otherwise interleaved with the application log stream.
- `taskSubmit.allowedHttpHosts` in config - allowlist (with optional per-host `basic`/`bearer` auth) for
  `/task/submit`'s http(s) downloader, beyond anything that resolves to a public IP.
- `downloaders.s3`/`uploaders` in config - S3 download/upload, disabled by default; S3 downloads also require
  `allowedBuckets` to be set.

## Running

```bash
./run-server.sh
```

## Testing

```bash
go test -short ./...   # short tests
go test ./...          # all tests, including S3 integration tests requiring AWS_REGION/*_S3_URL env vars
```

## Docker

```bash
docker run -it --name gatling-server -p 58080:58080 -e API_TOKEN=some-secret-token jecklgamis/gatling-server:main
```

## Known limitations

- Simulations run as subprocesses on whichever host/pod handles the request; task state is in-memory per instance,
  so running multiple replicas means a task's status/logs are only reachable from the replica that started it
  (no shared task store).
- Simulation jars must be self-contained (all dependencies, including Scala and Gatling itself, bundled in).
