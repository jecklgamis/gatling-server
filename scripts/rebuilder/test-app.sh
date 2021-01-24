#!/usr/bin/env bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
APP_DIR=${SCRIPT_DIR}/../..
source ${SCRIPT_DIR}/common.sh

cd ${APP_DIR} && go test -short ./... || (echo "Test failed" && speak "Test failed" && exit 1)

