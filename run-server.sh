#!/usr/bin/env bash
set -euo pipefail
BIN_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
cd "${BIN_DIR}"

if [ ! -f server.key ] || [ ! -f server.crt ]; then
  echo "No TLS cert found, generating one..."
  scripts/generate-ssl-certs.sh
fi

export APP_ENVIRONMENT=dev
export SCRIPTS_DIR=scripts
go run cmd/server/gatling-server.go
