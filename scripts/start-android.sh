#!/system/bin/sh
set -eu

BASE_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
RUN_DIR="${BASE_DIR}/runtime"
WATCHDOG_PATH="${BASE_DIR}/scripts/f50-web-watchdog.sh"
WATCHDOG_PID_PATH="${RUN_DIR}/f50-web-watchdog.pid"
WATCHDOG_LOG_PATH="${RUN_DIR}/f50-web-watchdog.log"

mkdir -p "${RUN_DIR}" "${BASE_DIR}/bin"
chmod 755 "${WATCHDOG_PATH}"

if [ -f "${WATCHDOG_PID_PATH}" ]; then
    OLD_PID="$(cat "${WATCHDOG_PID_PATH}")"
    if [ -n "${OLD_PID}" ] && kill -0 "${OLD_PID}" 2>/dev/null; then
        echo "f50-web watchdog 已在运行: ${OLD_PID}"
        exit 0
    fi
fi

cd "${BASE_DIR}"
nohup "${WATCHDOG_PATH}" >> "${WATCHDOG_LOG_PATH}" 2>&1 &
echo $! > "${WATCHDOG_PID_PATH}"
echo "f50-web watchdog 已启动: $(cat "${WATCHDOG_PID_PATH}")"
