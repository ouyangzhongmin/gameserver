package master

import (
	"fmt"
	"github.com/ouyangzhongmin/nano"
	"github.com/ouyangzhongmin/nano/cluster"
	"github.com/ouyangzhongmin/nano/cluster/clusterpb"
	"github.com/ouyangzhongmin/nano/serialize/json"
	"github.com/ouyangzhongmin/nano/session"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"strings"
)

var (
	logger = log.WithField("component", "master")
)

// Startup 初始化master服务器
func Startup() {

	version := viper.GetString("update.version")
	heartbeat := viper.GetInt("core.heartbeat")
	if heartbeat < 5 {
		heartbeat = 5
	}
	forceUpdate := viper.GetBool("update.force")
	// register game handler
	comps := Services

	listen := fmt.Sprintf(":%d", viper.GetInt("master.port"))
	logger.Infof("当前master server服务器版本: %s, 是否强制更新: %t, 当前心跳时间间隔: %d秒", version, forceUpdate, heartbeat)
	logger.Info("master service starup:", listen)
	nano.Listen(listen,
		nano.WithMaster(),
		//nano.WithPipeline(pip),
		nano.WithLogger(log.WithField("component", "master")),
		nano.WithSerializer(json.NewSerializer()),
		nano.WithComponents(comps),
		nano.WithCustomerRemoteServiceRoute(customerRemoteServiceRoute),
		nano.WithUnregisterCallback(nodeUnregisterCallback),
	)
}

// 集群模式下，需要获取用户所在的game node调用rpc
func customerRemoteServiceRoute(service string, session *session.Session, members []*clusterpb.MemberInfo) *clusterpb.MemberInfo {
	if strings.Contains(service, "SceneManager") {
		//根据用户id获取用户在哪个node上
		if session.String("remoteAddr") != "" {
			//根据用户id获取用户在哪个node上
			for _, m := range members {
				if session.String("remoteAddr") == m.ServiceAddr {
					return m
				}
			}
		} else {
			curSceneId := session.Int("sceneId")
			if curSceneId > 0 {
				for _, m := range members {
					if m.Label != "" {
						label := strings.ReplaceAll(m.Label, "scene:", "")
						tmpArr := strings.Split(label, ",")
						for _, tmp := range tmpArr {
							if tmp == fmt.Sprintf("%d", curSceneId) {
								return m
							}
						}
					}
				}
			}
		}
	}
	count := int64(len(members))
	var index = session.UID() % count
	fmt.Printf("remote service:%s route to :%v \n", service, members[index])
	return members[index]
}

// 有节点关闭掉了
func nodeUnregisterCallback(member cluster.Member) {
	cellManager.updateCellWithMemberShutdown(member.MemberInfo().ServiceAddr)
}
