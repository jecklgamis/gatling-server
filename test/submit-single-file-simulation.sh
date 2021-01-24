#!/usr/bin/env bash
set -ex
curl -v \
  -F 'file=@./testdata/SingleFileExampleSimulation.scala' \
  -F "simulation=gatling.test.example.simulation.SingleFileExampleSimulation" \
  -F "javaOpts=-DbaseUrl=http://localhost:8080 -DdurationMin=0.10 -DrequestPersecond=1" \
  http://localhost:58080/task/upload/http
