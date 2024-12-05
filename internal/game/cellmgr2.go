package game

import (
	"errors"
	"github.com/ouyangzhongmin/gameserver/constants"
	"github.com/ouyangzhongmin/gameserver/pkg/logger"
	"github.com/ouyangzhongmin/gameserver/protocol"
	"github.com/ouyangzhongmin/nano/session"
)

// 从其他cell迁移过来的对象
func (mgr *cellMgr) createMigrateMonsterFromOtherCell(data *protocol.MigrateMonsterRequest) error {
	if mgr.curCell.CellID != data.CellId {
		return errors.New("createMigrateMonsterFromOtherCell err: cellId非法")
	}
	if tmp, ok := mgr.scene.monsters.Load(data.MonsterObject.Id); ok {
		tmpm := tmp.(*Monster)
		logger.Warningln("createMigrateMonsterFromOtherCell: 当前cell已经存在monster:", data.MonsterObject.Id, tmpm.isGhost)
		tmpm.Destroy2()
	}
	data.MonsterObject.Data = *data.MonsterData
	//data.MonsterObject.Name = data.MonsterObject.Name + fmt.Sprintf("-migrate-fromcell-%d", data.FromCellId)
	m := NewMonster2(data.MonsterObject, false)
	m.CellId = data.CellId
	m.realCellId = data.CellId
	m.SetViewRange(data.XViewRange, data.YViewRange)
	m.SetSceneMonsterConfig(data.Cfg)
	m.SetMovableRect(data.MovableRect)
	if data.PreparePaths != nil {
		m.SetPreparePaths(data.PreparePaths)
	}
	m.SetPos(data.MonsterObject.Posx, data.MonsterObject.Posy, data.MonsterObject.Posz)
	if data.AiData != nil {
		m.SetAiData(newMonsterAi(m, data.AiData))
	}
	//m.SetSpells(rm.Spells)
	m.bornPos.Copy(m.GetPos())
	mgr.scene.addMonster(m)
	// 检测新迁移过来的对象是不是在边缘， 如果在边缘则反向cell内创建一个ghost
	mgr.ghostMonsterIfInEdge(m)
	req := &protocol.PropertyChangedRequest{Data: make(map[string]interface{})}
	req.EntityType = m.GetEntityType()
	req.EntityId = m.GetID()
	req.Data["cell_id"] = data.CellId
	req.Data["from_cell_id"] = data.FromCellId
	m.Broadcast(protocol.OnPropertyChanged, req)
	return nil
}

// 从其他cell过来要构建的ghost
func (mgr *cellMgr) createGhostMonsterFromOtherCell(data *protocol.CreateGhostMonsterReq) error {
	logger.Debugln("createGhostMonsterFromOtherCell:", data.MonsterObject.Id, data.CellId, *data.MonsterObject)
	if mgr.curCell.CellID != data.CellId {
		return errors.New("createGhostMonsterFromOtherCell err: cellId非法")
	}
	if tmp, ok := mgr.scene.monsters.Load(data.MonsterObject.Id); ok {
		tmpm := tmp.(*Monster)
		logger.Warningln("createGhostMonsterFromOtherCell: 当前cell已经存在monster:", data.MonsterObject.Id, tmpm.isGhost)
		tmpm.Destroy2()
	}
	data.MonsterObject.Data = *data.MonsterData
	data.MonsterObject.Name = data.MonsterObject.Name + "【Ghost】"
	m := NewMonster2(data.MonsterObject, true)
	m.CellId = data.CellId
	m.realCellId = data.FromCellId // 这个需要记录下来
	m.SetSceneMonsterConfig(data.Cfg)
	m.SetMovableRect(data.MovableRect)
	m.SetPos(data.MonsterObject.Posx, data.MonsterObject.Posy, data.MonsterObject.Posz)
	//m.SetSpells(rm.Spells)
	m.bornPos.Copy(m.GetPos())
	mgr.scene.addMonster(m)
	return nil
}

// 其他cell通知清除掉Ghost
func (mgr *cellMgr) removeGhostMonsterFromOtherCell(data *protocol.RemoveGhostMonsterReq) error {
	if tmp, ok := mgr.scene.monsters.Load(data.MonsterId); ok {
		tmpm := tmp.(*Monster)
		if !tmpm.isGhost {
			logger.Warningln("removeGhostMonsterFromOtherCell: 当前cell不是Ghost状态", data.MonsterId, tmpm.isGhost)
		}
		tmpm.Destroy2()
		return nil
	}
	return errors.New("未找到monster ghost")
}

// 其他cell通知清除掉Ghost
func (mgr *cellMgr) updateGhostMonsterFromOtherCell(data *protocol.SyncGhostMonsterReq) error {
	if tmp, ok := mgr.scene.monsters.Load(data.MonsterId); ok {
		tmpm := tmp.(*Monster)
		if !tmpm.isGhost {
			logger.Warningln("updateGhostMonsterFromOtherCell: 当前cell不是Ghost状态", data.MonsterId, tmpm.isGhost)
			return errors.New("找到的monster不是ghost")
		}
		tmpm.MonsterObject.Copy(data.MonsterObject)
		return nil
	}
	return errors.New("未找到monster ghost")
}

