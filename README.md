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
* Docker image on Docker Hub, plus prebuilt binaries on [GitHub Releases](https://github.com/jecklgamis/gatling-server/releases)

## API Reference

| Endpoint                         | Method | Auth   | Body / Params                                              | Description                                                     |
|-----------------------------------|--------|--------|--------------------------------------------------------------|-------------------------------------------------------------------|
| `/`                                | GET    | —      | —                                                            | Root info                                                          |
| `/buildInfo`                       | GET    | —      | —                                                            | Version/branch info                                                |
| `/probe/ready`                     | GET    | —      | —                                                            | Readiness probe                                                    |
| `/probe/live`                      | GET    | —      | —                                                            | Liveness probe                                                     |
| `/task/upload/http`                | POST   | Bearer | multipart: `file`, `simulation`, `javaOpts`                   | Upload a jar and submit + run it in one call                       |
| `/upload`                          | POST   | Bearer | multipart: `file`                                            | Upload any file; returns `{"id": "<uuid>"}`                        |
| `/uploads/{id}/{filename}`         | GET    | —      | —                                                            | Download/browse an uploaded file                                   |
| `/task/submit`                     | POST   | Bearer | JSON: `simulation`, `javaOpts`, `url` (http(s) or s3)         | Download the jar from `url` and submit + run it                    |
| `/task/download/s3`                | POST   | Bearer | JSON: `simulation`, `javaOpts`, `url` (s3 only)               | Same as above, S3-only (kept for compatibility; needs S3 downloader enabled) |
| `/task/{taskId}`                   | GET    | —      | —                                                            | Task runtime status                                                 |
| `/task/metadata/{taskId}`          | GET    | —      | —                                                            | Original submission metadata                                        |
| `/task/console/{taskId}`           | GET    | —      | —                                                            | Raw JVM console log                                                 |
| `/task/simulationLog/{taskId}`     | GET    | —      | —                                                            | Gatling's own simulation log                                        |
| `/task/results/{taskId}`           | GET    | —      | —                                                            | Results archive (`results.tar.gz`)                                  |
| `/task/abort/{taskId}`             | POST   | —      | —                                                            | Kill a running task                                                 |
| `/workspace/{taskId}/...`          | GET    | —      | —                                                            | Browse raw task workspace files                                     |
| `/blackhole`                       | POST   | —      | —                                                            | No-op sink (default HTTP event-notifier target)                     |

`Bearer` means `Authorization: Bearer <API_TOKEN>` is required; missing/invalid tokens get `401 Unauthorized`, and
repeated failures from the same client are rate-limited with `429 Too Many Requests`.

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

### Via a generic submit (upload once, submit anywhere)

Upload a jar to get a URL back, then submit a task referencing any http(s) or s3 URL — including the one you just
got:

```bash
curl -H "Authorization: Bearer ${API_TOKEN}" -F "file=@target/gatling-scala-example.jar" http://localhost:58080/upload
# => {"id":"<uuid>"}
```

Uploaded files are stored at `uploads/<uuid>/<filename>` and served directly (no auth) from `/uploads/<uuid>/<filename>`.

```bash
curl -v -H "Content-Type: application/json" http://localhost:58080/task/submit -d @request.json
```

`request.json`:

```json
{
  "url": "http://localhost:58080/uploads/<uuid>/gatling-scala-example.jar",
  "simulation": "gatling.test.example.simulation.ExampleSimulation",
  "javaOpts": "-DbaseUrl=http://localhost:8080 -DdurationMin=0.10 -DrequestPerSecond=1"
}
```

`url` also accepts `s3://...` locations, in which case this behaves the same as the S3 download flow above.

### Aborting a task

```bash
curl -X POST http://localhost:58080/task/abort/{taskId}
```

## Retrieving Artifacts

A simulation run produces a console log, Gatling report, simulation log, and the original request's metadata (see the
`/task/*` rows in the API Reference above). These are available directly from the server, and are also uploaded to S3
if an S3 uploader is configured. The test report is a downloadable `tar.gz` archive. The whole workspace directory
(one subdirectory per task, containing the raw files above) is also browsable directly at
`http://localhost:58080/workspace/{taskId}/`.

## Authoring Simulations

Gatling simulations are written in Scala. Simple simulations can be submitted as-is; for anything more involved, a
build project (Maven, for example) makes packaging much easier. See the example projects for a working setup in your
language of choice:

* [gatling-scala-example](https://github.com/jecklgamis/gatling-scala-example)
* [gatling-java-example](https://github.com/jecklgamis/gatling-java-example)
* [gatling-kotlin-example](https://github.com/jecklgamis/gatling-kotlin-example)
