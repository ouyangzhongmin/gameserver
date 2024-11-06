package game

import (
	"github.com/ouyangzhongmin/gameserver/db/model"
	"github.com/ouyangzhongmin/gameserver/internal/game/constants"
	"github.com/ouyangzhongmin/gameserver/pkg/shape"
	"math/rand"
	"time"
)

type monsterai struct {
	aidata        *model.Aiconfig
	monster       *Monster
	originX       shape.Coord
	originY       shape.Coord
	behaviorState constants.BEHAVIOR
	preparePathId int

	nextBehaviorTime   int64
	nextRandomMoveTime int64
	nextAttackTime     int64

	enemy IMovableEntity
}

func newMonsterAi(m *Monster, aidata *model.Aiconfig) *monsterai {
	a := &monsterai{}
	a.monster = m
	a.aidata = aidata
	a.behaviorState = constants.BEHAVIOR_STATE_IDLE
	a.refreshNextBehaviorTime()
	a.refreshNextRandomMoveTime()
	return a
}

func (a *monsterai) GetAiData() interface{} {
	return a.aidata
}

func (a *monsterai) GetOwner() IMovableEntity {
	return a.monster
}

func (a *monsterai) update(curMilliSecond int64, elapsedTime int64) error {
	if curMilliSecond < a.nextBehaviorTime {
		return nil
	}
	defer func() {
		a.refreshNextBehaviorTime()
	}()
	var err error
	switch a.behaviorState {
	case constants.BEHAVIOR_STATE_IDLE:
		err = a.processIdleState(curMilliSecond, elapsedTime)
	case constants.BEHAVIOR_STATE_ATTACK:
		err = a.processAttackState(curMilliSecond, elapsedTime)
	case constants.BEHAVIOR_STATE_RETURN:
		err = a.processReturnState(curMilliSecond, elapsedTime)
	default:
		return nil
	}
	return err
}

func (a *monsterai) processIdleState(curMilliSecond int64, elapsedTime int64) error {
	if a.monster.State == constants.ACTION_STATE_IDLE {
		rect := a.monster.GetMovableRect()
		if !rect.Contains(int64(a.monster.GetPos().X), int64(a.monster.GetPos().Y)) {
			//返回原点
			return a.backOrigin()
		}
		enemy := a.scanEnemy()
		if enemy != nil {
			a.setEnemy(enemy)
			return nil
		}
		if a.nextRandomMoveTime < curMilliSecond {
			a.refreshNextRandomMoveTime()
			rd := rand.Intn(100)
			if rd < 5 { // 5%概率
				//这里的ai随机位置，可以改成通过预制固定的寻路路径，并将寻路路径保存为文件载入，这样可以减少在游戏内的动态Astar
				if a.monster.preparePaths != nil && len(a.monster.preparePaths.Paths) > 0 {
					a.preparePathId = a.preparePathId % len(a.monster.preparePaths.Paths)
					paths := a.monster.preparePaths.Paths[a.preparePathId]
					//logger.Debugf("monster:%d 使用预制路径:%d移动:%v", a.monster.GetID(), a.preparePathId, paths)
					a.monster.SetState(constants.ACTION_STATE_WALK)
					a.preparePathId += 1
					return a.monster.MoveByPaths(paths.Paths)
				} else {
					rx, ry, err := a.monster.scene.GetRandomXY(rect, 20)
					if err != nil {
						return err
					}
					a.monster.SetState(constants.ACTION_STATE_WALK)
					//logger.Debugf("monster:%d, %d,%d walk to :%d,%d, cur is walkable :%v \n", a.monster.GetID(), a.monster.GetPosX(), a.monster.GetPosY(), rx, ry, a.monster.scene.blockInfo.IsWalkable(int32(a.monster.GetPosX()), int32(a.monster.GetPosY())))
					return a.monster.MoveTo(rx, ry, 0)
				}
			}
		}
	}
	return nil
}

