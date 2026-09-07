#!/usr/bin/env bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
PID_FILE=${SCRIPT_DIR}/server.pid

wait_for_exit() {
  local pid=$1
  for _ in $(seq 1 20); do
    kill -0 "${pid}" 2>/dev/null || return 0
    sleep 0.25
  done
  return 1
}

stop_pid() {
  local pid=$1
  kill -0 "${pid}" 2>/dev/null || return 0
  echo "Stopping process id ${pid}"
  kill -TERM "${pid}" 2>/dev/null
  if ! wait_for_exit "${pid}"; then
    echo "Process ${pid} did not stop gracefully, killing it"
    kill -KILL "${pid}" 2>/dev/null
    wait_for_exit "${pid}" || echo "Warning: process ${pid} may still be running"
  fi
}

PID=""
if [ -r "${PID_FILE}" ]; then
  PID=$(cat "${PID_FILE}")
fi
if [ -z "${PID}" ] || ! kill -0 "${PID}" 2>/dev/null; then
  PID=$(pgrep -f "bin/gatling-server" | head -n1)
fi

if [ -n "${PID}" ]; then
  stop_pid "${PID}"
else
  echo "No running gatling-server process found"
fi
rm -f "${PID_FILE}"

# Best-effort cleanup of Gatling simulation processes left running after the
# parent server was stopped (they are spawned via exec.Command and are not
# in the server's process group, so killing the server alone leaves them
# orphaned).
for jpid in $(pgrep -f "io.gatling.app.Gatling" 2>/dev/null); do
  echo "Killing orphaned Gatling process ${jpid}"
  kill -TERM "${jpid}" 2>/dev/null
done

exit 0
