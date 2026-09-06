#!/usr/bin/env bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
APP_DIR=${SCRIPT_DIR}/../..

if [ -x "${APP_DIR}/bin/gatling-server" ]; then
  cd "${APP_DIR}"
  if [ ! -f server.key ] || [ ! -f server.crt ]; then
    echo "Generating SSL certs"
    scripts/generate-ssl-certs.sh
  fi
  echo "Running app"
  ./bin/gatling-server &
  echo $! >${SCRIPT_DIR}/server.pid
fi
exit 0