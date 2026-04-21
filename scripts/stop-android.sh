#!/system/bin/sh
set -eu

BASE_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
RUN_DIR="${BASE_DIR}/runtime"
PID_PATH="${RUN_DIR}/f50-web.pid"
WATCHDOG_PID_PATH="${RUN_DIR}/f50-web-watchdog.pid"

if [ -f "${WATCHDOG_PID_PATH}" ]; then
    WATCHDOG_PID="$(cat "${WATCHDOG_PID_PATH}" 2>/dev/null || true)"
    if [ -n "${WATCHDOG_PID}" ] && kill -0 "${WATCHDOG_PID}" 2>/dev/null; then
        kill "${WATCHDOG_PID}" 2>/dev/null || true
    fi
    rm -f "${WATCHDOG_PID_PATH}"
fi

if [ -f "${PID_PATH}" ]; then
    PID="$(cat "${PID_PATH}" 2>/dev/null || true)"
    if [ -n "${PID}" ] && kill -0 "${PID}" 2>/dev/null; then
        kill "${PID}" 2>/dev/null || true
    fi
    rm -f "${PID_PATH}"
fi

echo "f50-web 已停止"
