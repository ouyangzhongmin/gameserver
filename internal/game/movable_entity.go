package game

import (
	"github.com/ouyangzhongmin/gameserver/db/model"
	"github.com/ouyangzhongmin/gameserver/internal/game/object"
	"github.com/ouyangzhongmin/gameserver/pkg/coord"
	"github.com/ouyangzhongmin/gameserver/pkg/shape"
	"sync"
)

type movableEntity struct {
	Entity
	//我能看到的对象
	viewList sync.Map
	//能看到我的对象
	canSeeMeViewList sync.Map
	//能看到的x范围
	xViewRange int
	//能看到的y范围
	yViewRange int
	//视野移动的x范围
	xViewMovableRange int
	//视野移动的y范围
	yViewMovableRange int
	//当前的屏幕
	screenRect shape.Rect
	//视野,屏幕可移动范围
	viewRect shape.Rect

	buffers map[int]*Buffer

	isGhost     bool //是否是镜像体, 针对超大无缝地图的英雄镜像实现
	realCellId  int  // 记录本体在哪个cell
	ghostCellId int  // 记录ghost在哪个cell
}

func (m *movableEntity) initEntity(id int64, name string, entityType int, bufSize int) {
	m.Entity.initEntity(id, name, entityType, bufSize)
	m.SetViewRange(30, 30)
	m.buffers = make(map[int]*Buffer)
}

func (m *movableEntity) onEnterScene(scene *Scene) {
	m.Entity.onEnterScene(scene)
}

func (m *movableEntity) onExitScene(scene *Scene) {
	m.Entity.onExitScene(scene)
}

// 对象进入我的视野
func (m *movableEntity) onEnterView(target IMovableEntity) {
	m.viewList.Store(target.GetUUID(), target)
}

// 对象离开我的视野
func (m *movableEntity) onExitView(target IMovableEntity) {
	//delete(m.viewList, target.GetUUID())
	m.viewList.Delete(target.GetUUID())
}

// 我进入对象的视野
func (m *movableEntity) onEnterOtherView(target IMovableEntity) {
	//m.canSeeMeViewList[target.GetUUID()] = target
	m.canSeeMeViewList.Store(target.GetUUID(), target)
}

// 我离开对象的视野
func (m *movableEntity) onExitOtherView(target IMovableEntity) {
	m.canSeeMeViewList.Delete(target.GetUUID())
}

func (m *movableEntity) SetPos(x, y, z coord.Coord) {
	m.Entity.SetPos(x, y, z)
	m.updateViewRect()

}

func (m *movableEntity) SetViewRange(width int, height int) {
	m.xViewRange = width
	m.yViewRange = height
	m.xViewMovableRange = width + 10
	m.yViewMovableRange = height + 10
}

func (m *movableEntity) GetViewRange() (int, int) {
	return m.xViewRange, m.yViewRange
}

func (m *movableEntity) GetViewList() map[string]IMovableEntity {
	list := make(map[string]IMovableEntity)
	m.viewList.Range(func(key, value interface{}) bool {
		list[key.(string)] = value.(IMovableEntity)
		return true
	})
	return list
}

func (m *movableEntity) GetCanSeeMeViewList() map[string]IMovableEntity {
	list := make(map[string]IMovableEntity)
	m.canSeeMeViewList.Range(func(key, value interface{}) bool {
		list[key.(string)] = value.(IMovableEntity)
		return true
	})
	return list
}

func (m *movableEntity) IsInViewList(target IMovableEntity) bool {
	if _, ok := m.viewList.Load(target.GetUUID()); ok {
		return true
	}
	return false
}

func (m *movableEntity) GetViewRect() shape.Rect {
	return m.viewRect
}

func (m *movableEntity) Destroy() {
	m.Entity.Destroy()
	//清理掉引用
	m.viewList.Range(func(key, value interface{}) bool {
		m.viewList.Delete(key)
		return true
	})
	m.canSeeMeViewList.Range(func(key, value interface{}) bool {
		m.canSeeMeViewList.Delete(key)
		return true
	})
	m.buffers = nil
}

func (m *movableEntity) update(curMilliSecond int64, elapsedTime int64) error {
	m.updateBuffers(curMilliSecond, elapsedTime)
	return nil
}

func (m *movableEntity) updateViewRect() {
	//刷新当前角色屏幕范围
	x := m.GetPos().X
	y := m.GetPos().Y
	//视野的范围
	m.screenRect.X = int64(int(x) - m.xViewRange)
	m.screenRect.Y = int64(int(y) - m.yViewRange)
	m.screenRect.Width = int64(2 * m.xViewRange)
	m.screenRect.Height = int64(2 * m.yViewRange)
	if m.screenRect.X < 0 {
		m.screenRect.X = 0
	}
	if m.screenRect.Y < 0 {
		m.screenRect.Y = 0
	}
	//视野范围未变化
	if !m.viewRect.ContainsRect(m.screenRect) {
		//重新刷新视野范围
		m.viewRect.X = int64(int(x) - m.xViewMovableRange)
		m.viewRect.Y = int64(int(y) - m.yViewMovableRange)
		m.viewRect.Width = int64(2 * m.xViewMovableRange)
		m.viewRect.Height = int64(2 * m.yViewMovableRange)
		if m.viewRect.X < 0 {
			m.viewRect.X = 0
		}
		if m.viewRect.Y < 0 {
			m.viewRect.Y = 0
		}
	}
}

func (m *movableEntity) CanSee(target IEntity) bool {
	return m.viewRect.Contains(int64(target.GetPos().X), int64(target.GetPos().Y))
}

func (m *movableEntity) addBuffer(owner IMovableEntity, state *model.BufferState) {
	m.PushTask(func() {
		if m.isGhost {

		} else {
			if o, ok := m.buffers[state.Id]; ok {
				//叠加
				o.Add(state)
			} else {
				o := NewBuffer(owner, state)
				m.buffers[state.Id] = o
			}
		}

	})
}

func (m *movableEntity) removeBuffer(bufId int) {
	m.PushTask(func() {
		if m.isGhost {

		} else {
			delete(m.buffers, bufId)
		}
	})
}

func (m *movableEntity) updateBuffers(curMilliSecond int64, elapsedTime int64) {
	for _, buf := range m.buffers {
		buf.update(curMilliSecond, elapsedTime)
	}
}

func (m *movableEntity) GetBuffers() []*object.BufferObject {
	result := make([]*object.BufferObject, 0)
	for _, buf := range m.buffers {
		result = append(result, buf.BufferObject)
	}
	return result
}

// 是否拥有镜像
func (m *movableEntity) HaveGhost() bool {
	return m.ghostCellId > 0
}

func (m *movableEntity) IsGhost() bool {
	return m.isGhost
}

func (m *movableEntity) clearGhost() {
	m.ghostCellId = 0
}
