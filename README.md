# gameserver
基于nanoserver直接改造的的mmrpg server, 小部分代码雷同
* gate -- 网关服务器
* master- 集群master服务器, 处理用户选择创建等
* game - 游戏场景逻辑服务器, 可以多开，每个game对应1个或多个场景管理
* web 登录http服务
aoi下载于开源：https://github.com/knight0zh/aoi

目前问题:
```aiignore
1.hero 推送消息在同屏数据量大时消息会堵塞,nano作者在agent.push内会丢弃超过agentWriteBacklog(16)缓冲区的数据，我加了个循环尝试，但是同屏1000以上时还是会大量堵塞消息, 
另外消息发送量大后cpu也会增加，后续看看有没有合适的方案

2.game服务器重启entity数据都丢失了，无法做到重启用户无感知，可以参考goworld方式将entity全部序列化到文件，重启时载入
```
