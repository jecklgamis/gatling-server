#!/bin/bash
set -e
cd /app
if [ ! -f server.key ] || [ ! -f server.crt ]; then
  scripts/generate-ssl-certs.sh
fi
bin/gatling-server
