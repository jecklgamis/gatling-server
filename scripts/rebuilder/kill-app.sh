#!/usr/bin/env bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"

PID_FILE=${SCRIPT_DIR}/server.pid
if [ -r "${PID_FILE}" ]; then
  PID=$(cat ${PID_FILE})
  echo "Killing process id ${PID}"
  kill -9 ${PID} >/dev/null 2>&1
  rm -f ${PID_FILE}
fi

PID=$(ps -ef | grep "bin/gatling-server" | grep -v "grep" | awk '{ print $2 }')
if [ ! -z "${PID}" ]; then
  echo "It seems process is still running, killing it anyway"
  kill -9 ${PID} >/dev/null 2>&1
  rm -f ${PID_FILE}
fi
exit 0
