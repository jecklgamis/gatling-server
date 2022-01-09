## gatling-server

![Go](https://github.com/jecklgamis/gatling-server/workflows/Go/badge.svg?branch=main) [![Go Report Card](https://goreportcard.com/badge/github.com/jecklgamis/gatling-server)](https://goreportcard.com/report/github.com/jecklgamis/gatling-server)

gatling-server is an API server for [Gatling](https://gatling.io/) OSS. 

Heads up: this is currently in alpha testing but feel free to try the latest Docker image from Docker Hub.

Security warning: The server can execute arbitrary code as Gatling simulations are normal Scala codes themselves.
Use at your own risk.

## Current Features
* Run single file or packaged simulations (simulations, resources, bodies packaged as tar.gz) 
* Task submission via HTTP upload or S3 download 
* Artifacts upload to S3 (metadata, console log, results, etc.)
* Endpoints for task metadata, console log, and results, etc.
* HTTP and SNS event notifiers (heartbeat and task cycle events)
* Docker image available in Docker Hub

## Getting Started 
1. Run the Docker image available from Docker Hub:
```
docker run -i -t -p 58080:58080 jecklgamis/gatling-server:latest
```
Ensure it's up by hitting the `/buildInfo` endpoint:
```
$ curl http://localhost:58080/buildInfo 
```

### Submitting Tasks
Gatling tasks are submitted via HTTP endpoints.  Artifacts referenced in a request can be a single file simulation or a 
packaged simulation. A single file simulation class is a single Scala file that uses only on standard Scala or Gatling 
libraries. Packaged simulations contains resources such as request bodies, feeders, or additional utility classes. 

A task identifier is returned in the task submission response. This can be used to query the server for generated
artifacts such as console logs or Gatling reports.

### Submitting Task Using HTTP Upload
1. Run a single file test simulation

Example:
```
$ cd test
curl -v \
  -F 'file=@./testdata/SingleFileExampleSimulation.scala' \
  -F "simulation=gatling.test.example.simulation.SingleFileExampleSimulation" \
  -F "javaOpts=-DbaseUrl=http://localhost:8080 -DdurationMin=0.5 -DrequestPersecond=1" \
  http://localhost:58080/task/upload/http
```
2. Run a packaged test simulation
```
curl -v \
  -F 'file=@./testdata/gatling-test-example-user-files.tar.gz' \
  -F "simulation=gatling.test.example.simulation.ExampleSimulation" \
  -F "javaOpts=-DbaseUrl=http://localhost:8080 -DdurationMin=0.10 -DrequestPersecond=1" \
  http://localhost:58080/task/upload/http
```

### Submitting Task Using S3 Download
Gatling server can fetch and run simulations stored in a configured s3 bucket. Ensure the s3 uploader is configured
in the `config.yml` and that the container can access the s3 bucket.
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
To abort a task,  send a POST request to the `/task/abort/{taskId}` endpoint:

Example:
```
$ curl -X POST http://localhost:58080/task/abort/e6a80550
```


## Generated Artifacts
A Gatling test run generates artifacts. This includes the console log, the Gatling reports, simulation logs,
and metadata of the original request. If the s3 uploader type is configured, these artifacts are uploaded to 
a configured bucket name. 

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
Gatling simulations are written in Scala. It is compiled on the fly and thus will result to  Gatling test run failure
if there  are compile errors. For simple simulations you can submit it straight away. For a fairly complicated one, creating a 
build project (Maven for example) can make the experience less painful. Gatling is quite flexible, it can be setup as a 
code along side with your code base or maintained in a different repo.

See [gatling-test-example](git@github.com:jecklgamis/gatling-test-example.git) for an example Maven project.

## Packaging Simulations
There is no tooling on this at the moment, you can simply package your simulations into `tar.gz`. Ensure it 
contains the following top level directories:
```
simulations (should contain simulation classes)
resources (should contain feeder and data files)
lib (should contain external jar dependencies) 
```

The `lib` directory should contain the external jar dependencies, if any. If you're using Maven, you can use the
dependency plugin to copy it to one location before archiving.  

```xml
       <plugin>
                <groupId>org.apache.maven.plugins</groupId>
                <artifactId>maven-dependency-plugin</artifactId>
                <version>3.2.0</version>
                <executions>
                    <execution>
                        <id>copy</id>
                        <phase>package</phase>
                        <goals>
                            <goal>copy-dependencies</goal>
                        </goals>
                        <configuration>
                            <includeGroupIds>org.json4s</includeGroupIds>
                        </configuration>
                    </execution>
                </executions>
            </plugin>
```


See `scripts/package-artifacts.sh` for reference script.

```
#!/usr/bin/env bash

## This utility packages simulations and resources into a tar.gz that can be fed into the gatling-server.
## See https://github.com/jecklgamis/gatling-server

CURRENT_DIR=$(pwd)

NAME=gatling-test-example
USER_FILES_DIR=${NAME}-user-files
TAR_GZ_FILE=${USER_FILES_DIR}.tar.gz

rm -rf ${USER_FILES_DIR}
mkdir -p ${USER_FILES_DIR}/simulations
mkdir -p ${USER_FILES_DIR}/bodies
mkdir -p ${USER_FILES_DIR}/resources
mkdir -p ${USER_FILES_DIR}/binaries

cp -rf src/main/scala/* ${USER_FILES_DIR}/simulations/
[[ $(ls -A "${USER_FILES_DIR}/bodies/") ]] && cp -rf src/main/resources/bodies/* ${USER_FILES_DIR}/bodies/
[[ $(ls -A "${USER_FILES_DIR}/resources/") ]] && cp -rf src/main/resources/data/* ${USER_FILES_DIR}/resources/
[[ $(ls -A "target/dependency") ]] && cp -rf target/dependency/* ${USER_FILES_DIR}/lib/

rm -f ${TAR_GZ_FILE}
tar cvzf ${TAR_GZ_FILE} ${USER_FILES_DIR}

echo "Created ${CURRENT_DIR}/${TAR_GZ_FILE}"
```


Example `gatling-test-example-user-files.tar.gz` contents
```
$ tar tf test/testdata/gatling-test-example-user-files.tar.gz 
gatling-test-example-user-files/
gatling-test-example-user-files/bodies/
gatling-test-example-user-files/resources/
gatling-test-example-user-files/simulations/
gatling-test-example-user-files/simulations/gatling/
gatling-test-example-user-files/simulations/gatling/test/
gatling-test-example-user-files/simulations/gatling/test/example/
gatling-test-example-user-files/simulations/gatling/test/example/simulation/
gatling-test-example-user-files/simulations/gatling/test/example/simulation/ExamplePostSimulation.scala
gatling-test-example-user-files/simulations/gatling/test/example/simulation/ExampleSimulation.scala
gatling-test-example-user-files/simulations/gatling/test/example/simulation/PerfTestConfig.scala
gatling-test-example-user-files/simulations/gatling/test/example/simulation/ExampleGetSimulation.scala
gatling-test-example-user-files/simulations/gatling/test/example/simulation/SystemPropertiesUtil.scala
```
