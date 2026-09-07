#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
cd "${SCRIPT_DIR}/.."

if ! command -v golangci-lint &>/dev/null; then
  echo "golangci-lint not found. Install it: https://golangci-lint.run/welcome/install/" >&2
  exit 1
fi

golangci-lint run ./...
