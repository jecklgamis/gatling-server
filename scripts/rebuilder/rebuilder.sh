#!/usr/bin/env bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
APP_DIR=${SCRIPT_DIR}/../..

function check_binaries() {
  if ! command -v go &>/dev/null; then
    echo "go binary not found!"
    exit -1
  fi
  if ! command -v fswatch &>/dev/null; then
    echo "fswatch binary not found!"
    exit -1
  fi
}

function sigint_handler() {
  ${SCRIPT_DIR}/kill-app.sh
}

trap 'sigint_handler' SIGINT

check_binaries
${SCRIPT_DIR}/test-app.sh && ${SCRIPT_DIR}/build-app.sh && ${SCRIPT_DIR}/kill-app.sh && ${SCRIPT_DIR}/run-app.sh
fswatch -o ${APP_DIR}/pkg -o ${APP_DIR}/cmd -o ${APP_DIR}/internal | xargs -n1 -I{} sh -c "${SCRIPT_DIR}/test-app.sh && ${SCRIPT_DIR}/build-app.sh && ${SCRIPT_DIR}/kill-app.sh && ${SCRIPT_DIR}/run-app.sh"
