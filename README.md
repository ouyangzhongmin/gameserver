# gameserver
## 基于nanoserver直接改造的的mmrpg server, 小部分代码雷同
* gate -- 网关服务器
* master- 集群master服务器, 处理用户选择创建等
* game - 游戏场景逻辑服务器, 可以多开，每个game对应1个或多个场景管理
* web 登录http服务 + html client demo
* aoi下载于开源：https://github.com/knight0zh/aoi
* nanaserver: https://github.com/lonng/nanoserver
* nano: https://github.com/lonng/nano

## 启动
导入docs/jsmx.sql到mysql,修改configs/config.toml配置,然后分别运行cmd/master、gate、game、web start_server.sh启动所有服务，
内置html demo: http://localhost:12307/static/client/

## 目前问题:
```aiignore
1.hero 推送消息在同屏数据量大时消息会堵塞,nano作者在agent.push内会丢弃超过agentWriteBacklog(16)缓冲区的数据，我加了个消息合并发送的方式

2.game服务器重启entity数据都丢失了，无法做到重启用户无感知，可以参考goworld方式将entity全部序列化到文件，重启时载入

3.目前实现的视野刷新问题，如果同屏1万人在线，那么服务器的视野刷新处理计算非常高, 不适合大量同屏的处理
```
