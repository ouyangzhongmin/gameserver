package game

import (
	"errors"
	"fmt"
	"math"
	"sync/atomic"
	"time"

	"github.com/ouyangzhongmin/gameserver/constants"
	"github.com/ouyangzhongmin/gameserver/db/model"
	"github.com/ouyangzhongmin/gameserver/internal/game/object"
	"github.com/ouyangzhongmin/gameserver/pkg/coord"
	"github.com/ouyangzhongmin/gameserver/pkg/logger"
	"github.com/ouyangzhongmin/gameserver/pkg/path"
	"github.com/ouyangzhongmin/gameserver/pkg/shape"
	"github.com/ouyangzhongmin/gameserver/protocol"
)

type rebornMonster struct {
	Uid             string
	Data            *model.Monster
	PreparePaths    *path.SerialPaths
	MovableRect     shape.Rect
	PathFinder      *PathFinder
	Aidata          interface{}
	Cfg             *model.SceneMonsterConfig
	Spells          []*object.SpellObject
	RebornTimestamp int64
}

// PhaseModifier 阶段修饰器
type PhaseModifier struct {
	SkillCooldownMultiplier  float64 // 技能冷却时间乘数
	AttackCooldownMultiplier float64 // 技能冷却时间乘数
	DamageMultiplier         float64 // 伤害乘数
	DefenseMultiplier        float64 // 防御乘数
	Transform                string  // 变身形态
}

type Monster struct {
	*object.MonsterObject
	movableEntity
	tracePath       [][]int32 //当前移动路径
	traceIndex      int       //当前已移动到第几步
	traceTotalTime  int64     //当前移动的总时间
	movableRect     shape.Rect
	aimgr           IAiManager
	pathFinder      *PathFinder
	preparePaths    *path.SerialPaths //预制的移动路径
	cfg             *model.SceneMonsterConfig
	bornPos         coord.Vector3
	spells          []*object.SpellObject
	propertyChanged atomic.Bool
	stateEnterTime  time.Time

	// Boss AI相关字段
	currentPhaseModifier *PhaseModifier // 当前阶段修饰器
}

func NewMonster(data *model.Monster, offset int) *Monster {
	m := &Monster{
		MonsterObject: object.NewMonsterObject(data, offset),
	}
	m.initEntity(m.MonsterObject.Id, data.Name, constants.ENTITY_TYPE_MONSTER, 128)
	m.GameObject.Uuid = m.GetUUID()
	return m
}

func NewMonster2(obj *object.MonsterObject, isGhost bool) *Monster {
	m := &Monster{
		MonsterObject: obj,
	}
	m.initEntity(m.MonsterObject.Id, obj.Name, constants.ENTITY_TYPE_MONSTER, 128)
	m.isGhost = isGhost
	return m
}

func (m *Monster) SetMovableRect(rect shape.Rect) {
	m.movableRect = rect
}

func (m *Monster) GetMovableRect() shape.Rect {
	return m.movableRect
}

func (m *Monster) SetViewRange(width int, height int) {
	m.movableEntity.SetViewRange(width, height)
	if m.scene != nil {
		//更新block数据
		m.scene.addToBuildViewList(m)
	}
}

func (m *Monster) SetAiData(aimgr IAiManager) {
	m.aimgr = aimgr
}

// SetPhaseModifier 设置阶段修饰器
func (m *Monster) SetPhaseModifier(modifier *PhaseModifier) {
	m.currentPhaseModifier = modifier
}

// GetPhaseModifier 获取阶段修饰器
func (m *Monster) GetPhaseModifier() *PhaseModifier {
	return m.currentPhaseModifier
}

// 返回普通攻击间隔
func (m *Monster) GetAttackDuration() int {
	if m.currentPhaseModifier != nil && m.currentPhaseModifier.AttackCooldownMultiplier > 0 {
		return int(float64(m.Data.AttackDuration) * m.currentPhaseModifier.AttackCooldownMultiplier)
	}
	return m.Data.AttackDuration
}

