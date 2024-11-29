package master

import (
	"github.com/ouyangzhongmin/nano/component"
)

var (
	// All services in master server
	Services = &component.Components{}
)

func init() {
	Services.Register(defaultManager)
}
