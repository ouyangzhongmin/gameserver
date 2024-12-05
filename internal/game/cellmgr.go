package game

import (
	"errors"
	"github.com/ouyangzhongmin/gameserver/db/model"
	"github.com/ouyangzhongmin/gameserver/pkg/logger"
	"github.com/ouyangzhongmin/gameserver/pkg/shape"
	"github.com/ouyangzhongmin/gameserver/protocol"
	"github.com/ouyangzhongmin/nano"
	"github.com/ouyangzhongmin/nano/session"
)

type cell struct {
	protocol.Cell
	leftCellId  int //左边的cell, 这里只做竖向一刀切，不做横竖九宫格类似的
	rightCellId int //右边的cell
}

// 是否在cell左边缘
func (c *cell) IsInCellLeftEdge(pos shape.Vector3) bool {
	x, y := int64(pos.X), int64(pos.Y)
	return x >= c.Bounds.X && x <= c.Bounds.X+int64(c.EdgeSize) &&
		y >= c.Bounds.Y && y <= c.Bounds.Y+c.Bounds.Height
}

// 是否在cell右边缘
func (c *cell) IsInCellRightEdge(pos shape.Vector3) bool {
	x, y := int64(pos.X), int64(pos.Y)
	return x >= c.Bounds.X+c.Bounds.Width-int64(c.EdgeSize) && x <= c.Bounds.X+int64(c.Bounds.Width) &&
		y >= c.Bounds.Y && y <= c.Bounds.Y+c.Bounds.Height
}

func (c *cell) IsInCellBounds(pos shape.Vector3) bool {
	return c.Bounds.Contains(int64(pos.X), int64(pos.Y))
}

func (c *cell) UpdateBounds(newbounds shape.Rect) bool {
	return false
}

type cellMgr struct {
	scene   *Scene
	cells   []*protocol.Cell
	curCell *cell
}

func newCellMgr(scene *Scene) *cellMgr {
	return &cellMgr{scene: scene, cells: make([]*protocol.Cell, 0)}
}

// 更新cell的信息
func (mgr *cellMgr) updateCells(curCellId int, cells []*protocol.Cell) error {
	mgr.curCell = nil
	mgr.cells = cells
	for i := 0; i < len(cells); i++ {
		c := mgr.cells[i]
		logger.Println("cell：", i, *c)
		if c.CellID == curCellId {
			mgr.curCell = &cell{
				Cell: *c,
			}
			if i > 0 {
				mgr.curCell.leftCellId = cells[i-1].CellID
			}
			if i < len(cells)-1 {
				mgr.curCell.rightCellId = cells[i+1].CellID
			}
		}
	}
	if mgr.curCell == nil {
		return errors.New("没有找到当前的cell信息")
	}

	return nil
}

func (mgr *cellMgr) getSession(cell *protocol.Cell) (*session.Session, error) {
	var err error
	if cell.Session == nil {
		cell.Session, err = nano.NewRpcSession(cell.GateAddr)
		if err != nil {
			return nil, err
		}
	}
	cell.Session.Set("cellId", cell.CellID)
	cell.Session.Set("sceneId", mgr.scene.sceneId)
	cell.Session.Set("remoteAddr", cell.RemoteAddr)
	cell.Session.Router().Bind("SceneManager", cell.RemoteAddr)
	return cell.Session, nil
}

// 迁移对象到响应的cell服务器
func (mgr *cellMgr) migrateEntities() error {
	logger.Println("migrateEntities...")
	mgr.scene.monsters.Range(func(k, v interface{}) bool {
		m := v.(*Monster)
		isInBounds := mgr.curCell.IsInCellBounds(m.GetPos())
		if !isInBounds {
			//不在当前cell范围内，需要迁移
			err := mgr.migrateMonster(m)
			if err != nil {
				logger.Errorln(err)
			}
		} else {
			//如果在边缘需要创建Ghost对象
			mgr.ghostMonsterIfInEdge(m)
		}
		return true
	})
	return nil
}

func (mgr *cellMgr) ghostMonsterIfInEdge(m *Monster) {
	if len(mgr.cells) <= 1 {
		return
	}
	isInLeftEdge := mgr.curCell.IsInCellLeftEdge(m.GetPos())
	isInRightEdge := mgr.curCell.IsInCellRightEdge(m.GetPos())
	if isInLeftEdge || isInRightEdge {
		//在边缘，需要在邻居cell创建ghost
		err := mgr.createGhostMonster(m, isInLeftEdge, isInRightEdge)
		if err != nil {
			logger.Errorln(err)
		}
	}
}