// ApplyTransform 应用变身
func (m *Monster) ApplyTransform(transform string) {
	// 这里可以实现具体的变身逻辑
	// 比如改变外观、属性等
	logger.Printf("Monster %d applying transform: %s", m.GetID(), transform)

	// 示例：发送变身消息给客户端
	m.Broadcast("OnMonsterTransform", &protocol.MonsterTransformResponse{
		ID:        m.GetID(),
		Transform: transform,
	})
}

func (m *Monster) SetSpells(spells []*object.SpellObject) {
	m.spells = make([]*object.SpellObject, 0)
	if spells != nil {
		for _, spell := range spells {
			// 需要拷贝对象，不能直接用指针，否则cd会共用一个指针
			var spell2 = *spell
			m.spells = append(m.spells, &spell2)
		}
	}
}

func (m *Monster) SetPreparePaths(p *path.SerialPaths) {
	m.preparePaths = p
}

func (m *Monster) SetSceneMonsterConfig(cfg *model.SceneMonsterConfig) {
	m.cfg = cfg
}

func (m *Monster) SetPos(x, y, z coord.Coord) {
	oldx, oldy := m.GetPos().X, m.GetPos().Y
	m.Posx = x
	m.Posy = y
	m.Posz = z
	m.movableEntity.SetPos(x, y, z)
	if m.scene != nil && (oldx != x || oldy != y) {
		//fmt.Println("monster SetPos :", m._uuid, x, y, oldx, oldy)
		//更新block数据 go的继承关系是组合关系，这个逻辑如果写在movableEntity会导致存储的对象是*moveableEntity，并不是*Monster
		m.scene.entityMoved(m, x, y, oldx, oldy)
		m.propertyChanged.Store(true)
	}
}

func (m *Monster) GetData() *object.MonsterObject {
	return m.MonsterObject
}

func (m *Monster) onEnterScene(scene *Scene) {
	m.movableEntity.onEnterScene(scene)
	//更新block数据
	m.scene.addToBuildViewList(m)
	m.pathFinder = NewPathFinder(m.scene.blockInfo.GetBlockTable())
}

func (m *Monster) onExitScene(scene *Scene) {
	m.movableEntity.onExitScene(scene)
}

func (m *Monster) onEnterView(target IMovableEntity) {
	m.movableEntity.onEnterView(target)
	switch val := target.(type) {
	case *Hero:
		logger.Debugf("hero::%d-%s monster:%d-%s视野\n", val.GetID(), val._name, m.GetID(), m._name)
	case *Monster:
		//有对象进入自己的视野了，推送给前端创建对象
		logger.Debugf("monster::%d-%s 进入monster:%d-%s视野\n", val.GetID(), val._name, m.GetID(), m._name)
	case *SpellEntity:
		//有对象进入自己的视野了，推送给前端创建对象
	}
}

func (m *Monster) onExitView(target IMovableEntity) {
	m.movableEntity.onExitView(target)
	ttype := -1
	switch target.(type) {
	case *Hero:
		ttype = constants.ENTITY_TYPE_HERO
	case *Monster:
		ttype = constants.ENTITY_TYPE_MONSTER
	case *SpellEntity:
		ttype = constants.ENTITY_TYPE_SPELL
	}
	logger.Debugf("对象:%d-%d离开monster:%d_%s视野:", target.GetID(), ttype, m.GetID(), m._name)
}

func (m *Monster) ToString() string {
	baseInfo := fmt.Sprintf("id:%d,uuid:%s, posX:%d, posY:%d, posZ:%d", m.GetID(), m.GetUUID(), m.GetPos().X, m.GetPos().Y, m.GetPos().Z)
	return fmt.Sprintf("baseInfo:%s,,,data::%v", baseInfo, m.Data)
}

func (m *Monster) Destroy() {
	if m.scene != nil {
		m.scene.cellMgr.BeenDestroyed(m)
		m.scene.removeMonster(m)
	}
	m.viewList.Range(func(key, value interface{}) bool {
		value.(IMovableEntity).onExitOtherView(m)
		return true
	})
	m.canSeeMeViewList.Range(func(key, value interface{}) bool {
		value.(IMovableEntity).onExitView(m)
		return true
	})
	m.movableEntity.Destroy()
	m.pathFinder = nil
	if m.aimgr != nil {
		m.aimgr.clear()
	}
	m.aimgr = nil
}

