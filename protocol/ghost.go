package protocol

import (
	"github.com/ouyangzhongmin/gameserver/pkg/coord"
	"github.com/ouyangzhongmin/gameserver/pkg/shape"
)

type GhostEntity struct {
	SceneID    int
	Uuid       string
	Id         int64
	Name       string
	EntityType int
	Pos        coord.Vector3
	//视野,屏幕可移动范围
	ViewRect shape.Rect
}

// 简单的镜像体
type GhostEntitySimple struct {
	SceneID    int
	Uuid       string
	EntityType int
}
