package master

import (
	"fmt"
	"github.com/lonng/nano"
	"github.com/lonng/nano/cluster/clusterpb"
	"github.com/lonng/nano/serialize/json"
	"github.com/lonng/nano/session"
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
	)
}

// 集群模式下，需要获取用户所在的game node调用rpc
func customerRemoteServiceRoute(service string, session *session.Session, members []*clusterpb.MemberInfo) *clusterpb.MemberInfo {
	if strings.Contains(service, "SceneManager") {
		//根据用户id获取用户在哪个node上
		curSceneId := session.Value("sceneId").(int)
		if curSceneId > 0 {
			for _, m := range members {
				label := m.Label
				if label != "" {
					label = strings.ReplaceAll(label, "scene:", "")
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
	count := int64(len(members))
	var index = session.UID() % count
	fmt.Printf("remote service:%s route to :%v \n", service, members[index])
	return members[index]
}
