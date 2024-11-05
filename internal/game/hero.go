package game

import (
	"fmt"
	"github.com/lonng/nano/cluster"
	"github.com/lonng/nano/session"
	"github.com/ouyangzhongmin/gameserver/db/model"
	"github.com/ouyangzhongmin/gameserver/internal/game/constants"
	"github.com/ouyangzhongmin/gameserver/internal/game/object"
	"github.com/ouyangzhongmin/gameserver/pkg/shape"
	"github.com/ouyangzhongmin/gameserver/protocol"
	"time"
)

type routeMsg struct {
	Route string
	Msg   interface{}
}

type Hero struct {
	*object.HeroObject
	movableEntity
	session                   *session.Session
	tracePath                 [][]int32 //当前移动路径
	traceIndex                int       //当前已移动到第几步
	traceTotalTime            int64     //当前移动的总时间
	targetX, targetY, targetZ int       //移动的目标点
	messagesCh                chan routeMsg
	destroyCh                 chan struct{}
}

func NewHero(s *session.Session, data *model.Hero) *Hero {
	h := &Hero{
		HeroObject: object.NewHeroObject(data),
		session:    s,
		messagesCh: make(chan routeMsg, 2048),
		destroyCh:  make(chan struct{}),
	}
	h.initEntity(h.HeroObject.Id, data.Name, constants.ENTITY_TYPE_HERO, 1024)
	h.GameObject.Uuid = h.GetUUID()
	go h.doMessageChFunc()
	return h
}

func (h *Hero) doMessageChFunc() {
	i := 0
	var ts int64 = 0
	for {
		select {
		case msg := <-h.messagesCh:
			if h.session != nil {
				ts = time.Now().UnixMilli()
				i = 0
				for ; i < 20; i++ {
					err := h.session.Push(msg.Route, msg.Msg)
					if err != nil {
						if err.Error() == cluster.ErrBufferExceed.Error() {
							//session send buffer exceed 同屏数量大时快速调用消息发送会导致消息发送失败，超出chan buffer范围
							if i >= 19 {
								logger.Errorf("hero: %s .SendMsg msg.Route:%s,msg:%s, useTime::%d, ErrBufferExceed err::: %v, %d \n", h._name, msg.Route, msg.Msg, time.Now().UnixMilli()-ts, err, i)
							}
							// todo 目前测试屏幕内1000以上还是会出现， 可以尝试修改源码的读写缓冲区和agent.chSend的size大小测试一下
							time.Sleep(30)
							continue
						} else {
							if i >= 3 {
								logger.Errorf("hero: %s .SendMsg msg.Route:%s,msg:%s  err::: %v, %d \n", h._name, msg.Route, msg.Msg, err, i)
								break
							}
							continue
						}
					}
					break
				}
				//logger.Debugf("hero: %s sendMsg msg.Route:%s,msg:%s, useTime::%d, retryCnt :%d", h._name, msg.Route, msg.Msg, time.Now().UnixMilli()-ts, i)
			} else {
				logger.Warningln("hero.SendMsg err: hero is offline", msg.Route, msg.Msg)
			}

		case <-h.destroyCh:
			return
		}
	}
}

func (h *Hero) onEnterScene(scene *Scene) {
	h.movableEntity.onEnterScene(scene)
	h.SceneId = scene.sceneId
	//刷新视野
	h.scene.addToBuildViewList(h)
}

func (h *Hero) onExitScene(scene *Scene) {
	h.movableEntity.onExitScene(scene)
}

func (h *Hero) SetPos(x, y, z shape.Coord) {
	oldx, oldy := h.GetPos().X, h.GetPos().Y
	h.Posx = x
	h.Posy = y
	h.Posz = z
	h.movableEntity.SetPos(x, y, z)
	if h.scene != nil {
		//更新block数据 go的继承关系是组合关系，这个逻辑如果写在movableEntity会导致存储的对象是*moveableEntity，并不是*Hero
		h.scene.entityMoved(h, oldx, oldy)
	}
}

func (h *Hero) SetViewRange(width int, height int) {
	h.movableEntity.SetViewRange(width, height)
	if h.scene != nil {
		//刷新视野
		h.scene.addToBuildViewList(h)
	}
}

func (h *Hero) GetID() int64 {
	return h.Id
}

func (h *Hero) GetUID() int64 {
	return h.Uid
}

func (h *Hero) GetData() *object.HeroObject {
	return h.HeroObject
}

func (h *Hero) GetAttack() int64 {
	h.Attack = constants.CaculateAttack(h.AttrType, h.BaseAttack, h.Strength, h.Agility, h.Intelligence)
	return h.Attack
}

func (h *Hero) GetDefense() int64 {
	h.Defense = constants.CaculateDefense(h.BaseDefense, h.Agility)
	return h.Defense
}

func (h *Hero) bindSession(s *session.Session) {
	h.session = s
	if h.session != nil {
		h.session.Set(constants.KCurHero, h)
	}
}

func (h *Hero) IsOffline() bool {
	return h.session == nil
}

