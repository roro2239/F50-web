#!/system/bin/sh
set -eu

BASE_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
PID_PATH="${BASE_DIR}/runtime/f50-web.pid"

if [ ! -f "${PID_PATH}" ]; then
    echo "未找到 pid 文件"
    exit 1
fi

PID="$(cat "${PID_PATH}")"
if [ -n "${PID}" ] && kill -0 "${PID}" 2>/dev/null; then
    kill "${PID}"
fi
rm -f "${PID_PATH}"
echo "f50-web 已停止"
