# gatling-server

An API server for running [Gatling](https://gatling.io/) OSS load test simulations: accepts a self-contained
simulation jar (via HTTP upload, S3 download, or a generic http(s)/s3 URL), runs it as a subprocess, and exposes
endpoints for task status, console/simulation logs, results, and abort.

This site covers **running, deploying, and using** gatling-server. For building the project itself (Go toolchain,
tests, project layout), see the [repo's README](https://github.com/jecklgamis/gatling-server#readme).

## Features

* Runs simulations packaged as a self-contained jar (simulation classes and resources bundled together)
* Task submission via HTTP upload, S3 download, or a generic http(s)/s3 URL
* Standalone file upload endpoint with a browsable uploads directory
* Artifact upload to S3 (metadata, console log, results, etc.)
* Endpoints for task metadata, console log, simulation log, and results
* HTTP and SNS event notifiers for heartbeat and task lifecycle events
* Docker image on Docker Hub, plus prebuilt binaries and a Helm chart

## Where to go next

- **[Deployment](deployment.md)** - run it with Docker, a prebuilt binary, or deploy it to Kubernetes.
- **[Usage](usage.md)** - submit a simulation, retrieve artifacts, and author one from scratch.
- **[API Reference](api.md)** - every endpoint, its auth scheme, and its body/params.
- **[AI Integration](ai-integration.md)** - driving gatling-server from an MCP-capable AI client via
  [gatling-mcp-server](https://github.com/jecklgamis/gatling-mcp-server).
