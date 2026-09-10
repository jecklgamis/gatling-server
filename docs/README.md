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

## Quick Start

### 1. Run gatling-server

```bash
docker run -it --name gatling-server -p 58080:58080 -e API_TOKEN=some-secret-token jecklgamis/gatling-server:main
```

### 2. Build a simulation jar

```bash
git clone https://github.com/jecklgamis/gatling-scala-example.git
cd gatling-scala-example
./mvnw clean package
```

This produces a self-contained `target/gatling-scala-example.jar` (simulation classes, resources, and all
dependencies - including Scala and Gatling itself - bundled together).

### 3. Submit it

```bash
curl -v \
  -H "Authorization: Bearer some-secret-token" \
  -F "file=@target/gatling-scala-example.jar" \
  -F "simulation=gatling.test.example.simulation.ExampleSimulation" \
  -F "javaOpts=-DbaseUrl=http://localhost:8080 -DdurationMin=1 -DrequestPerSecond=10" \
  http://localhost:58080/task/upload
```

The response includes a `taskId` - poll `http://localhost:58080/task/{taskId}` for status.

### 4. Browse the Console Log and Report

Open `http://localhost:58080/workspace/{taskId}/` in a browser to view the raw task workspace - console log,
Gatling report, and simulation log included. It's protected by HTTP Basic Auth (username/password both `default`
unless `BROWSE_USERNAME`/`BROWSE_PASSWORD` were set) - your browser will prompt for them.

See [Usage](usage.md) for retrieving these same artifacts via the API instead of a browser, [Deployment](deployment.md)
for other ways to run gatling-server, and [AI Integration](ai-integration.md) to drive this whole flow from an AI
agent instead of curl.

## Where to go next

- **[Deployment](deployment.md)** - run it with Docker, a prebuilt binary, or deploy it to Kubernetes.
- **[Usage](usage.md)** - submit a simulation, retrieve artifacts, and author one from scratch.
- **[API Reference](api.md)** - every endpoint, its auth scheme, and its body/params.
- **[AI Integration](ai-integration.md)** - driving gatling-server from an MCP-capable AI client via
  [gatling-mcp-server](https://github.com/jecklgamis/gatling-mcp-server).
