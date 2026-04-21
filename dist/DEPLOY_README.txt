F50-web arm64 部署说明

目录内容：
- f50-web：主程序二进制，已内置前端资源
- scripts/start-android.sh：启动脚本
- scripts/stop-android.sh：停止脚本
- bin/：可选外部二进制目录，目前仅 ttyd 功能会从这里查找 bin/ttyd

推荐部署目录：
- /data/local/tmp/f50-web

部署示例：
1. 上传整个目录到设备
2. 执行：
   su -c 'cd /data/local/tmp/f50-web && chmod 755 f50-web scripts/start-android.sh scripts/stop-android.sh && ./scripts/start-android.sh'

停止命令：
   su -c 'cd /data/local/tmp/f50-web && ./scripts/stop-android.sh'

访问方式：
- 设备 IP:8081

说明：
- 主程序已包含前端页面，不需要再上传源码或 web 目录
- ddns-go 与 AList 已移除
- 如需 ttyd，请额外放置 arm64 的 bin/ttyd
