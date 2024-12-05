package master

import (
	"github.com/ouyangzhongmin/nano/component"
)

var (
	// All services in master server
	Services    = &component.Components{}
	userManager = NewManager()
	cellManager = NewCellManager()
)

func init() {
	Services.Register(userManager)
	Services.Register(cellManager)
}
