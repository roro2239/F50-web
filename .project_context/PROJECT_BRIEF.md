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

已完成从 `Termux` 运行时到“原生二进制 + 原生目录”模式的基础改造。当前代码已切换到统一运行目录、原生 `sh`，并已移除 `ddns-go`、`AList` 的前后端入口。本地构建产物已清理，新增 GitHub Actions release workflow 自动构建 `Android arm64` 二进制并用当前版本提交日志发布更新日志。README 已删除在线安装说明，更新项目地址、码云地址，并补充上游来源声明。手动触发 release workflow 时已支持输入版本号。新增 watchdog 脚本用于后台运行和进程级保活。新增 GitHub Release 同步到 Gitee Release 的 workflow。

# 下一步

推送 `v*` 标签，或手动触发 `构建并发布` workflow 并填写版本号，生成并发布 `f50-web-arm64` 与保活脚本。
