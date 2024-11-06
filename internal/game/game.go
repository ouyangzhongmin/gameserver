package game

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/lonng/nano"
	"github.com/lonng/nano/component"
	"github.com/lonng/nano/serialize/json"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	version     = "" // 游戏版本
	forceUpdate = false
	logger      = log.WithField("component", "game")
)

// Startup 初始化游戏服务器
func Startup() {
	rand.Seed(time.Now().Unix())
	version = viper.GetString("update.version")

	heartbeat := viper.GetInt("core.heartbeat")
	if heartbeat < 5 {
		heartbeat = 5
	}

	forceUpdate = viper.GetBool("update.force")
	// register game handler
	comps := &component.Components{}
	comps.Register(defaultSceneManager)

	masterHost := viper.GetString("master.host")
	masterPort := viper.GetInt("master.port")
	masterAddr := fmt.Sprintf("%s:%d", masterHost, masterPort)

	// 加密管道
	//c := crypto.NewCrypto()
	//pip := pipeline.New()
	//pip.Inbound().PushBack(c.Inbound)
	//pip.Outbound().PushBack(c.Outbound)

	listen := fmt.Sprintf(":%d", viper.GetInt("game-server.port"))
	logger.Infof("当前游戏服务器版本: %s, 是否强制更新: %t, 当前心跳时间间隔: %d秒", version, forceUpdate, heartbeat)
	logger.Info("game service starup:", listen)
	nano.Listen(listen,
		nano.WithAdvertiseAddr(masterAddr),
		nano.WithDebugMode(),
		//nano.WithPipeline(pip),
		nano.WithLogger(log.WithField("component", "game")),
		nano.WithSerializer(json.NewSerializer()),
		nano.WithComponents(comps),
	)
}
