#!/usr/bin/env bash
set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
BRANCH=$(git rev-parse --abbrev-ref HEAD)

if [ "${BRANCH}" != "main" ]; then
  echo "WARNING: you are not in the main branch!"
fi

source ${SCRIPT_DIR}/release-version
git tag -a ${TAG} -m "${TAG}" && echo "Created tag ${TAG}"
