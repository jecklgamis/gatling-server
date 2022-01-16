## gatling-server

![Go](https://github.com/jecklgamis/gatling-server/workflows/Go/badge.svg?branch=main) [![Go Report Card](https://goreportcard.com/badge/github.com/jecklgamis/gatling-server)](https://goreportcard.com/report/github.com/jecklgamis/gatling-server)

gatling-server is an API server for [Gatling](https://gatling.io/) OSS. 

Heads up: this is currently in alpha testing but feel free to try the latest Docker image from Docker Hub.

## Features

* Run single file or packaged simulations (simulations, and resources packaged as jar file)
* Task submission via HTTP upload or S3 download
* Artifacts upload to S3 (metadata, console log, results, etc.)
* Endpoints for task metadata, console log, and results, etc.
* HTTP and SNS event notifiers (heartbeat and task cycle events)
* Docker image available in Docker Hub

## Getting Started

Start the server using Docker image available from Docker Hub:

```bash
docker run -it -p 58080:58080 jecklgamis/gatling-server:latest
```

Ensure it's up by hitting the `/buildInfo` endpoint:

```bash
curl http://localhost:58080/buildInfo 
```

### Submitting Tasks

Gatling tasks are submitted via HTTP endpoints. Artifacts referenced in a request can be a single file simulation or a
packaged simulation. A single file simulation class is a single Scala file that uses only on standard Scala or Gatling
libraries. Packaged simulations contains resources such as request bodies, feeders, or additional utility classes.

A task identifier is returned in the task submission response. This can be used to query the server for generated
artifacts such as console logs or Gatling reports.

### Submitting Task Using HTTP Upload

Running a single file simulation

Example:

```bash
$ cd test
curl -v \
  -F 'file=@./testdata/SingleFileExampleSimulation.scala' \
  -F "simulation=gatling.test.example.simulation.SingleFileExampleSimulation" \
  -F "javaOpts=-DbaseUrl=http://localhost:8080 -DdurationMin=0.5 -DrequestPersecond=1" \
  http://localhost:58080/task/upload/http
```

Running a a packaged simulation

In the [gatling-test-example](https://github.com/jecklgamis/gatling-test-example) project dir

```bash
curl -v \
  -F 'file=@target/gatling-test-example-lean.jar' \
  -F "simulation=gatling.test.example.simulation.ExamplePostSimulation" \
  -F "javaOpts=-DbaseUrl=http://localhost:8080 -DdurationMin=1 -DrequestPersecond=10" \
  http://localhost:58080/task/upload/http
```

### Submitting Task Using S3 Download

Gatling server can fetch and run simulations stored in a configured s3 bucket. Ensure the s3 download is configured in
the `config.yml` and that the container can access the s3 bucket.

```
downloaders:
  s3:
    enabled: false
    configMap:
      region: some-region
```

Example using S3 download:

```
$ curl -v -H "Content-Type:application/json" http://localhost:58080/task/download/s3 -d@request.json
```

request.json:

```
{
  "url": "s3://gatling-server-incoming/SingleFileExampleSimulation.scala",
  "simulation": "gatling.test.example.simulation.SingleFileExampleSimulation",
  "javaOpts": "-DbaseUrl=http://localhost:8080 -DdurationMin=0.10 -DrequestPerSecond=1"
}
```

### Aborting Task

To abort a task, send a POST request to the `/task/abort/{taskId}` endpoint:

Example:

```
$ curl -X POST http://localhost:58080/task/abort/e6a80550
```

## Generated Artifacts

A Gatling test run generates artifacts. This includes the console log, the Gatling reports, simulation logs, and
metadata of the original request. If the s3 uploader type is configured, these artifacts are uploaded to a configured
bucket name.

These artifacts are also available from the server itself.

Task Metadata:

```
http://localhost:5080/task/metadata/{taskId}
```

Console Output:

``` 
http://localhost:5080/task/console/{taskId}
```

Simulation Log

```
http://localhost:5080/task/simulationLog/{taskId}
```

Test Report:

```
http://localhost:5080/task/results/{taskId}
```

The report is a downloadable file in tar.gz format.

## Authoring Simulations

Gatling simulations are written in Scala. For simple simulations you can submit it straight away. For a
fairly complicated one, creating a build project (Maven for example) can make the experience less painful. Gatling is
quite flexible, it can be setup as a code alongside with your code base or maintained in a different repo.

See the following examples depending on the language you're using:

* [gatling-test-example](https://github.com/jecklgamis/gatling-test-example)
* [gatling-java-example](https://github.com/jecklgamis/gatling-java-example)
* [gatling-kotlin-example](https://github.com/jecklgamis/gatling-kotlin-example)

## Packaging Simulations In Jar Format

This is the recommended way of submitting packaged simulations.

`gatling-server` supports execution of simulations packaged as jar file. The jar should contain the compiled
simulations, resources, as well as class dependencies (that is, excluding Scala or Gatling dependencies). If you're
using Maven to author your simulations, this can be done using
the [maven-shade-plugin](https://maven.apache.org/plugins/maven-shade-plugin/). See
the [gatling-test-example] (https://github.com/jecklgamis/gatling-test-example) project as an example. It builds
`target/gatling-test-example-lean.jar` which you can submit to the server.

```bash
curl -v \
  -F 'file=@./target/gatling-test-example-lean.jar' \
  -F "simulation=gatling.test.example.simulation.ExamplePostSimulation" \
  -F "javaOpts=-DbaseUrl=http://172.16.0.50:8080 -DdurationMin=1 -DrequestPersecond=1" \
  http://localhost:58080/task/upload/http
```

## Packaging Simulations In tar.gz Format
There is no tooling on this at the moment, you can simply package your simulations into `tar.gz`. Ensure it
contains the following top level directories:
```
simulations # should contain simulation sources
resources #should contain feeder and data files
lib #$should contain external jar dependencies 
```

The `lib` directory should contain the external jar dependencies, if any. If you're using Maven, you can use the
dependency plugin to copy it to one location before archiving.


See [package-artifacts.sh](https://github.com/jecklgamis/gatling-test-example/blob/main/package-artifacts.sh) for reference script.