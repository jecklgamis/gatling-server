## gatling-server

[![Build](https://github.com/jecklgamis/gatling-server/actions/workflows/build.yaml/badge.svg)](https://github.com/jecklgamis/gatling-server/actions/workflows/build.yaml)

An API server for running [Gatling](https://gatling.io/) OSS load test simulations.

## Features

* Runs simulations packaged as a self-contained jar (simulation classes and resources bundled together)
* Task submission via HTTP upload or S3 download
* Artifact upload to S3 (metadata, console log, results, etc.)
* Endpoints for task metadata, console log, simulation log, and results
* HTTP and SNS event notifiers for heartbeat and task lifecycle events
* Docker image on Docker Hub, plus prebuilt binaries on [GitHub Releases](https://github.com/jecklgamis/gatling-server/releases)

## Getting Started

### Using Docker

```bash
docker run -it --name gatling-server -p 58080:58080 -e API_TOKEN=some-secret-token jecklgamis/gatling-server:main
```

### Using a prebuilt binary

Download a release for your platform from [GitHub Releases](https://github.com/jecklgamis/gatling-server/releases), then:

```bash
tar xzf gatling-server-<os>-<arch>-<version>.tar.gz
cd gatling-server-<os>-<arch>-<version>
API_TOKEN=some-secret-token ./run-server.sh
```

### Verify it's up

```bash
curl http://localhost:58080/buildInfo
```

HTTP uploads require an API token, sent as a bearer token in the `Authorization` header. It defaults to `default`
unless the server was started with its own `API_TOKEN`; requests with a missing or invalid token get `401 Unauthorized`.

## Submitting a Simulation

Simulations must be packaged as a self-contained (uber) jar containing the compiled simulation classes, resources,
and all dependencies — including Scala and Gatling itself — since the server runs it directly off that jar's
classpath. If you're using Maven, the [maven-shade-plugin](https://maven.apache.org/plugins/maven-shade-plugin/) can
build this for you; see [gatling-scala-example](https://github.com/jecklgamis/gatling-scala-example) for a working
setup that produces `target/gatling-scala-example.jar`.

### Via HTTP upload

```bash
curl -v \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -F "file=@target/gatling-scala-example.jar" \
  -F "simulation=gatling.test.example.simulation.ExampleSimulation" \
  -F "javaOpts=-DbaseUrl=http://localhost:8080 -DdurationMin=1 -DrequestPerSecond=10" \
  http://localhost:58080/task/upload/http
```

The response includes a `taskId`, used to query the server for artifacts such as console logs or Gatling reports.

### Via S3 download

Enable the S3 downloader in `configs/config-<env>.yaml`:

```yaml
downloaders:
  s3:
    enabled: true
    configMap:
      region: some-region
```

Then submit a task referencing a jar already in S3:

```bash
curl -v -H "Content-Type: application/json" http://localhost:58080/task/download/s3 -d @request.json
```

`request.json`:

```json
{
  "url": "s3://gatling-server-incoming/gatling-scala-example.jar",
  "simulation": "gatling.test.example.simulation.ExampleSimulation",
  "javaOpts": "-DbaseUrl=http://localhost:8080 -DdurationMin=0.10 -DrequestPerSecond=1"
}
```

### Aborting a task

```bash
curl -X POST http://localhost:58080/task/abort/{taskId}
```

## Retrieving Artifacts

A simulation run produces a console log, Gatling report, simulation log, and the original request's metadata. These
are available directly from the server, and are also uploaded to S3 if an S3 uploader is configured:

| Artifact       | Endpoint                                         |
|----------------|---------------------------------------------------|
| Task metadata  | `http://localhost:58080/task/metadata/{taskId}`      |
| Console output | `http://localhost:58080/task/console/{taskId}`       |
| Simulation log | `http://localhost:58080/task/simulationLog/{taskId}` |
| Test report    | `http://localhost:58080/task/results/{taskId}`       |

The test report is a downloadable `tar.gz` archive.

## Authoring Simulations

Gatling simulations are written in Scala. Simple simulations can be submitted as-is; for anything more involved, a
build project (Maven, for example) makes packaging much easier. See the example projects for a working setup in your
language of choice:

* [gatling-scala-example](https://github.com/jecklgamis/gatling-scala-example)
* [gatling-java-example](https://github.com/jecklgamis/gatling-java-example)
* [gatling-kotlin-example](https://github.com/jecklgamis/gatling-kotlin-example)
