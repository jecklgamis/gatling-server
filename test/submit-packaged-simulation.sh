#!/usr/bin/env bash
set -ex
curl -v \
  -F 'file=@./testdata/gatling-test-example-user-files.tar.gz' \
  -F "simulation=gatling.test.example.simulation.ExampleSimulation" \
  -F "javaOpts=-DbaseUrl=http://localhost:8080 -DdurationMin=0.10 -DrequestPersecond=1" \
  http://localhost:58080/task/upload/http
