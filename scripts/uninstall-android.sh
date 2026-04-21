#!/system/bin/sh
set -eu

BASE_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
BOOT_SCRIPT="/sdcard/ufi_tools_boot.sh"
BOOT_MARK="# F50-WEB-UFI-TOOLS"

if [ -x "${BASE_DIR}/scripts/stop-android.sh" ]; then
    sh "${BASE_DIR}/scripts/stop-android.sh" || true
fi

if [ -f "${BOOT_SCRIPT}" ]; then
    sed -i "\|${BOOT_MARK}|d" "${BOOT_SCRIPT}" 2>/dev/null || true
fi

cd /data 2>/dev/null || cd /
rm -rf "${BASE_DIR}"

echo "f50-web 已卸载"