// 转发来自Ghost-》Real的消息广播
func (mgr *cellMgr) ghostSendMsgFromOtherCell(data *protocol.SendMsgFromGhostReq) error {
	if val, ok := mgr.scene.heros.Load(data.HeroId); ok {
		h := val.(*Hero)
		if h.IsGhost() {
			logger.Warningln("sendMsgFromGhost: 当前cell是Ghost状态", data.HeroId, h.isGhost)
			return errors.New("找到的hero是ghost")
		}
		h.SendMsg(data.Route, data.Msg)
		return nil
	}
	return errors.New("没有找到hero")
}

// 转发来自Real-》Ghost的消息广播
func (mgr *cellMgr) broadcastToGhostFromOtherCell(data *protocol.BroadcastToGhostReq) error {
	if data.EntityType == constants.ENTITY_TYPE_HERO {
		if val, ok := mgr.scene.heros.Load(data.EntityId); ok {
			h := val.(*Hero)
			if !h.IsGhost() {
				logger.Warningln("sendMsgFromGhost: 当前cell不是Ghost状态", data.EntityId, h.isGhost)
				return errors.New("找到的hero不是ghost")
			}
			h.Broadcast(data.Route, data.Msg, false)
			return nil
		}
		return errors.New("没有找到hero")
	}
	if val, ok := mgr.scene.monsters.Load(data.EntityId); ok {
		m := val.(*Monster)
		if !m.IsGhost() {
			logger.Warningln("sendMsgFromGhost: 当前cell不是Ghost状态", data.EntityId, m.isGhost)
			return errors.New("找到的monster不是ghost")
		}
		m.Broadcast(data.Route, data.Msg)
		return nil
	}
	return errors.New("没有找到monster")
}

// 从其他cell迁移过来的对象
func (mgr *cellMgr) createMigrateHeroFromOtherCell(ss *session.Session, data *protocol.MigrateHeroRequest) error {
	if mgr.curCell.CellID != data.CellId {
		return errors.New("createMigrateHeroFromOtherCell err: cellId非法")
	}
	if tmp, ok := mgr.scene.heros.Load(data.HeroObject.Id); ok {
		tmpm := tmp.(*Hero)
		logger.Warningln("createMigrateHeroFromOtherCell: 当前cell已经存在hero:", data.HeroObject.Id, tmpm.isGhost)
		tmpm.Destroy2()
	}
	data.HeroObject.Hero = *data.HeroData
	//data.HeroObject.Name = data.HeroObject.Name + fmt.Sprintf("-cell-%d", data.CellId)
	h := NewHero2(ss, data.HeroObject)
	h.CellId = data.CellId
	h.realCellId = data.CellId
	ss.Bind(data.HeroData.Uid)
	h.bindSession(ss)
	h.SetViewRange(data.XViewRange, data.YViewRange)
	h.SetPos(data.HeroObject.Posx, data.HeroObject.Posy, data.HeroObject.Posz)
	h.tracePath = data.TracePath
	h.traceIndex = data.TraceIndex
	h.traceTotalTime = data.TraceTotalTime
	h.targetX = data.TargetX
	h.targetY = data.TargetY
	h.targetZ = data.TargetZ
	mgr.rebuildHeroViewList(h, data.ViewListIds, data.CanSeemeIds)
	h.onEnterScene(mgr.scene)
	mgr.scene.heros.Store(h.GetID(), h)
	mgr.scene.aoiMgr.Enter(h)
	mgr.scene.addToBuildViewList(h)
	req := &protocol.PropertyChangedRequest{Data: make(map[string]interface{})}
	req.EntityType = h.GetEntityType()
	req.EntityId = h.GetID()
	req.Data["cell_id"] = data.CellId
	req.Data["from_cell_id"] = data.FromCellId
	h.Broadcast(protocol.OnPropertyChanged, req, true)
	// 检测新迁移过来的对象是不是在边缘， 如果在边缘则反向cell内创建一个ghost
	//mgr.ghostHeroIfInEdge(h)
	return nil
}

// 刚迁移过来时视野的数据是丢失的
func (mgr *cellMgr) rebuildHeroViewList(entity *Hero, viewListIds, canSeeMeIds []string) {
	entites := mgr.scene.aoiMgr.Search(entity.GetPos().X, entity.GetPos().Y)
	for _, e0 := range entites {
		if e0 == nil {
			continue
		}
		e := e0.(IMovableEntity)
		if e != entity {
			//如果我是英雄， 判定我能不能看见对方
			if entity.CanSee(e) && !entity.IsInViewList(e) {
				find := false
				for _, id := range viewListIds {
					if id == e.GetUUID() {
						find = true
						break
					}
				}
				if find {
					// 之前就在视野范围内，重新加入，但是不需要发送消息给前端了
					entity.movableEntity.onEnterView(e) //进入他的视野
				} else {
					entity.onEnterView(e) //进入他的视野
				}
				e.onEnterOtherView(entity) //记录我进入了谁的视野
			}
			// 以前能看见的重新加入
			find := false
			for _, id := range canSeeMeIds {
				if id == e.GetUUID() {
					find = true
					break
				}
			}
			if find {
				entity.onEnterOtherView(e)
			}
		}
	}
}
