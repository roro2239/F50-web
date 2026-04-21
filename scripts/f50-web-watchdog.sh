#!/system/bin/sh
set -eu

BASE_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
BIN_PATH="${F50_WEB_BIN:-${BASE_DIR}/f50-web-arm64}"
if [ ! -x "${BIN_PATH}" ]; then
    BIN_PATH="${BASE_DIR}/f50-web"
fi

RUN_DIR="${BASE_DIR}/runtime"
PID_PATH="${RUN_DIR}/f50-web.pid"
WATCHDOG_PID_PATH="${RUN_DIR}/f50-web-watchdog.pid"
LOG_PATH="${RUN_DIR}/f50-web.log"
WATCHDOG_LOG_PATH="${RUN_DIR}/f50-web-watchdog.log"
INTERVAL="${F50_WEB_KEEPALIVE_INTERVAL:-10}"
MAX_LOG_SIZE="${F50_WEB_MAX_LOG_SIZE:-1048576}"

mkdir -p "${RUN_DIR}" "${BASE_DIR}/bin"
chmod 755 "${BIN_PATH}"
echo $$ > "${WATCHDOG_PID_PATH}"

log_msg() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*" >> "${WATCHDOG_LOG_PATH}"
}

trim_log() {
    target="$1"
    if [ -f "${target}" ]; then
        size="$(wc -c < "${target}" 2>/dev/null || echo 0)"
        if [ "${size}" -gt "${MAX_LOG_SIZE}" ]; then
            tail -c "${MAX_LOG_SIZE}" "${target}" > "${target}.tmp" 2>/dev/null || true
            mv "${target}.tmp" "${target}" 2>/dev/null || true
        fi
    fi
}

is_running() {
    pid="$1"
    [ -n "${pid}" ] && kill -0 "${pid}" 2>/dev/null
}

start_app() {
    cd "${BASE_DIR}"
    "${BIN_PATH}" >> "${LOG_PATH}" 2>&1 &
    app_pid="$!"
    echo "${app_pid}" > "${PID_PATH}"
    log_msg "f50-web 已启动: ${app_pid}"
}

stop_children() {
    if [ -f "${PID_PATH}" ]; then
        pid="$(cat "${PID_PATH}" 2>/dev/null || true)"
        if is_running "${pid}"; then
            kill "${pid}" 2>/dev/null || true
        fi
        rm -f "${PID_PATH}"
    fi
    rm -f "${WATCHDOG_PID_PATH}"
    log_msg "watchdog 已停止"
    exit 0
}

trap stop_children TERM INT
log_msg "watchdog 已启动"

while true; do
    pid=""
    if [ -f "${PID_PATH}" ]; then
        pid="$(cat "${PID_PATH}" 2>/dev/null || true)"
    fi

    if ! is_running "${pid}"; then
        start_app
    fi

    trim_log "${LOG_PATH}"
    trim_log "${WATCHDOG_LOG_PATH}"
    sleep "${INTERVAL}"
done
