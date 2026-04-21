#!/system/bin/sh
set -eu

BASE_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
CONFIG_FILE="${BASE_DIR}/F50-web.conf"
BOOT_SCRIPT="/sdcard/ufi_tools_boot.sh"
BOOT_MARK="# F50-WEB-UFI-TOOLS"
PLUGIN_STORE="/data/data/com.minikano.f50_sms/shared_prefs/kano_plugin_store.xml"
PLUGIN_NAME="f50-web-ufi-tools-plugin.js"
LAN_IP="192.168.0.1"
ADB_PORT="5555"

if [ -f "${CONFIG_FILE}" ]; then
    ADB_PORT="$(grep -E '^adbPort\s*=' "${CONFIG_FILE}" 2>/dev/null | tail -n 1 | sed 's/.*=//; s/[[:space:]]//g')"
    [ -n "${ADB_PORT}" ] || ADB_PORT="5555"
fi

delete_nat_rule() {
    while iptables -t nat -D PREROUTING -p tcp -m tcp -i br0 -d "${LAN_IP}" --dport "${1}" -j DNAT --to-destination "${2}" 2>/dev/null; do
        :
    done
}

delete_input_rule() {
    while iptables -D INPUT -p tcp -s 127.0.0.1 --dport 8080 -j ACCEPT 2>/dev/null; do
        :
    done
}

delete_ip6_rule() {
    while ip6tables -D INPUT -p tcp --dport "${1}" -j DROP 2>/dev/null; do
        :
    done
}

cleanup_network_rules() {
    delete_nat_rule 80 "${LAN_IP}:8081"
    delete_nat_rule 80 "${LAN_IP}:8080"
    delete_nat_rule 443 "${LAN_IP}:8443"
    iptables -t nat -I PREROUTING -p tcp -m tcp -i br0 -d "${LAN_IP}" --dport 80 -j DNAT --to-destination "${LAN_IP}:8080" 2>/dev/null || true
    iptables -t nat -I PREROUTING -p tcp -m tcp -i br0 -d "${LAN_IP}" --dport 443 -j DNAT --to-destination "${LAN_IP}:8443" 2>/dev/null || true
    delete_input_rule
    delete_ip6_rule 8080
    delete_ip6_rule "${ADB_PORT}"
    delete_ip6_rule 445
}

restore_system_settings() {
    cmd device_config set_sync_disabled_for_tests none >/dev/null 2>&1 || true
    cmd device_config delete activity_manager max_phantom_processes >/dev/null 2>&1 || true
    settings delete global settings_enable_monitor_phantom_procs >/dev/null 2>&1 || true
    setprop service.adb.tcp.port -1
    stop adbd >/dev/null 2>&1 || true
    start adbd >/dev/null 2>&1 || true
}

cleanup_files() {
    if [ -f "${BOOT_SCRIPT}" ]; then
        sed -i "\|${BOOT_MARK}|d" "${BOOT_SCRIPT}" 2>/dev/null || true
        grep -q '[^[:space:]]' "${BOOT_SCRIPT}" 2>/dev/null || rm -f "${BOOT_SCRIPT}" 2>/dev/null || true
    fi
    rm -f /sdcard/Download/f50-web-arm64-new 2>/dev/null || true
    rm -f /sdcard/Download/start-android.sh 2>/dev/null || true
    rm -f /sdcard/Download/stop-android.sh 2>/dev/null || true
    rm -f /sdcard/Download/uninstall-android.sh 2>/dev/null || true
}

cleanup_plugin_store() {
    [ -f "${PLUGIN_STORE}" ] || return 0
    TMP_STORE="${PLUGIN_STORE}.tmp.$$"
    START_MARK="\\[KANO_PLUGIN_START\\] ${PLUGIN_NAME}"
    END_MARK="\\[KANO_PLUGIN_END\\] ${PLUGIN_NAME}"
    if ! grep -q "${START_MARK}" "${PLUGIN_STORE}" 2>/dev/null; then
        return 0
    fi

    OWNER_GROUP="$(stat -c '%u:%g' "${PLUGIN_STORE}" 2>/dev/null || echo '')"
    FILE_MODE="$(stat -c '%a' "${PLUGIN_STORE}" 2>/dev/null || echo '')"
    sed "/${START_MARK}/,/${END_MARK}/d" "${PLUGIN_STORE}" > "${TMP_STORE}" 2>/dev/null || {
        rm -f "${TMP_STORE}" 2>/dev/null || true
        return 0
    }
    cat "${TMP_STORE}" > "${PLUGIN_STORE}" 2>/dev/null || true
    rm -f "${TMP_STORE}" 2>/dev/null || true
    [ -n "${OWNER_GROUP}" ] && chown "${OWNER_GROUP}" "${PLUGIN_STORE}" 2>/dev/null || true
    [ -n "${FILE_MODE}" ] && chmod "${FILE_MODE}" "${PLUGIN_STORE}" 2>/dev/null || true
}

if [ -x "${BASE_DIR}/scripts/stop-android.sh" ]; then
    sh "${BASE_DIR}/scripts/stop-android.sh" || true
fi

cleanup_network_rules
restore_system_settings
cleanup_files
cleanup_plugin_store

cd /data 2>/dev/null || cd /
rm -rf "${BASE_DIR}"

echo "f50-web 已卸载并清理残留"
