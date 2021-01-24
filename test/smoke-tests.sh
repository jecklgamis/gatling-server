#!/usr/bin/env bash
set -e
BASE_URL="http://localhost:58080"
curl ${BASE_URL}/buildInfo
curl ${BASE_URL}/probe/ready
curl ${BASE_URL}/probe/live
curl ${BASE_URL}/api
curl ${BASE_URL}