func (mgr *cellMgr) migrateMonster(m *Monster) error {
	if len(mgr.cells) <= 1 {
		return nil
	}

	for _, c := range mgr.cells {
		if c.Bounds.Contains(int64(m.GetPos().X), int64(m.GetPos().Y)) {
			//找到迁移的目标cell
			ss, err := mgr.getSession(c)
			if err != nil {
				return err
			}
			logger.Debugf("向cell:%d迁移monster:%d \n", c.CellID, m.GetID())
			err = ss.RPC("SceneManager.CreateMigrateMonster", &protocol.MigrateMonsterRequest{
				SceneId:       mgr.scene.sceneId,
				CellId:        c.CellID,
				FromCellId:    mgr.curCell.CellID,
				MonsterObject: m.MonsterObject,
				MonsterData:   &m.MonsterObject.Data,
				Cfg:           m.cfg,
				AiData:        m.aimgr.GetAiData().(*model.Aiconfig),
				MovableRect:   m.movableRect,
				PreparePaths:  m.preparePaths,
				XViewRange:    m.xViewRange,
				YViewRange:    m.yViewRange,
			})
			if err != nil {
				return err
			}
			//传输完成了需要销毁对象
			m.Destroy2()
			break
		}
	}
	return nil
}

// 像邻居cell传输镜像
func (mgr *cellMgr) createGhostMonster(m *Monster, left, right bool) error {
	if len(mgr.cells) <= 1 {
		return nil
	}
	var err error
	var ghostCell *protocol.Cell
	for i, c := range mgr.cells {
		if c.CellID == mgr.curCell.CellID {
			//找到镜像的目标cell
			if i > 0 && left {
				ghostCell = mgr.cells[i-1]
				break
			}
			if i < len(mgr.cells)-1 && right {
				ghostCell = mgr.cells[i+1]
			}
		}
	}
	if ghostCell == nil {
		if left {
			logger.Warningln("没有左邻居cell")
		} else if right {
			logger.Warningln("没有右邻居cell")
		}
		return nil
	}

	ss, err := mgr.getSession(ghostCell)
	if err != nil {
		return err
	}
	logger.Debugf("向cell:%d创建monster Ghost:%d \n", ghostCell.CellID, m.GetID())
	err = ss.RPC("SceneManager.CreateGhostMonster", &protocol.CreateGhostMonsterReq{
		SceneId:       mgr.scene.sceneId,
		CellId:        ghostCell.CellID,
		FromCellId:    mgr.curCell.CellID,
		MonsterObject: m.MonsterObject,
		MonsterData:   &m.MonsterObject.Data,
		Cfg:           m.cfg,
		MovableRect:   m.movableRect,
		XViewRange:    m.xViewRange,
		YViewRange:    m.yViewRange,
	})
	if err != nil {
		return err
	}
	// 这里需要记录哪个cell里有镜像
	m.ghostCellId = ghostCell.CellID
	return nil
}

func (mgr *cellMgr) getCell(cellId int) *protocol.Cell {
	for i, c := range mgr.cells {
		if c.CellID == cellId {
			//找到镜像的目标cell
			return mgr.cells[i]
		}
	}
	return nil
}

func (mgr *cellMgr) removeGhostMonster(m *Monster) error {
	if len(mgr.cells) <= 1 {
		return nil
	}
	var err error
	var ghostCell = mgr.getCell(m.ghostCellId)
	if ghostCell == nil {
		return errors.New("没有找到Ghost的目标cell数据")
	}
	ss, err := mgr.getSession(ghostCell)
	if err != nil {
		return err
	}
	logger.Debugf("向cell:%d删除monster Ghost:%d \n", ghostCell.CellID, m.GetID())
	err = ss.RPC("SceneManager.RemoveGhostMonster", &protocol.RemoveGhostMonsterReq{
		SceneId:   mgr.scene.sceneId,
		CellId:    ghostCell.CellID,
		MonsterId: m.GetID(),
	})
	if err != nil {
		return err
	}

	m.clearGhost()
	return nil
}

func (mgr *cellMgr) updateGhostMonster(m *Monster) error {
	if len(mgr.cells) <= 1 {
		return nil
	}
	var err error
	var ghostCell = mgr.getCell(m.ghostCellId)
	if ghostCell == nil {
		return errors.New("没有找到Ghost的目标cell数据")
	}
	ss, err := mgr.getSession(ghostCell)
	if err != nil {
		return err
	}
	logger.Debugf("向cell:%d同步monster Ghost:%d \n", ghostCell.CellID, m.GetID())
	err = ss.RPC("SceneManager.SyncGhostMonster", &protocol.SyncGhostMonsterReq{
		SceneId:       mgr.scene.sceneId,
		CellId:        ghostCell.CellID,
		MonsterId:     m.GetID(),
		MonsterObject: m.MonsterObject,
	})
	if err != nil {
		return err
	}

	return nil
}

