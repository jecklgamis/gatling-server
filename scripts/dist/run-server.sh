#!/usr/bin/env bash

APP_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
cd "${APP_DIR}"
export APP_ENVIRONMENT=prod
./gatling-server
