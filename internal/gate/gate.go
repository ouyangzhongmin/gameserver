package gate

import (
	"fmt"
	"github.com/lonng/nano"
	"github.com/lonng/nano/serialize/json"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
	"time"
)

var (
	logger = log.WithField("component", "gate")
)

// Startup 初始化gate服务器
func Startup() {

	version := viper.GetString("update.version")
	heartbeat := viper.GetInt("core.heartbeat")
	if heartbeat < 5 {
		heartbeat = 5
	}
	forceUpdate := viper.GetBool("update.force")
	// register game handler
	comps := Services

	// 加密管道
	//c := crypto.NewCrypto()
	//pip := pipeline.New()
	//pip.Inbound().PushBack(c.Inbound)
	//pip.Outbound().PushBack(c.Outbound)
	masterHost := viper.GetString("master.host")
	masterPort := viper.GetInt("master.port")
	masterAddr := fmt.Sprintf("%s:%d", masterHost, masterPort)
	gateAddress := viper.GetString("gate.gate-address")

	listen := fmt.Sprintf(":%d", viper.GetInt("gate.port"))
	logger.Infof("当前gate server服务器版本: %s, 是否强制更新: %t, 当前心跳时间间隔: %d秒", version, forceUpdate, heartbeat)
	logger.Info("gate service starup: ", listen, ",,,gateAddress: ", gateAddress)
	nano.Listen(listen,
		nano.WithAdvertiseAddr(masterAddr),
		nano.WithClientAddr(gateAddress),
		nano.WithIsWebsocket(true),
		nano.WithCheckOriginFunc(func(_ *http.Request) bool { return true }),
		//nano.WithPipeline(pip),
		nano.WithHeartbeatInterval(time.Duration(heartbeat)*time.Second),
		nano.WithLogger(log.WithField("component", "gate")),
		nano.WithSerializer(json.NewSerializer()),
		nano.WithComponents(comps),
	)
}
