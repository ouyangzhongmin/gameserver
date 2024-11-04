package master

import (
	"fmt"
	"github.com/lonng/nano"
	"github.com/lonng/nano/serialize/json"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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
	)
}
