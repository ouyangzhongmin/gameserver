package ghost

import (
	"fmt"
	"github.com/ouyangzhongmin/gameserver/pkg/logger"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/ouyangzhongmin/nano"
	"github.com/ouyangzhongmin/nano/component"
	"github.com/ouyangzhongmin/nano/serialize/json"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Startup 初始化游戏服务器
func Startup(scenes string) {
	rand.Seed(time.Now().Unix())
	heartbeat := viper.GetInt("core.heartbeat")
	if heartbeat < 5 {
		heartbeat = 5
	}

	// register game handler
	sceneIds := parseScenes(scenes)
	defaultGhostManager.setSceneIds(sceneIds)
	comps := &component.Components{}
	comps.Register(defaultGhostManager)

	masterHost := viper.GetString("master.host")
	masterPort := viper.GetInt("master.port")
	masterAddr := fmt.Sprintf("%s:%d", masterHost, masterPort)

	// 加密管道
	//c := crypto.NewCrypto()
	//pip := pipeline.New()
	//pip.Inbound().PushBack(c.Inbound)
	//pip.Outbound().PushBack(c.Outbound)

	listen := fmt.Sprintf(":%d", viper.GetInt("ghost-server.port"))
	logger.Log.WithField("component", "ghost")
	logger.Info("ghost service starup:", listen)
	nano.Listen(listen,
		nano.WithLabel("ghost:"+scenes), //通过这个实现的消息对应场景服务器的处理
		nano.WithAdvertiseAddr(masterAddr),
		nano.WithDebugMode(),
		//nano.WithPipeline(pip),
		nano.WithLogger(log.WithField("component", "ghost")),
		nano.WithSerializer(json.NewSerializer()),
		nano.WithComponents(comps),
	)
}

func parseScenes(scenes string) []int {
	sceneIds := make([]int, 0)
	for _, s := range strings.Split(scenes, ",") {
		sid, err := strconv.Atoi(s)
		if err != nil {
			panic(err)
		}
		sceneIds = append(sceneIds, sid)
	}
	return sceneIds
}