// Ghost受伤转发给Real处理
func (mgr *cellMgr) ghostMonsterBeenHurted(m *Monster, damage int64) error {
	if len(mgr.cells) <= 1 {
		return nil
	}
	var err error
	var ghostCell = mgr.getCell(m.realCellId)
	if ghostCell == nil {
		return errors.New("没有找到Real的目标cell数据")
	}

	ss, err := mgr.getSession(ghostCell)
	if err != nil {
		return err
	}
	err = ss.RPC("SceneManager.GhostMonsterBeenHurted", &protocol.GhostMonsterBeenHurtedReq{
		SceneId:   mgr.scene.sceneId,
		CellId:    ghostCell.CellID,
		MonsterId: m.GetID(),
		Damage:    damage,
	})
	if err != nil {
		return err
	}
	return nil
}

// Ghost被攻击转发给Real处理
func (mgr *cellMgr) ghostMonsterBeenAttaced(m *Monster, target IMovableEntity) error {
	if len(mgr.cells) <= 1 {
		return nil
	}
	var err error
	var ghostCell = mgr.getCell(m.realCellId)
	if ghostCell == nil {
		return errors.New("没有找到Real的目标cell数据")
	}
	ss, err := mgr.getSession(ghostCell)
	if err != nil {
		return err
	}
	err = ss.RPC("SceneManager.GhostMonsterBeenAttaced", &protocol.GhostMonsterBeenAttacedReq{
		SceneId:      mgr.scene.sceneId,
		CellId:       ghostCell.CellID,
		MonsterId:    m.GetID(),
		AttackerId:   target.GetID(),
		AttackerType: target.GetEntityType(),
	})
	if err != nil {
		return err
	}
	return nil
}

func (mgr *cellMgr) migrateHero(h *Hero) error {
	if len(mgr.cells) <= 1 {
		return nil
	}

	for _, c := range mgr.cells {
		if c.Bounds.Contains(int64(h.GetPos().X), int64(h.GetPos().Y)) {
			//找到迁移的目标cell
			//ss, err := mgr.getSession(c)
			//if err != nil {
			//	return err
			//}
			logger.Debugf("向cell:%d迁移hero:%d \n", c.CellID, h.GetID())

			viewListIds := make([]string, 0)
			canSeeMeIds := make([]string, 0)
			h.viewList.Range(func(key, value any) bool {
				viewListIds = append(viewListIds, key.(string))
				return true
			})
			h.canSeeMeViewList.Range(func(key, value any) bool {
				canSeeMeIds = append(canSeeMeIds, key.(string))
				return true
			})

			err := mgr.scene.masterSession.RPC("CellManager.MigrateHero", &protocol.MigrateHeroRequest{
				SceneId:        mgr.scene.sceneId,
				CellId:         c.CellID,
				FromCellId:     mgr.curCell.CellID,
				HeroObject:     h.HeroObject,
				HeroData:       &h.HeroObject.Hero,
				XViewRange:     h.xViewRange,
				YViewRange:     h.yViewRange,
				ViewListIds:    viewListIds,
				CanSeemeIds:    canSeeMeIds,
				TracePath:      h.tracePath,
				TraceIndex:     h.traceIndex,
				TraceTotalTime: h.traceTotalTime,
				TargetX:        h.targetX,
				TargetY:        h.targetY,
				TargetZ:        h.targetZ,
			})
			if err != nil {
				return err
			}
			//传输完成了需要销毁对象
			h.Destroy2()
			break
		}
	}
	return nil
}

// Ghost的消息转发消息给real
func (mgr *cellMgr) GhostSendMsg(h *Hero, route string, msg interface{}) error {
	if len(mgr.cells) <= 1 {
		return nil
	}
	if !h.IsGhost() {
		return nil
	}
	cellId := h.realCellId
	var err error
	var realCell = mgr.getCell(cellId)
	if realCell == nil {
		return errors.New("没有找到Real的目标cell数据")
	}
	ss, err := mgr.getSession(realCell)
	if err != nil {
		return err
	}
	logger.Debugf("Ghost向cell:%d Real转发消息:%d \n", realCell.CellID, h.GetID())
	err = ss.RPC("SceneManager.SendMsgFromGhost", &protocol.SendMsgFromGhostReq{
		SceneId: mgr.scene.sceneId,
		CellId:  realCell.CellID,
		HeroId:  h.GetID(),
		Route:   route,
		Msg:     msg,
	})
	if err != nil {
		return err
	}
	return nil
}

