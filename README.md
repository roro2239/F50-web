F50-web 使用说明

项目地址：https://github.com/roro2239/F50-web

码云地址：https://gitee.com/su-su2239/F50-web

上游来源：https://gitee.com/atwat/F50-web/tree/master/web

本项目基于上述项目的 Web 资源与相关实现改造，调整为当前仓库维护的构建与运行方式。
 
一、适用设备与版本要求
 
- 设备型号：中兴F50
- 系统版本：需更新至B09及以上版本
 
二、新增功能列表
 
- ipv6防火墙：默认阻止ipv6访问8080，445端口，提示设备安全性
- adb网络端口：支持通过网络远程连接设备调试，具备鉴权机制：仅登录后台系统后，方可通过网络远程连接adb端口
- 系统监控：实时显示CPU、内存占用率及设备温度
- 自定义Samba文件管理：支持自定义Samba文件服务地址，可设置Samba访问密码，保障文件传输安全
- AT命令：便捷调用设备底层通信指令
- 短信转发：可将短信同步转发至指定目标（建议使用termux-sms-send或curl命令）
- ttyd：基于网页的终端访问（含root权限）

三、后台保活运行

推荐使用发布包 `f50-web-arm64-package.zip`，解压到 `/data/local/tmp/f50-web/` 后执行：

```sh
su -c 'cd /data/local/tmp/f50-web && chmod 755 f50-web-arm64 scripts/*.sh && ./scripts/start-android.sh'
```

停止服务：

```sh
su -c 'cd /data/local/tmp/f50-web && ./scripts/stop-android.sh'
```

保活脚本会检测主进程状态，主进程退出后会自动重新拉起。