func (m *Monster) Destroy2() {
	if m.scene != nil {
		m.scene.removeMonster(m)
	}
	m.movableEntity.Destroy()
	m.pathFinder = nil
	if m.aimgr != nil {
		m.aimgr.clear()
	}
	m.aimgr = nil
}

func (m *Monster) IsNpc() bool {
	return m.MonsterType == constants.MONSTER_TYPE_NPC
}

// 广播给所有能看见自己的对象
func (m *Monster) Broadcast(route string, msg interface{}) {
	m.PushTask(func() {
		m.canSeeMeViewList.Range(func(key, value interface{}) bool {
			switch val := value.(type) {
			case *Hero:
				val.SendMsg(route, msg)
			}
			return true
		})
		if m.scene != nil && !m.IsGhost() {
			if m.HaveGhost() {
				err := m.scene.cellMgr.BroadcastToGhost(m, route, msg)
				if err != nil {
					logger.Errorln(err)
				}
			}
		}
	})
}

// 动作状态
func (m *Monster) SetState(state constants.ActionState) {
	if m.State != state {
		m.State = state
		m.propertyChanged.Store(true)
		m.stateEnterTime = time.Now()
	}
}

func (m *Monster) GetState() constants.ActionState {
	return m.State
}

func (m *Monster) AttackAction() {
	m.SetState(constants.ACTION_STATE_ATTACK)
}

func (m *Monster) Idle() {
	m.SetState(constants.ACTION_STATE_IDLE)
}

func (m *Monster) Walk() {
	m.SetState(constants.ACTION_STATE_WALK)
}

func (m *Monster) Run() {
	m.SetState(constants.ACTION_STATE_RUN)
}

func (m *Monster) Chase() {
	m.SetState(constants.ACTION_STATE_CHASE) //追击
}

func (m *Monster) Escape() {
	m.SetState(constants.ACTION_STATE_ESCAPE) //逃跑
}

func (m *Monster) Die() {
	m.SetState(constants.ACTION_STATE_DIE)
	if !m.IsGhost() {
		logger.Debugf("hero:%d-%s die", m.GetID(), m._name)
		m.Broadcast(protocol.OnEntityDie, &protocol.EntityDieResponse{
			ID:         m.GetID(),
			EntityType: constants.ENTITY_TYPE_MONSTER,
		})
		// 加入复活队列中
		m.scene.addRebornMonster(&rebornMonster{
			Uid:             m.GetUUID(),
			Data:            &m.MonsterObject.Data,
			PreparePaths:    m.preparePaths,
			MovableRect:     m.movableRect,
			PathFinder:      m.pathFinder,
			Aidata:          m.aimgr.GetAiData(),
			Cfg:             m.cfg,
			Spells:          m.spells,
			RebornTimestamp: time.Now().UnixMilli() + int64(m.cfg.Reborn)*1000,
		})

		m.PushTask(func() {
			m.Destroy()
		})
	}
}

// update都会在task携程内执行
func (m *Monster) update(curMilliSecond int64, elapsedTime int64) error {
	err := m.movableEntity.update(curMilliSecond, elapsedTime)
	if m.haveStepsToGo() {
		m.updateMonsterPosition(curMilliSecond, elapsedTime)
	}
	if m.aimgr != nil {
		err = m.aimgr.update(curMilliSecond, elapsedTime)
		if err != nil {
			logger.Errorln("aimgr update err:", err)
		}
	}
	if m.spells != nil {
		for _, spell := range m.spells {
			if spell.CurCdTime > 0 {
				spell.Update(elapsedTime)
			}
		}
	}
	if m.propertyChanged.Load() && m.HaveGhost() {
		if m.scene != nil {
			m.scene.cellMgr.PropertyChanged(m)
		}
		m.propertyChanged.Store(false)
	}
	if m.GetState() == constants.ACTION_STATE_ATTACK {
		// 如果是攻击状态，这个状态需要段时间内结束
		if time.Now().Sub(m.stateEnterTime) > 300*time.Millisecond {
			// 这里直接固定一个攻击动作时间就是300ms
			m.Idle()
		}
	}
	return err
}

