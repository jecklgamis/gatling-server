## gatling-server

[![Build](https://github.com/jecklgamis/gatling-server/actions/workflows/build.yaml/badge.svg)](https://github.com/jecklgamis/gatling-server/actions/workflows/build.yaml)

An API server for running [Gatling](https://gatling.io/) OSS load test simulations.

## Features

* Runs simulations packaged as a self-contained jar (simulation classes and resources bundled together)
* Task submission via HTTP upload, S3 download, or a generic http(s)/s3 URL
* Standalone file upload endpoint with a browsable uploads directory
* Artifact upload to S3 (metadata, console log, results, etc.)
* Endpoints for task metadata, console log, simulation log, and results
* HTTP and SNS event notifiers for heartbeat and task lifecycle events
* Docker image on Docker Hub, plus prebuilt binaries and a Helm chart
* AI integration via [gatling-mcp-server](https://github.com/jecklgamis/gatling-mcp-server)

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

See the [docs site](https://jecklgamis.github.io/gatling-server/) for deployment options and AI integration.

## Getting Started

```bash
git clone https://github.com/jecklgamis/gatling-server.git
cd gatling-server
./run-server.sh
```

## Documentation

Running it with Docker or a prebuilt binary, deploying it to Kubernetes, the full HTTP API reference, and AI
integration via gatling-mcp-server all live in the docs site:
**[jecklgamis.github.io/gatling-server](https://jecklgamis.github.io/gatling-server/)** (source in [`docs/`](docs)).