// Broadcast消息转发消息给Ghost
func (mgr *cellMgr) BroadcastToGhost(e IMovableEntity, route string, msg interface{}) error {
	if len(mgr.cells) <= 1 {
		return nil
	}
	if e.IsGhost() {
		return nil
	}
	cellId := 0
	switch val := e.(type) {
	case *Monster:
		cellId = val.ghostCellId
	case *Hero:
		cellId = val.ghostCellId
	}
	var err error
	var ghostCell = mgr.getCell(cellId)
	if ghostCell == nil {
		return errors.New("没有找到Real的目标cell数据")
	}
	ss, err := mgr.getSession(ghostCell)
	if err != nil {
		return err
	}
	logger.Debugf("向cell:%d Ghost转发Broad消息:%d \n", ghostCell.CellID, e.GetID())
	err = ss.RPC("SceneManager.BroadcastToGhost", &protocol.BroadcastToGhostReq{
		SceneId:    mgr.scene.sceneId,
		CellId:     ghostCell.CellID,
		EntityId:   e.GetID(),
		EntityType: e.GetEntityType(),
		Route:      route,
		Msg:        msg,
	})
	if err != nil {
		return err
	}
	return nil
}

func (mgr *cellMgr) Moved(e IMovableEntity, oldx, oldy shape.Coord) (bool, error) {
	if len(mgr.cells) <= 1 {
		return false, nil
	}
	leavedCell := false
	switch val := e.(type) {
	case *Monster:
		if val.IsGhost() {
			return false, nil
		}
		m := val
		isInBounds := mgr.curCell.IsInCellBounds(m.GetPos())
		if !isInBounds {
			//不在当前cell范围内，需要迁移
			leavedCell = true
			err := mgr.migrateMonster(m)
			if err != nil {
				return leavedCell, err
			}
		} else {
			isInLeftEdge := mgr.curCell.IsInCellLeftEdge(m.GetPos())
			isInRightEdge := mgr.curCell.IsInCellRightEdge(m.GetPos())
			if isInLeftEdge || isInRightEdge {
				//在边缘，需要在邻居cell创建ghost
				if !m.HaveGhost() {
					err := mgr.createGhostMonster(m, isInLeftEdge, isInRightEdge)
					if err != nil {
						return leavedCell, err
					}
				}
			} else {
				if m.HaveGhost() {
					//需要清理掉Ghost
					err := mgr.removeGhostMonster(m)
					if err != nil {
						return leavedCell, err
					}
				}
			}
		}
	case *Hero:
		if val.IsGhost() {
			return false, nil
		}
		isInBounds := mgr.curCell.IsInCellBounds(val.GetPos())
		if !isInBounds {
			//不在当前cell范围内，需要迁移
			leavedCell = true
			err := mgr.migrateHero(val)
			if err != nil {
				return leavedCell, err
			}
		} else {
			isInLeftEdge := mgr.curCell.IsInCellLeftEdge(val.GetPos())
			isInRightEdge := mgr.curCell.IsInCellRightEdge(val.GetPos())
			if isInLeftEdge || isInRightEdge {
				//在边缘，需要在邻居cell创建ghost
				if !val.HaveGhost() {
					//err := mgr.createGhostHero(val, isInLeftEdge, isInRightEdge)
					//if err != nil {
					//	return leavedCell, err
					//}
				}
			} else {
				if val.HaveGhost() {
					//需要清理掉Ghost
					//err := mgr.removeGhostHero(val)
					//if err != nil {
					//	return leavedCell, err
					//}
				}
			}
		}
	}
	return leavedCell, nil
}

func (mgr *cellMgr) BeenDestroyed(e IMovableEntity) error {
	if len(mgr.cells) <= 1 {
		return nil
	}
	switch val := e.(type) {
	case *Monster:
		if val.HaveGhost() {
			return mgr.removeGhostMonster(val)
		}
	case *Hero:
		if val.HaveGhost() {

		}
	}
	return nil
}

func (mgr *cellMgr) PropertyChanged(e IMovableEntity) {
	if len(mgr.cells) <= 1 {
		return
	}
	switch val := e.(type) {
	case *Monster:
		if val.HaveGhost() {
			err := mgr.updateGhostMonster(val)
			if err != nil {
				logger.Errorln(err)
			}
		}
	case *Hero:
		if val.HaveGhost() {

		}
	}
}

func (mgr *cellMgr) GhostBeenHurted(e IMovableEntity, damage int64) {
	if len(mgr.cells) <= 1 {
		return
	}
	switch val := e.(type) {
	case *Monster:
		if val.HaveGhost() {
			err := mgr.ghostMonsterBeenHurted(val, damage)
			if err != nil {
				logger.Errorln(err)
			}
		}
	case *Hero:
		if val.HaveGhost() {

		}
	}
}

func (mgr *cellMgr) GhostBeenAttacked(e IMovableEntity, target IMovableEntity) {
	if len(mgr.cells) <= 1 {
		return
	}
	switch val := e.(type) {
	case *Monster:
		if val.HaveGhost() {
			err := mgr.ghostMonsterBeenAttaced(val, target)
			if err != nil {
				logger.Errorln(err)
			}
		}
	case *Hero:
		if val.HaveGhost() {

		}
	}
}
