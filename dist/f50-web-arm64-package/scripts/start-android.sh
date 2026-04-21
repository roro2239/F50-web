#!/system/bin/sh
set -eu

BASE_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
BIN_PATH="${BASE_DIR}/f50-web"
RUN_DIR="${BASE_DIR}/runtime"
LOG_PATH="${RUN_DIR}/f50-web.log"
PID_PATH="${RUN_DIR}/f50-web.pid"

mkdir -p "${RUN_DIR}" "${BASE_DIR}/bin"
chmod 755 "${BIN_PATH}"

if [ -f "${PID_PATH}" ]; then
    OLD_PID="$(cat "${PID_PATH}")"
    if [ -n "${OLD_PID}" ] && kill -0 "${OLD_PID}" 2>/dev/null; then
        echo "f50-web 已在运行: ${OLD_PID}"
        exit 0
    fi
fi

cd "${BASE_DIR}"
"${BIN_PATH}" >> "${LOG_PATH}" 2>&1 &
echo $! > "${PID_PATH}"
echo "f50-web 已启动: $(cat "${PID_PATH}")"
