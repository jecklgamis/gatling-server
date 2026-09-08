#!/usr/bin/env bash

BIN_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
APP_DIR="$(cd "${BIN_DIR}/.." >/dev/null 2>&1 && pwd)"
cd "${APP_DIR}"
export APP_ENVIRONMENT=prod
export SCRIPTS_DIR=bin
bin/gatling-server
