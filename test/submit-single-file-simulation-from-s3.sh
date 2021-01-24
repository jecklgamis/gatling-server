#!/usr/bin/env bash
set -e
cat <<-'EOF' >request.json
{
  "url": "s3://gatling-server-incoming/SingleFileExampleSimulation.scala",
  "simulation": "gatling.test.example.simulation.SingleFileExampleSimulation",
  "javaOpts": "-DbaseUrl=http://localhost:8080 -DdurationMin=0.10 -DrequestPerSecond=1"
}
EOF
curl -v -H "Content-Type:application/json" http://localhost:58080/task/download/s3 -d@request.json
