package gate

import (
	"github.com/lonng/nano/component"
	"github.com/lonng/nano/examples/cluster/protocol"
	"github.com/lonng/nano/session"
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

func (ts *GateService) Stats(s *session.Session, msg *protocol.MasterStats) error {
	// It's OK to use map without lock because of this service running in main thread

	return nil
}
