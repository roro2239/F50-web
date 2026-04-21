//<script>
(() => {
  const CARD_ID = 'f50_web_plugin_card';
  const BASE_DIR = '/data/local/tmp/f50-web';
  const BIN_PATH = `${BASE_DIR}/f50-web-arm64`;
  const START_SCRIPT = `${BASE_DIR}/scripts/start-android.sh`;
  const STOP_SCRIPT = `${BASE_DIR}/scripts/stop-android.sh`;
  const INSTALL_SCRIPT = '/data/local/tmp/f50-web-install.sh';
  const INSTALL_LOG = `${BASE_DIR}/install.log`;
  const BOOT_SCRIPT = '/sdcard/ufi_tools_boot.sh';
  const BOOT_MARK = '# F50-WEB-UFI-TOOLS';
  const BOOT_LINE = `cd ${BASE_DIR} && sh ${START_SCRIPT} ${BOOT_MARK}`;
  const LATEST_API =
    'https://gitee.com/api/v5/repos/su-su2239/F50-web/releases/latest';
  const RELEASES_URL = 'https://gitee.com/su-su2239/F50-web/releases';
  const PACKAGE_NAME = 'f50-web-arm64-package.zip';

  if (document.getElementById(CARD_ID)) return;

  const escapeHtml = (value = '') =>
    String(value)
      .replaceAll('&', '&amp;')
      .replaceAll('<', '&lt;')
      .replaceAll('>', '&gt;')
      .replaceAll('"', '&quot;')
      .replaceAll("'", '&#39;');

  const toast = (msg, color = '') => {
    if (typeof createToast === 'function') createToast(msg, color);
    else console.log(msg);
  };

  const checkRoot = async () => {
    const res = await runShellWithRoot('whoami', 10000);
    return res.success && (res.content || '').includes('root');
  };

  const getWebUrl = () => {
    const ip = window.UFI_DATA?.lan_ipaddr || location.hostname;
    return `http://${ip}:8081/`;
  };

  const fetchJson = async (url) => {
    const direct = await fetch(url, { cache: 'no-store' }).catch(() => null);
    if (direct && direct.ok) return direct.json();

    const proxy = await fetch(`${KANO_baseURL}/proxy/--${url}`, {
      cache: 'no-store',
    }).catch(() => null);
    if (proxy && proxy.ok) return proxy.json();
    throw new Error('无法获取最新版本信息');
  };

  const getLatestPackage = async () => {
    const release = await fetchJson(LATEST_API);
    const assets = Array.isArray(release.assets) ? release.assets : [];
    const exact = assets.find((item) => item.name === PACKAGE_NAME);
    const fallback = assets.find(
      (item) =>
        item.name &&
        item.name.includes('f50-web-arm64-package') &&
        item.name.endsWith('.zip'),
    );
    const asset = exact || fallback;
    if (!asset?.browser_download_url) {
      throw new Error(`最新版没有找到 ${PACKAGE_NAME}`);
    }
    return {
      tag: release.tag_name || release.name || 'latest',
      url: asset.browser_download_url,
    };
  };

  const readInstallLog = async () => {
    const res = await runShellWithRoot(
      `if [ -f "${INSTALL_LOG}" ]; then tail -n 30 "${INSTALL_LOG}"; else echo "等待安装日志生成"; fi`,
      10000,
    );
    return res.content || '';
  };

  const refreshStatus = async () => {
    const res = await runShellWithRoot(
      `
INSTALLED=0
RUNNING=0
HTTP_OK=0
BOOT_OK=0
[ -x "${BIN_PATH}" ] && INSTALLED=1
if [ -f "${BASE_DIR}/runtime/f50-web.pid" ]; then
  PID="$(cat "${BASE_DIR}/runtime/f50-web.pid" 2>/dev/null)"
  [ -n "$PID" ] && kill -0 "$PID" 2>/dev/null && RUNNING=1
fi
[ "$RUNNING" = "1" ] || pgrep -f "${BIN_PATH}" >/dev/null 2>&1 && RUNNING=1
if command -v toybox >/dev/null 2>&1; then
  printf 'GET / HTTP/1.1\\r\\nHost: 127.0.0.1\\r\\nConnection: close\\r\\n\\r\\n' | timeout 3 toybox nc 127.0.0.1 8081 2>/dev/null | head -n 1 | grep -q 'HTTP/' && HTTP_OK=1
fi
grep -q "${BOOT_MARK}" "${BOOT_SCRIPT}" 2>/dev/null && BOOT_OK=1
echo "installed=$INSTALLED"
echo "running=$RUNNING"
echo "http=$HTTP_OK"
echo "boot=$BOOT_OK"
      `,
      10000,
    );

    const text = res.content || '';
    const data = Object.fromEntries(
      text
        .split('\n')
        .map((line) => line.trim().split('='))
        .filter((item) => item.length === 2),
    );
    const installed = data.installed === '1';
    const running = data.running === '1';
    const httpOk = data.http === '1';
    const bootOk = data.boot === '1';

    const title = document.querySelector('#f50_web_status_title');
    const sub = document.querySelector('#f50_web_status_sub');
    const dot = document.querySelector('#f50_web_status_dot');
    const bootBtn = document.querySelector('#f50_web_boot_btn');

    if (title) {
      if (!installed) title.textContent = 'F50-web - 未安装';
      else if (running && httpOk) title.textContent = 'F50-web - 运行中';
      else if (running) title.textContent = 'F50-web - 进程运行，网页未响应';
      else title.textContent = 'F50-web - 已停止';
    }
    if (sub) {
      sub.textContent = `安装:${installed ? '是' : '否'}  进程:${
        running ? '是' : '否'
      }  网页:${httpOk ? '可访问' : '不可访问'}  自启:${
        bootOk ? '已开启' : '未开启'
      }`;
    }
    if (dot) {
      dot.style.background = running && httpOk ? '#16a34a' : running ? '#f59e0b' : '#dc2626';
    }
    if (bootBtn) {
      bootBtn.textContent = bootOk ? '关闭自启' : '开机自启';
      bootBtn.style.background = bootOk ? 'var(--dark-btn-color-active)' : '';
    }
    return { installed, running, httpOk, bootOk };
  };

  const buildInstallScript = (downloadUrl, tag) =>
    [
      '#!/system/bin/sh',
      'set -u',
      `DOWNLOAD_URL="${downloadUrl}"`,
      `TAG_NAME="${tag}"`,
      `BASE_DIR="${BASE_DIR}"`,
      'TMP_DIR="/data/local/tmp/f50-web-new"',
      'PKG="/data/local/tmp/f50-web-arm64-package.zip"',
      `LOG_FILE="${INSTALL_LOG}"`,
      'CURL_BIN="/data/data/com.minikano.f50_sms/files/curl"',
      'if [ -e "$BASE_DIR" ] && [ ! -d "$BASE_DIR" ]; then',
      '  mv "$BASE_DIR" "${BASE_DIR}.legacy.$(date +%s)" || fail "旧文件占用安装目录"',
      'fi',
      'mkdir -p "$BASE_DIR"',
      ': > "$LOG_FILE"',
      'log(){ printf "[%s] %s\\n" "$(date +%H:%M:%S)" "$1" >> "$LOG_FILE"; }',
      'fail(){ log "[失败] $1"; echo "INSTALL_FAILED" >> "$LOG_FILE"; exit 1; }',
      'log "[信息] 开始安装 F50-web ${TAG_NAME}"',
      '[ -x "$CURL_BIN" ] || CURL_BIN="curl"',
      'command -v "$CURL_BIN" >/dev/null 2>&1 || fail "未找到 curl"',
      'command -v unzip >/dev/null 2>&1 || fail "未找到 unzip"',
      'if [ -x "$BASE_DIR/scripts/stop-android.sh" ]; then sh "$BASE_DIR/scripts/stop-android.sh" >> "$LOG_FILE" 2>&1 || true; fi',
      'rm -rf "$TMP_DIR"',
      'mkdir -p "$TMP_DIR"',
      'log "[信息] 下载发布包"',
      '"$CURL_BIN" -L --fail --connect-timeout 30 --retry 2 -o "$PKG" "$DOWNLOAD_URL" >> "$LOG_FILE" 2>&1 || fail "下载失败"',
      'log "[信息] 解压发布包"',
      'unzip -o "$PKG" -d "$TMP_DIR" >> "$LOG_FILE" 2>&1 || fail "解压失败"',
      'BIN_FOUND="$(find "$TMP_DIR" -type f -name f50-web-arm64 2>/dev/null | head -n 1)"',
      '[ -n "$BIN_FOUND" ] || fail "发布包内没有 f50-web-arm64"',
      'ROOT_DIR="$(dirname "$BIN_FOUND")"',
      'SCRIPT_DIR="$ROOT_DIR/scripts"',
      'if [ ! -d "$SCRIPT_DIR" ]; then',
      '  START_FOUND="$(find "$TMP_DIR" -type f -name start-android.sh 2>/dev/null | head -n 1)"',
      '  [ -n "$START_FOUND" ] || fail "发布包内没有启动脚本"',
      '  SCRIPT_DIR="$(dirname "$START_FOUND")"',
      'fi',
      'log "[信息] 更新程序文件"',
      'rm -rf "$BASE_DIR/scripts"',
      'cp "$BIN_FOUND" "$BASE_DIR/f50-web-arm64" || fail "复制二进制失败"',
      'mkdir -p "$BASE_DIR/scripts"',
      'cp -R "$SCRIPT_DIR"/. "$BASE_DIR/scripts/" || fail "复制脚本失败"',
      'chmod 755 "$BASE_DIR/f50-web-arm64" "$BASE_DIR/scripts/"*.sh 2>/dev/null || true',
      'log "[信息] 启动后台"',
      'cd "$BASE_DIR" || fail "进入安装目录失败"',
      'sh "$BASE_DIR/scripts/start-android.sh" >> "$LOG_FILE" 2>&1 || fail "启动失败"',
      'sleep 3',
      'pgrep -f "$BASE_DIR/f50-web-arm64" >/dev/null 2>&1 || fail "启动后没有检测到进程"',
      'rm -rf "$TMP_DIR" "$PKG"',
      'log "[完成] 安装完成"',
      'echo "INSTALL_DONE" >> "$LOG_FILE"',
    ].join('\n');

  const startBackgroundInstall = async (downloadUrl, tag) => {
    const script = buildInstallScript(downloadUrl, tag);
    const command = [
      `cat > "${INSTALL_SCRIPT}" <<'F50_WEB_INSTALL_EOF'`,
      script,
      'F50_WEB_INSTALL_EOF',
      `chmod 755 "${INSTALL_SCRIPT}"`,
      `nohup sh "${INSTALL_SCRIPT}" >/dev/null 2>&1 &`,
      'echo INSTALL_STARTED',
    ].join('\n');

    const res = await runShellWithRoot(command, 100000);
    if (!res.success || !(res.content || '').includes('INSTALL_STARTED')) {
      throw new Error(res.content || '启动安装任务失败');
    }
  };

  const showInstallLog = () => {
    const fixed =
      typeof createFixedToast === 'function'
        ? createFixedToast(
            'f50_web_install_log',
            '<pre style="white-space:pre-wrap;min-width:300px;max-width:80vw;max-height:60vh;overflow:auto;text-align:left;">等待安装日志生成</pre>',
            '',
          )
        : null;
    const box = fixed?.el?.querySelector('pre') || fixed?.el;
    let count = 0;
    const timer = setInterval(async () => {
      count += 1;
      const log = await readInstallLog();
      if (box) box.innerHTML = escapeHtml(log);
      if (log.includes('INSTALL_DONE') || log.includes('INSTALL_FAILED') || count > 900) {
        clearInterval(timer);
        setTimeout(() => fixed?.close?.(), 3000);
        refreshStatus();
      }
    }, 1500);
  };

  const installOnline = async () => {
    if (!(await checkRoot())) return toast('没有开启高级功能，无法使用', 'red');
    const latest = await getLatestPackage();
    toast(`开始安装 ${latest.tag}`, 'pink');
    await startBackgroundInstall(latest.url, latest.tag);
    showInstallLog();
  };

  const stopService = async () => {
    if (!(await checkRoot())) return toast('没有开启高级功能，无法使用', 'red');
    const res = await runShellWithRoot(
      `
if [ -x "${STOP_SCRIPT}" ]; then
  cd "${BASE_DIR}" && sh "${STOP_SCRIPT}"
else
  pkill -f "${BIN_PATH}" 2>/dev/null || true
  pkill -f "f50-web-watchdog.sh" 2>/dev/null || true
fi
echo STOP_DONE
      `,
      30000,
    );
    toast(res.success ? '已发送停止命令' : `停止失败：${res.content}`, res.success ? 'green' : 'red');
    setTimeout(refreshStatus, 1500);
  };

  const restartService = async () => {
    if (!(await checkRoot())) return toast('没有开启高级功能，无法使用', 'red');
    const res = await runShellWithRoot(
      `
if [ -x "${STOP_SCRIPT}" ]; then cd "${BASE_DIR}" && sh "${STOP_SCRIPT}" || true; fi
sleep 1
if [ -x "${START_SCRIPT}" ]; then
  cd "${BASE_DIR}" && sh "${START_SCRIPT}"
else
  echo "未找到启动脚本"
  exit 1
fi
echo RESTART_DONE
      `,
      30000,
    );
    toast(res.success ? '已发送重启命令' : `重启失败：${res.content}`, res.success ? 'green' : 'red');
    setTimeout(refreshStatus, 2500);
  };

  const toggleBoot = async () => {
    if (!(await checkRoot())) return toast('没有开启高级功能，无法使用', 'red');
    const status = await refreshStatus();
    const command = status.bootOk
      ? `sed -i '/${BOOT_MARK}/d' "${BOOT_SCRIPT}" && echo BOOT_OFF`
      : `
if [ ! -f "${BOOT_SCRIPT}" ]; then
  printf '%s\\n' '#!/system/bin/sh' > "${BOOT_SCRIPT}"
fi
grep -qxF '${BOOT_LINE}' "${BOOT_SCRIPT}" || echo '${BOOT_LINE}' >> "${BOOT_SCRIPT}"
chmod 755 "${BOOT_SCRIPT}" 2>/dev/null || true
echo BOOT_ON
`;
    const res = await runShellWithRoot(command, 10000);
    toast(res.success ? '自启状态已更新' : `自启设置失败：${res.content}`, res.success ? 'green' : 'red');
    setTimeout(refreshStatus, 500);
  };

  const openWeb = () => {
    window.open(getWebUrl(), '_blank');
  };

  const setupCard = async () => {
    const wait = (ms = 100) =>
      new Promise((resolve) => {
        setTimeout(resolve, ms);
      });

    while (!window.UFI_DATA?.lan_ipaddr) {
      await wait();
    }

    const host = document.querySelector('.functions-container');
    if (!host || document.getElementById(CARD_ID)) return;

    host.insertAdjacentHTML(
      'afterend',
      `
<div id="${CARD_ID}" style="width: 100%; margin-top: 10px;">
    <div class="title" style="margin: 6px 0;">
        <span id="f50_web_status_dot" style="width:10px;height:10px;border-radius:50%;background:#6b7280;display:inline-block;margin-right:6px;"></span>
        <strong id="f50_web_status_title">F50-web后台</strong>
        <div style="display: inline-block;" id="collapse_f50_web_btn"></div>
    </div>
    <div class="collapse" id="collapse_f50_web" data-name="close" style="height: 0px; overflow: hidden;">
        <div class="collapse_box">
            <div id="f50_web_status_sub" style="font-size:12px;opacity:.8;margin-bottom:8px;">正在读取状态</div>
            <div id="f50_web_action_box" style="margin-bottom:10px;display:flex;gap:10px;flex-wrap:wrap">
                <button class="btn" id="f50_web_install_btn">安装</button>
                <button class="btn" id="f50_web_stop_btn">停止</button>
                <button class="btn" id="f50_web_restart_btn">重启</button>
                <button class="btn" id="f50_web_boot_btn">开机自启</button>
                <button class="btn" id="f50_web_open_btn">打开网页</button>
                <button class="btn" id="f50_web_release_btn">发布页</button>
            </div>
        </div>
    </div>
</div>
`,
    );

    document.querySelector('#f50_web_install_btn').onclick = async () => {
      try {
        await installOnline();
      } catch (e) {
        toast(`安装失败：${e.message || e}`, 'red');
      }
    };
    document.querySelector('#f50_web_stop_btn').onclick = stopService;
    document.querySelector('#f50_web_restart_btn').onclick = restartService;
    document.querySelector('#f50_web_boot_btn').onclick = toggleBoot;
    document.querySelector('#f50_web_open_btn').onclick = openWeb;
    document.querySelector('#f50_web_release_btn').onclick = () => {
      window.open(RELEASES_URL, '_blank');
    };

    collapseGen('#collapse_f50_web_btn', '#collapse_f50_web', '#collapse_f50_web', () => {
      refreshStatus();
    });
    refreshStatus();
    setInterval(refreshStatus, 10000);
  };

  setupCard();
})();
//</script>