func (h *Hero) save() {
	h.PushTask(func() {
		h.InitPosx = int(h.Posx)
		h.InitPosy = int(h.Posy)
		h.InitPosz = int(h.Posz)
		h.UpdateProperty()
		if h.scene != nil {
			h.SceneId = h.scene.sceneId
		}
		// todo保存数据，可以推入mq执行保存操作, 还可以细分保存的频度，部分数据需要立即存储
	})
}

func (h *Hero) onEnterView(target IMovableEntity) {
	h.movableEntity.onEnterView(target)
	var data interface{}
	var buffers []*object.BufferObject
	ttype := -1
	switch val := target.(type) {
	case *Hero:
		ttype = constants.ENTITY_TYPE_HERO
		buffers = val.GetBuffers()
		data = val.HeroObject
		//有对象进入自己的视野了，推送给前端创建对象
		h.SendMsg(protocol.OnEnterView, &protocol.TargetEnterViewResponse{
			EntityType: ttype,
			Data:       data,
			Buffers:    buffers,
		})
	case *Monster:
		ttype = constants.ENTITY_TYPE_MONSTER
		data = val.MonsterObject
		//buffers = val.GetBuffers()
		//有对象进入自己的视野了，推送给前端创建对象
		//logger.Debugf("monster::%d-%s 进入hero:%d视野\n", val.GetID(), val._name, h.GetID())
		h.SendMsg(protocol.OnEnterView, &protocol.TargetEnterViewResponse{
			EntityType: ttype,
			Data:       data,
			Buffers:    buffers,
		})
		val.sendDataToHero(h)
	case *SpellEntity:
		ttype = constants.ENTITY_TYPE_SPELL
		//有对象进入自己的视野了，推送给前端创建对象
		h.SendMsg(protocol.OnEnterView, &protocol.TargetEnterViewResponse{
			EntityType: ttype,
			Data:       data,
			Buffers:    buffers,
		})
	}
}

func (h *Hero) onExitView(target IMovableEntity) {
	h.movableEntity.onExitView(target)
	ttype := -1
	switch target.(type) {
	case *Hero:
		ttype = constants.ENTITY_TYPE_HERO
	case *Monster:
		ttype = constants.ENTITY_TYPE_MONSTER
	case *SpellEntity:
		ttype = constants.ENTITY_TYPE_SPELL
	}
	//logger.Debugf("对象:%d-%d离开hero:%d视野:", target.GetID(), ttype, h.GetID())
	if ttype > -1 {
		//有对象离开自己的视野了，推送给前端删除对象
		h.SendMsg(protocol.OnExitView, protocol.TargetExitViewResponse{
			EntityType: ttype,
			ID:         target.GetID(),
		})
	}
}

func (h *Hero) SendMsg(route string, msg interface{}) {
	h.PushTask(func() {
		if h.session != nil {
			h.messagesCh <- routeMsg{
				Route: route,
				Msg:   msg,
			}
			//logger.Debugf("hero:%s msgchan len::%d", h._name, len(h.messagesCh))
		} else {
			logger.Warningln("hero.SendMsg err: hero is offline", route, msg)
		}
	})
}

func (h *Hero) ToString() string {
	baseInfo := fmt.Sprintf("id:%d,uuid:%s, posX:%d, posY:%d, posZ:%d", h.GetID(), h.GetUUID(), h.GetPos().X, h.GetPos().Y, h.GetPos().Z)
	return fmt.Sprintf("baseInfo:%s,,,data::%v", baseInfo, h.Hero)
}

func (h *Hero) Destroy() {
	if h.scene != nil {
		h.scene.removeHero(h)
	}
	h.movableEntity.Destroy()
	if h.session != nil {
		h.session.Clear()
		h.session.Close()
		h.bindSession(nil)
	}
	close(h.destroyCh)
}

// 广播给所有能看见自己的对象
func (h *Hero) Broadcast(route string, msg interface{}, includeSelf bool) {
	if includeSelf {
		h.SendMsg(route, msg)
	}
	//这里如果放入task里执行，在退Destroy的时候要注意这个task不会执行了
	h.canSeeMeViewList.Range(func(key, value interface{}) bool {
		switch val := value.(type) {
		case *Hero:
			if val != h {
				val.SendMsg(route, msg)
			}
		}
		return true
	})
}

// 动作状态
func (h *Hero) SetState(state constants.ActionState) {
	h.State = state
}

func (h *Hero) GetState() constants.ActionState {
	return h.State
}

func (h *Hero) AttackAction() {
	h.SetState(constants.ACTION_STATE_ATTACK)
}

func (h *Hero) Idle() {
	h.SetState(constants.ACTION_STATE_IDLE)
}

func (h *Hero) Walk() {
	h.SetState(constants.ACTION_STATE_WALK)
}

func (h *Hero) Run() {
	h.SetState(constants.ACTION_STATE_RUN)
}

func (h *Hero) Die() {
	h.SetState(constants.ACTION_STATE_DIE)
	logger.Debugf("hero:%d-%s die", h.GetID(), h._name)
	h.Broadcast(protocol.OnEntityDie, &protocol.EntityDieResponse{
		ID:         h.GetID(),
		EntityType: constants.ENTITY_TYPE_HERO,
	}, true)
}