func (m *Monster) haveStepsToGo() bool {
	return m.tracePath != nil && len(m.tracePath) > 0
}

func (m *Monster) updateMonsterPosition(curMilliSecond int64, elapsedTime int64) {
	m.traceTotalTime += elapsedTime
	stepTime := m.getStepTime()
	//当前移动总时间/速度得到移动到了第几步
	m.traceIndex = int(m.traceTotalTime / int64(stepTime))
	if m.traceIndex >= len(m.tracePath) {
		m.traceIndex = len(m.tracePath) - 1
		//移动到了目标点
		m.Stop()
		return
	}
	step := m.tracePath[m.traceIndex]
	//寻路返回的0是y坐标，1是X坐标，注意了
	m.SetPos(coord.Coord(step[1]), coord.Coord(step[0]), 0)
}

// 移动到目标位置
func (m *Monster) MoveTo(x, y, z coord.Coord) error {
	if m.IsGhost() {
		logger.Warningln("Ghost can't MoveTo")
		return nil
	}
	if m.GetPos().X == x && m.GetPos().Y == y {
		//在目标点了
		m.Idle()
		return nil
	}
	if m.scene == nil {
		return errors.New("scene is nil")
	}
	paths, err := m.pathFinder.FindPath(int(m.GetPos().X), int(m.GetPos().Y), int(x), int(y))
	if err != nil {
		return err
	}
	return m.MoveByPaths(paths)
}

func (m *Monster) MoveByPaths(paths [][]int32) error {
	if m.IsGhost() {
		logger.Warningln("Ghost can't MoveByPaths")
		return nil
	}
	if paths == nil || len(paths) == 0 {
		return errors.New("monster没有路径可走")
	}
	m.tracePath = paths
	m.traceIndex = 0
	m.traceTotalTime = 0
	m.Walk()

	stepTime := m.getStepTime()
	m.Broadcast(protocol.OnMonsterMoveTrace, &protocol.MonsterMoveTraceResponse{
		ID:         m.GetID(),
		TracePaths: paths,
		StepTime:   stepTime,
		PosX:       m.GetPos().X,
		PosY:       m.GetPos().Y,
		PosZ:       m.GetPos().Z,
	})
	return nil
}

// 把一些信息发给英雄
func (m *Monster) sendDataToHero(h *Hero) {
	m.PushTask(func() {
		if m.tracePath != nil && len(m.tracePath) > 0 && m.traceIndex < len(m.tracePath) {
			//正在行走中
			newPaths := m.tracePath[m.traceIndex:]
			//logger.Debugf("monster::%d-%s 发送移动路径:%v", m.GetID(), m._name, newPaths)
			h.SendMsg(protocol.OnMonsterMoveTrace, &protocol.MonsterMoveTraceResponse{
				ID:         m.GetID(),
				TracePaths: newPaths,
				StepTime:   m.getStepTime(),
				PosX:       m.GetPos().X,
				PosY:       m.GetPos().Y,
				PosZ:       m.GetPos().Z,
			})
		}
	})
}

func (m *Monster) Stop() {
	m.PushTask(func() {
		m.tracePath = nil
		m.traceIndex = 0
		m.traceTotalTime = 0
		m.Idle()
		//广播消息
		m.Broadcast(protocol.OnMonsterMoveStopped, &protocol.MonsterMoveStopResponse{
			ID:   m.GetID(),
			PosX: m.GetPos().X,
			PosY: m.GetPos().Y,
			PosZ: m.GetPos().Z,
		})
	})
}

func (m *Monster) CanAttackTarget(target IEntity) bool {
	switch val := target.(type) {
	case *Hero:
		return val.IsAlive() && !val.IsOffline() && !val.IsDestroyed()
	case *Monster:
		return false
	}
	return false
}

