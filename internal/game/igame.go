package game

import (
	"github.com/ouyangzhongmin/gameserver/db/model"
	"github.com/ouyangzhongmin/gameserver/pkg/shape"
)

type IEntity interface {
	onEnterScene(s *Scene)
	onExitScene(s *Scene)
	GetScene() *Scene
	SetPos(x, y, z shape.Coord)
	GetPos() shape.Vector3
	GetID() int64
	// 不存储在数据库，只作为运行对象的唯一值
	GetUUID() string
	GetEntityType() int
}

type IMovableEntity interface {
	IEntity
	//对象进入我的视野
	onEnterView(target IMovableEntity)
	//对象离开我的视野
	onExitView(target IMovableEntity)
	//我进入对象的视野
	onEnterOtherView(target IMovableEntity)
	//我离开对象的视野
	onExitOtherView(target IMovableEntity)
	update(curMilliSecond int64, elapsedTime int64) error
	GetViewList() map[string]IMovableEntity
	GetCanSeeMeViewList() map[string]IMovableEntity
	GetViewRect() shape.Rect
	SetViewRange(int, int)
	CanSee(target IEntity) bool
	IsInViewList(target IMovableEntity) bool
}

type IAiManager interface {
	update(curMilliSecond int64, elapsedTime int64) error
	onBeenAttacked(target IMovableEntity)
	GetAiData() *model.Aiconfig
	GetOwner() IMovableEntity
}

type IAoiManager interface {
	Enter(entity IMovableEntity)
	Leave(entity IMovableEntity)
	Moved(entity IMovableEntity, oldX, oldY shape.Coord)
}
