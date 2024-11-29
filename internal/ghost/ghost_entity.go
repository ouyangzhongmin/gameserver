package ghost

import (
	"github.com/ouyangzhongmin/gameserver/pkg/shape"
	"github.com/ouyangzhongmin/gameserver/protocol"
	"github.com/ouyangzhongmin/nano/session"
)

type ghostEntity struct {
	scene   *ghostScene
	session *session.Session
	protocol.GhostEntity
}

func (e *ghostEntity) GetPos() shape.Vector3 {
	return e.Pos
}

func (e *ghostEntity) GetID() int64 {
	return e.Id
}

func (e *ghostEntity) GetUUID() string {
	return e.Uuid
}

func (e *ghostEntity) GetEntityType() int {
	return e.EntityType
}

// 对象进入我的视野
func (m *ghostEntity) onEnterView(target *ghostEntity) {
}

// 对象离开我的视野
func (m *ghostEntity) onExitView(target *ghostEntity) {
	//delete(m.viewList, target.GetUUID())
}

// 我进入对象的视野
func (m *ghostEntity) onEnterOtherView(target *ghostEntity) {

}

// 我离开对象的视野
func (m *ghostEntity) onExitOtherView(target *ghostEntity) {

}

func (m *ghostEntity) CanSee(target *ghostEntity) bool {
	return m.ViewRect.Contains(int64(target.GetPos().X), int64(target.GetPos().Y))
}