func (m *Monster) IsInAttackRange(x, y coord.Coord) bool {
	if m.Data.AttackRange > 50 {
		return true
	}
	return int(math.Abs(float64(x-m.GetPos().X))) <= m.Data.AttackRange && int(math.Abs(float64(y-m.GetPos().Y))) <= m.Data.AttackRange
}

func (m *Monster) onBeenAttacked(attacker IMovableEntity) {
	if m.aimgr != nil {
		m.aimgr.onBeenAttacked(attacker)
	}
	if m.IsGhost() {
		//  转发给real处理
		if m.scene != nil {
			m.scene.cellMgr.GhostBeenAttacked(m, attacker)
		}
		return
	}
}

func (m *Monster) onBeenHurt(damage int64) {
	m.PushTask(func() {
		if m.IsGhost() {
			//  转发给real处理
			if m.scene != nil {
				m.scene.cellMgr.GhostBeenHurted(m, damage)
			}
			return
		}
		if !m.IsAlive() {
			logger.Warningln("hero is dead")
			return
		}
		m.Life -= damage
		if m.Life < 0 {
			m.Life = 0
		}
		if m.Life > m.MaxLife {
			m.Life = m.MaxLife
		}
		m.propertyChanged.Store(true)
		m.Broadcast(protocol.OnLifeChanged, &protocol.LifeChangedResponse{
			ID:         m.GetID(),
			EntityType: constants.ENTITY_TYPE_MONSTER,
			Damage:     damage,
			Life:       m.Life,
			MaxLife:    m.MaxLife,
		})
		if m.Life == 0 {
			m.Die()
			//死亡了
		}
	})
}

func (m *Monster) manaCost(mana int64) {
	m.PushTask(func() {
		if m.IsGhost() {
			logger.Warningln("Ghost can't manaCost")
			return
		}
		if !m.IsAlive() {
			logger.Warningln("monster is died")
			return
		}
		m.Mana -= mana
		if m.Mana < 0 {
			m.Mana = 0
		}
		if m.Mana > m.MaxMana {
			m.Mana = m.MaxMana
		}
		m.propertyChanged.Store(true)
		m.Broadcast(protocol.OnManaChanged, &protocol.ManaChangedResponse{
			ID:         m.GetID(),
			EntityType: constants.ENTITY_TYPE_MONSTER,
			Cost:       mana,
			Mana:       m.Mana,
			MaxMana:    m.MaxMana,
		})
	})
}

func (m *Monster) doAttackTarget(target IMovableEntity) {
	m.PushTask(func() {
		m.AttackAction()
		attack := m.GetAttack()

		// 应用伤害乘数
		if m.currentPhaseModifier != nil && m.currentPhaseModifier.DamageMultiplier > 0 {
			attack = int64(float64(attack) * m.currentPhaseModifier.DamageMultiplier)
		}

		var defense int64 = 0
		ttype := constants.ENTITY_TYPE_HERO
		switch val := target.(type) {
		case *Hero:
			ttype = constants.ENTITY_TYPE_HERO
			defense = val.GetDefense()

			// 应用防御乘数
			if m.currentPhaseModifier != nil && m.currentPhaseModifier.DefenseMultiplier > 0 {
				defense = int64(float64(defense) * m.currentPhaseModifier.DefenseMultiplier)
			}
		case *Monster:
			ttype = constants.ENTITY_TYPE_MONSTER
			defense = val.GetDefense()

			// 应用防御乘数
			if m.currentPhaseModifier != nil && m.currentPhaseModifier.DefenseMultiplier > 0 {
				defense = int64(float64(defense) * m.currentPhaseModifier.DefenseMultiplier)
			}
		}
		damage := attack - defense
		if damage < 1 { //至少有1点伤害
			damage = 1
		}
		switch val := target.(type) {
		case *Hero:
			val.onBeenHurt(damage)
			val.onBeenAttacked(m)
		case *Monster:
			val.onBeenHurt(damage)
			val.onBeenAttacked(m)
		}

		m.Broadcast(protocol.OnMonsterCommonAttack, &protocol.MonsterAttackResponse{
			ID:         m.GetID(),
			Action:     "common",
			Damage:     damage,
			TargetId:   target.GetID(),
			EntityType: ttype,
			PosX:       m.GetPos().X,
			PosY:       m.GetPos().Y,
			PosZ:       m.GetPos().Z,
		})

	})
}

