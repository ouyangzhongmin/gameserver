package gate

import (
	"github.com/ouyangzhongmin/gameserver/protocol"
	"github.com/ouyangzhongmin/nano/component"
	"github.com/ouyangzhongmin/nano/session"
)

var (
	// All services in master server
	Services = &component.Components{}

	bindService = newGateService()
)

func init() {
	Services.Register(bindService)
}

type GateService struct {
	component.Base
	nextGateUid int64
}

func newGateService() *GateService {
	return &GateService{}
}

// 在进入场景的时候需要记录session和对应的sceneId， 在调用SceneManager时需要查找对应的node服务器
func (ts *GateService) RecordScene(s *session.Session, msg *protocol.UserSceneId) error {
	s.Bind(msg.Uid)
	s.Set("sceneId", msg.SceneId)
	s.Set("cellId", msg.CellId)
	if msg.RemoteAddr != "" {
		s.Set("remoteAddr", msg.RemoteAddr)
		s.Router().Bind("SceneManager", msg.RemoteAddr)
	}

	return nil
}