func (a *monsterai) processAttackState(curMilliSecond int64, elapsedTime int64) error {
	needClearEnemy := a.enemy == nil
	if a.enemy != nil {
		switch val := a.enemy.(type) {
		case *Hero:
			if val.IsOffline() || !val.IsAlive() || val.IsDestroyed() {
				needClearEnemy = true
			}
		case *Monster:
			if !val.IsAlive() || val.IsDestroyed() {
				needClearEnemy = true
			}
		}
	}
	if needClearEnemy {
		a.enemy = nil
		if a.monster.haveStepsToGo() {
			a.monster.Stop()
		}
		//返回原点
		return a.backOrigin()
	}
	if a.monster.State != constants.ACTION_STATE_ATTACK {
		if a.monster.State == constants.ACTION_STATE_CHASE {
			//追击过程，看是否超出边界范围了
			rect := a.monster.GetMovableRect()
			if !rect.Contains(int64(a.monster.GetPos().X), int64(a.monster.GetPos().Y)) {
				//返回原点
				return a.backOrigin()
			}
		}

		if a.monster.IsInAttackRange(a.enemy.GetPos().X, a.enemy.GetPos().Y) {
			//如果在攻击范围
			if a.nextAttackTime <= curMilliSecond {
				a.attackEnemy()
			}
		} else {
			//走到敌人附近去
			if !a.monster.haveStepsToGo() {
				tpos, err := a.monster.GetCanAttackPos(a.enemy, 1)
				if err != nil {
					return err
				}
				a.monster.SetState(constants.ACTION_STATE_CHASE)
				if a.monster.preparePaths != nil && len(a.monster.preparePaths.Paths) > 0 {
					//要回到预制路线的起点上去
					a.preparePathId = a.preparePathId % len(a.monster.preparePaths.Paths)
					paths := a.monster.preparePaths.Paths[a.preparePathId]
					a.originX = shape.Coord(paths.Sx)
					a.originY = shape.Coord(paths.Sy)
				} else {
					a.originX = a.monster.GetPos().X
					a.originY = a.monster.GetPos().Y
					if !a.monster.GetMovableRect().Contains(int64(a.originX), int64(a.originY)) {
						//超出范围了，回到出生点
						a.originX = a.monster.bornPos.X
						a.originY = a.monster.bornPos.Y
					}
				}
				//logger.Debugf("monster:%d 设置原点:%d,%d \n", a.monster.GetID(), a.originX, a.originY)
				return a.monster.MoveTo(tpos.X, tpos.Y, 0)
			}
		}
	} else {
		//正在攻击
		if a.nextAttackTime <= curMilliSecond {
			//攻击结束，进入站立状态, 下次update将进入上面的attackEnemy逻辑
			a.monster.SetState(constants.ACTION_STATE_IDLE)
		}
	}
	return nil
}

func (a *monsterai) processReturnState(curMilliSecond int64, elapsedTime int64) error {
	if a.monster.GetPos().X == a.originX && a.monster.GetPos().Y == a.originY {
		//回到原点后恢复到idle状态
		a.behaviorState = constants.BEHAVIOR_STATE_IDLE
		return nil
	}
	if !a.monster.haveStepsToGo() {
		return a.backOrigin()
	}
	return nil
}

func (a *monsterai) attackEnemy() {
	logger.Debugf("monster:%d attack enemy:%d-%d \n", a.monster.GetID(), a.enemy.GetID(), a.enemy.GetEntityType())
	if a.monster.haveStepsToGo() {
		a.monster.Stop()
	}
	a.monster.doAttackTarget(a.enemy)
	a.refreshNextAttackTime()
}

func (a *monsterai) scanEnemy() IMovableEntity {
	entities := a.monster.scene.getEntitiesByRange(a.monster.GetPos().X, a.monster.GetPos().Y, shape.Coord(a.aidata.AlertRange))
	var dist float64 = 10000000
	var enemy IMovableEntity = nil
	if entities != nil && len(entities) > 0 {
		for _, e := range entities {
			if e == a.monster {
				continue
			}
			if !a.monster.CanAttackTarget(e) {
				continue
			}
			tmpDist := shape.CalculateDistance(float64(a.monster.GetPos().X), float64(a.monster.GetPos().Y), float64(e.GetPos().X), float64(e.GetPos().Y))
			if tmpDist < dist {
				dist = tmpDist
				enemy = e
			}
		}
	}
	return enemy
}

func (a *monsterai) backOrigin() error {
	if !a.monster.haveStepsToGo() {
		a.behaviorState = constants.BEHAVIOR_STATE_RETURN
		a.monster.SetState(constants.ACTION_STATE_RUN)
		if !a.monster.GetMovableRect().Contains(int64(a.originX), int64(a.originY)) || (a.originX == 0 && a.originY == 0) {
			//超出范围了，回到出生点
			a.originX = a.monster.bornPos.X
			a.originY = a.monster.bornPos.Y
		}
		logger.Debugf("monster:%d_%s返回原点:%d,%d \n", a.monster.GetID(), a.monster._name, a.originX, a.originY)
		//这里由于是中途被打断的，只能使用寻路回到原点去
		return a.monster.MoveTo(a.originX, a.originY, 0)
	}
	return nil
}

func (a *monsterai) refreshNextBehaviorTime() {
	a.nextBehaviorTime = time.Now().UnixMilli() + 500
}

func (a *monsterai) refreshNextAttackTime() {
	a.nextAttackTime = time.Now().UnixMilli() + int64(a.monster.AttackDuration) + 50
}

func (a *monsterai) refreshNextRandomMoveTime() {
	a.nextRandomMoveTime = time.Now().UnixMilli() + 5000
}

func (a *monsterai) onBeenAttacked(target IMovableEntity) {
	if !a.monster.CanAttackTarget(target) {
		return
	}
	if a.aidata.AutoBeatback == 0 {
		logger.Debugln("配置不自动反击")
		return
	}
	if a.enemy != nil {
		return
	}
	a.setEnemy(target)
}

func (a *monsterai) setEnemy(target IMovableEntity) {
	a.enemy = target
	//被攻击停下来
	if a.monster.haveStepsToGo() {
		a.monster.Stop()
	}
	a.behaviorState = constants.BEHAVIOR_STATE_ATTACK
}