// 返回对目标点的可攻击位置
func (m *Monster) GetCanAttackPos(target IEntity, offset int) (v coord.Vector3, err error) {
	if offset >= 20 {
		return v, errors.New("附近没有可以站立的位置")
	}
	var tx coord.Coord = 0
	var ty coord.Coord = 0
	if m.GetPos().X < target.GetPos().X {
		// +-1是为了一定能到attackRange范围去
		tx = target.GetPos().X - coord.Coord(m.Data.AttackRange) + coord.Coord(offset)
	} else {
		tx = target.GetPos().X + coord.Coord(m.Data.AttackRange) - coord.Coord(offset)
	}
	if m.GetPos().Y < target.GetPos().Y {
		ty = target.GetPos().Y - coord.Coord(m.Data.AttackRange) + coord.Coord(offset)
	} else {
		ty = target.GetPos().Y + coord.Coord(m.Data.AttackRange) - coord.Coord(offset)
	}
	if tx < 0 {
		tx = 0
	}
	if ty < 0 {
		ty = 0
	}
	if !m.scene.IsWalkable(tx, ty) {
		return m.GetCanAttackPos(target, offset+1)
	}
	return coord.Vector3{
		X: tx,
		Y: ty,
		Z: 0,
	}, nil
}

func (m *Monster) getStepTime() int {
	stepTime := m.Data.IdleStepTime
	if m.State == constants.ACTION_STATE_RUN {
		stepTime = m.Data.RunStepTime
	} else if m.State == constants.ACTION_STATE_ESCAPE {
		stepTime = m.Data.EscapeStepTime
	} else if m.State == constants.ACTION_STATE_CHASE {
		stepTime = m.Data.ChaseStepTime
	}
	return stepTime
}

func (m *Monster) GetSpell(spellId int) *object.SpellObject {
	if m.spells == nil || len(m.spells) == 0 {
		return nil
	}
	for _, spell := range m.spells {
		if spell.SpellId == spellId {
			return spell
		}
	}
	return nil
}

func (m *Monster) GetCanUseSpell(spellType int) *object.SpellObject {
	if m.spells == nil || len(m.spells) == 0 {
		return nil
	}
	for _, spell := range m.spells {
		if spell.SpellType == spellType && spell.CurCdTime <= 0 && m.Mana >= spell.Data.Mana {
			return spell
		}
	}
	return nil
}

func (m *Monster) IsInSpellAttackRange(spell *object.SpellObject, x, y coord.Coord) bool {
	return int(math.Abs(float64(x-m.GetPos().X))) <= spell.Data.AttackRange && int(math.Abs(float64(y-m.GetPos().Y))) <= spell.Data.AttackRange
}

func (m *Monster) SpellAttack(spell *object.SpellObject, target IMovableEntity) error {
	//logger.Debugf("monster:%d attack SpellAttack:%d-%d \n", m.GetID(),  a.readyUseSpell.Id, a.readyUseSpell.Name)
	if m.IsGhost() {
		logger.Warningln("Ghost can't SpellAttack")
		return nil
	}
	if m.haveStepsToGo() {
		m.Stop()
	}
	m.AttackAction()

	// 应用技能冷却时间乘数
	cdTime := time.Duration(spell.CdTime) * time.Millisecond
	if m.currentPhaseModifier != nil && m.currentPhaseModifier.SkillCooldownMultiplier > 0 {
		cdTime = time.Duration(float64(cdTime) * m.currentPhaseModifier.SkillCooldownMultiplier)
	}

	// 重置技能冷却时间
	spell.ResetCDTime()

	// 创建法术实体
	m.scene.CreateSpellEntity(m, spell, target)
	return nil
}
