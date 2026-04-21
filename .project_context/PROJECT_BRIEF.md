# 项目概览

F50-web 是面向中兴 F50 设备的 Go 后台程序，负责接管原设备 Web 管理入口并扩展 ADB、AT 命令、短信转发、ttyd、iptables 管理等能力。项目地址为 https://github.com/roro2239/F50-web，码云地址为 https://gitee.com/su-su2239/F50-web。项目基于 https://gitee.com/atwat/F50-web/tree/master/web 的 Web 资源与相关实现改造。

# 技术栈

- Go
- 嵌入式静态资源 `embed`
- SQLite
- SMB
- HTTP / WebSocket

# 目录结构

- `main.go`：主入口
- `server.go`：HTTP 服务与 API
- `rootShell.go`：通过 Samba `root preexec` 执行 root 脚本
- `cmd.go`：外部命令执行器与附加服务启动
- `web/`：嵌入式管理页面资源

# 关键模块

- 登录与转发：`req.go`、`server.go`
- root 执行链路：`rootShell.go`
- ADB 代理：`adbSafe.go`
- 系统监控：`sys.go`
- 扩展服务：`cmd.go`

# 构建与运行

- 目标平台：Android `arm64`
- 已补齐 `go.mod`
- 需要 Android NDK 才能完成 `sqlite3` 的 CGO 交叉编译

# 当前任务状态

已完成从 `Termux` 运行时到“原生二进制 + 原生目录”模式的基础改造。当前代码已切换到统一运行目录、原生 `sh`，并已移除 `ddns-go`、`AList` 的前后端入口。本地构建产物已清理，新增 GitHub Actions release workflow 自动构建 `Android arm64` 二进制并用当前版本提交日志发布更新日志。README 已删除在线安装说明，更新项目地址、码云地址，并补充上游来源声明。手动触发 release workflow 时已支持输入版本号。新增 watchdog 脚本用于后台运行和进程级保活。新增 GitHub Release 同步到 Gitee Release 的 workflow。新增 `UFI-TOOLS` 插件脚本，用于从 Gitee Release 自动识别最新版发布包并安装、停止、重启、配置开机自启和打开后台网页；插件已按参考实现改为在 `.functions-container` 后插入独立卡片，并使用 `collapseGen` 管理展开收起，同时兼容 `/data/local/tmp/f50-web` 旧单文件占用安装目录的场景。新增原厂后台地址自动探测，启动时优先测试 `127.0.0.1:8080`，不通时切换到 `192.168.0.1:8080`，修复 F50 上 `8081/goform` 代理超时。登录链路已修正：当本地未持有真实登录密码时，不再拦截原厂登录请求，也不再定时发送空密码登录，避免错误消耗原厂后台的密码尝试次数。相关改动已在设备上完成部署验证。

# 下一步

根据后续需求继续收敛调试日志，并补齐发布包与设备自启链路的最终使用说明。
