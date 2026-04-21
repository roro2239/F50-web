#!/system/bin/sh
set -eu

BASE_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
RUN_DIR="${BASE_DIR}/runtime"
PID_PATH="${RUN_DIR}/f50-web.pid"
WATCHDOG_PID_PATH="${RUN_DIR}/f50-web-watchdog.pid"
LAN_IP="192.168.0.1"

delete_nat_rule() {
    while iptables -t nat -D PREROUTING -p tcp -m tcp -i br0 -d "${LAN_IP}" --dport 80 -j DNAT --to-destination "${1}" 2>/dev/null; do
        :
    done
}

start_factory_web() {
    if pgrep -f com.zte.web >/dev/null 2>&1; then
        return 0
    fi

    monkey -p com.zte.web -c android.intent.category.LAUNCHER 1 >/dev/null 2>&1 || true
    sleep 2

    if pgrep -f com.zte.web >/dev/null 2>&1; then
        return 0
    fi

    ACTIVITY="$(cmd package resolve-activity --brief com.zte.web 2>/dev/null | tail -n 1)"
    case "${ACTIVITY}" in
        */*)
            am start -n "${ACTIVITY}" >/dev/null 2>&1 || true
            ;;
    esac
}

restore_factory_web() {
    delete_nat_rule "${LAN_IP}:8081"
    delete_nat_rule "${LAN_IP}:8080"
    iptables -t nat -I PREROUTING -p tcp -m tcp -i br0 -d "${LAN_IP}" --dport 80 -j DNAT --to-destination "${LAN_IP}:8080" 2>/dev/null || true
    start_factory_web
}

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

restore_factory_web

echo "f50-web 已停止"
