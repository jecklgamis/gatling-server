#!/usr/bin/env bash
set -euo pipefail
OUT_DIR=${1:-.}
mkdir -p "${OUT_DIR}"
SERVER_KEY="${OUT_DIR}/server.key"
SERVER_CERT="${OUT_DIR}/server.crt"
rm -f ${SERVER_KEY}
rm -f ${SERVER_CERT}
SUBJECT="/C=AU/ST=NSW/L=Sydney/O=Org/OU=OrgUnit/CN=gatling-server"
openssl req -x509 -nodes -days 365 -newkey rsa:4096 -keyout ${SERVER_KEY} -out ${SERVER_CERT} -subj ${SUBJECT} >/dev/null
openssl x509 -in ${SERVER_CERT} -text -noout >/dev/null
echo "Wrote ${SERVER_KEY}"
echo "Wrote ${SERVER_CERT}"