// update都会在chTask携程内执行
func (h *Hero) update(curMilliSecond int64, elapsedTime int64) error {
	err := h.movableEntity.update(curMilliSecond, elapsedTime)
	if h.haveStepsToGo() {
		h.updateHeroPosition(curMilliSecond, elapsedTime)
	}
	return err
}

func (h *Hero) updateHeroPosition(curMilliSecond int64, elapsedTime int64) {
	h.traceTotalTime += elapsedTime
	stepTime := h.StepTime
	//当前移动总时间/速度得到移动到了第几步
	h.traceIndex = int(h.traceTotalTime / int64(stepTime))
	if h.traceIndex >= len(h.tracePath) {
		//移动到了目标点
		h.clearTracePaths()
		return
	}
	//根据路径预测更新用户在的位置
	step := h.tracePath[h.traceIndex]
	//寻路返回的0是y坐标，1是X坐标，注意了
	h.SetPos(shape.Coord(step[1]), shape.Coord(step[0]), 0)
}

func (h *Hero) haveStepsToGo() bool {
	return h.tracePath != nil && len(h.tracePath) > 0
}

func (h *Hero) clearTracePaths() {
	h.tracePath = nil
	h.traceIndex = 0
	h.traceTotalTime = 0
	h.targetX = 0
	h.targetY = 0
	h.targetZ = 0
}

// 前端移动到目标位置
func (h *Hero) MoveByPaths(targetx, targety, targetz int, paths [][]int32) error {
	h.PushTask(func() {
		logger.Debugf("hero:%s moveByPaths:%v", h._name, paths)
		if h.scene == nil {
			return
		}
		if h.targetX == targetx && h.targetY == targety {
			//目标点一致,什么都不做
			return
		}
		// todo 这里需要校验当前服务器的位置与前端的差距是否合理范围
		h.tracePath = paths
		h.traceIndex = 0
		h.traceTotalTime = 0
		h.targetX = targetx
		h.targetY = targety
		h.targetZ = targetz
		if len(paths) > 0 {
			firstStep := paths[0]
			if h.GetPos().X != shape.Coord(firstStep[1]) || h.GetPos().Y != shape.Coord(firstStep[0]) {
				h.SetPos(shape.Coord(firstStep[1]), shape.Coord(firstStep[0]), h.GetPos().Z)
				//todo 这里看是否需要调用scene.refreshEntityViewList立即刷新视野
			}
		}
		h.Broadcast(protocol.OnHeroMoveTrace, &protocol.HeroMoveTraceResponse{
			ID:         h.GetID(),
			TracePaths: paths,
			StepTime:   h.StepTime,
			PosX:       h.GetPos().X,
			PosY:       h.GetPos().Y,
			PosZ:       h.GetPos().Z,
		}, false)
	})
	return nil
}

func (h *Hero) MoveStop(x, y, z shape.Coord) error {
	h.PushTask(func() {
		h.clearTracePaths()
		h.SetPos(x, y, z)
		h.Broadcast(protocol.OnHeroMoveStopped, &protocol.HeroMoveStopResponse{
			ID:   h.GetID(),
			PosX: h.GetPos().X,
			PosY: h.GetPos().Y,
			PosZ: h.GetPos().Z,
		}, false)
	})
	return nil
}

func (h *Hero) AttackTarget(targetId int64) error {
	return nil
}

func (h *Hero) BeenAttacked(damage int64) (int, error) {

	return 0, nil
}

func (h *Hero) CanAttackTarget(target IEntity) bool {
	return true
}

func (h *Hero) onBeenHurt(damage int64) {
	h.PushTask(func() {
		if !h.IsAlive() {
			logger.Warningln("hero is dead")
			return
		}
		h.Life -= damage
		if h.Life < 0 {
			h.Life = 0
		}
		if h.Life > h.MaxLife {
			h.Life = h.MaxLife
		}
		h.Broadcast(protocol.OnLifeChanged, &protocol.LifeChangedResponse{
			ID:         h.GetID(),
			EntityType: constants.ENTITY_TYPE_HERO,
			Damage:     damage,
			Life:       h.Life,
			MaxLife:    h.MaxLife,
		}, true)
		if h.Life <= 0 {
			h.Die()
			//死亡了
		}
	})
}

func (h *Hero) onBeenAttacked(target IMovableEntity) {
}

func (h *Hero) manaCost(mana int64) {
	h.PushTask(func() {
		if !h.IsAlive() {
			logger.Warningln("hero is dead")
			return
		}
		h.Mana -= mana
		if h.Mana < 0 {
			h.Mana = 0
		}
		if h.Mana > h.MaxMana {
			h.Mana = h.MaxMana
		}
		h.Broadcast(protocol.OnManaChanged, &protocol.ManaChangedResponse{
			ID:         h.GetID(),
			EntityType: constants.ENTITY_TYPE_HERO,
			Cost:       mana,
			Mana:       h.Mana,
			MaxMana:    h.MaxMana,
		}, true)
	})
}
