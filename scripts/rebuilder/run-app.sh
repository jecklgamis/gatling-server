#!/usr/bin/env bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
APP_DIR=${SCRIPT_DIR}/../..

if [ -x "${APP_DIR}/bin/gatling-server" ]; then
  echo "Running app"
  ${APP_DIR}/bin/gatling-server &
  echo $! >${SCRIPT_DIR}/server.pid
fi
exit 0