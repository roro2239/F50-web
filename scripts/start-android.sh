#!/system/bin/sh
set -eu

BASE_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
RUN_DIR="${BASE_DIR}/runtime"
WATCHDOG_PATH="${BASE_DIR}/scripts/f50-web-watchdog.sh"
WATCHDOG_PID_PATH="${RUN_DIR}/f50-web-watchdog.pid"
WATCHDOG_LOG_PATH="${RUN_DIR}/f50-web-watchdog.log"
LAN_IP="192.168.0.1"

delete_nat_rule() {
    while iptables -t nat -D PREROUTING -p tcp -m tcp -i br0 -d "${LAN_IP}" --dport 80 -j DNAT --to-destination "${1}" 2>/dev/null; do
        :
    done
}

apply_takeover() {
    delete_nat_rule "${LAN_IP}:8080"
    delete_nat_rule "${LAN_IP}:8081"
    iptables -t nat -I PREROUTING -p tcp -m tcp -i br0 -d "${LAN_IP}" --dport 80 -j DNAT --to-destination "${LAN_IP}:8081" 2>/dev/null || true
    pkill -f com.zte.web 2>/dev/null || true
}

mkdir -p "${RUN_DIR}" "${BASE_DIR}/bin"
chmod 755 "${WATCHDOG_PATH}"
apply_takeover

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
