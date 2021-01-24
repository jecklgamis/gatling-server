#!/usr/bin/env bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
APP_DIR=${SCRIPT_DIR}/../..
source ${SCRIPT_DIR}/common.sh

echo "Building app"
cd ${APP_DIR} && (go build -o bin/gatling-server cmd/server/gatling-server.go \
 && chmod +x bin/gatling-server) || (echo "Build failed" && speak "Build failed" && exit 1)
exit 0
