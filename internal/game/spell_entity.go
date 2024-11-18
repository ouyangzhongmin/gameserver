package game

import (
	"errors"
	"github.com/ouyangzhongmin/gameserver/constants"
	"github.com/ouyangzhongmin/gameserver/internal/game/object"
	"github.com/ouyangzhongmin/gameserver/pkg/shape"
	"github.com/ouyangzhongmin/gameserver/protocol"
)

type SpellEntity struct {
	*object.SpellObject
	movableEntity
	caster      IMovableEntity
	target      IMovableEntity
	elapsedTime int64
	totalTime   int64
}

func NewSpellEntity(spellObject *object.SpellObject, caster IMovableEntity) *SpellEntity {
	e := &SpellEntity{
		caster:      caster,
		SpellObject: spellObject,
	}
	e.initEntity(int64(e.SpellObject.Id), "spell"+spellObject.Name, constants.ENTITY_TYPE_SPELL, 64)
	e.GameObject.Uuid = e.GetUUID()
	e.GenId()
	e.CasterId = e.caster.GetID()
	e.CasterType = e.caster.GetEntityType()
	e.SetPos(e.caster.GetPos().X, e.caster.GetPos().Y, e.caster.GetPos().Z)
	e.SetViewRange(e.caster.GetViewRange())
	e.casterManaCost()
	return e
}

func (e *SpellEntity) onEnterScene(scene *Scene) {
	e.movableEntity.onEnterScene(scene)

	//广播创建对象事件
	switch val := e.caster.(type) {
	case *Hero:
		val.Broadcast(protocol.OnReleaseSpell, &protocol.ReleaseSpellResponse{
			SpellObject: e.SpellObject,
		}, false)
	case *Monster:
		val.Broadcast(protocol.OnReleaseSpell, &protocol.ReleaseSpellResponse{
			SpellObject: e.SpellObject,
		})
	}
}

func (e *SpellEntity) onExitScene(scene *Scene) {
	e.movableEntity.onExitScene(scene)
}

func (e *SpellEntity) SetPos(x, y, z shape.Coord) {
	e.Posx = x
	e.Posy = y
	e.Posz = z
	e.movableEntity.SetPos(x, y, z)
}

// 需要在添加到场景内之前执行
func (e *SpellEntity) SetTarget(target IMovableEntity) {
	e.target = target
	e.TargetId = e.target.GetID()
	e.TargetType = e.target.GetEntityType()
	e.SetTargetPos(target.GetPos())
}

// 需要在添加到场景内之前执行
func (e *SpellEntity) SetTargetPos(target shape.Vector3) {
	e.TargetPos.Copy(target)

	// 计算两点距离 再算出需要移动的总时间
	dist := int(shape.CalculateDistance(float64(e.caster.GetPos().X), float64(e.caster.GetPos().Y), float64(e.TargetPos.X), float64(e.TargetPos.Y)))
	//重新计算距离的时候需要把已行走的时间叠加起来
	e.totalTime = e.elapsedTime + int64(dist*e.FlyStepTime)
}

func (e *SpellEntity) update(curMilliSecond int64, elapsedTime int64) error {
	if e.scene == nil {
		return nil
	}
	if e.target != nil && (e.TargetPos.X != e.target.GetPos().X || e.TargetPos.Y != e.target.GetPos().Y) {
		//需要跟随对象
		e.SetTargetPos(e.target.GetPos())
	}
	err := e.movableEntity.update(curMilliSecond, elapsedTime)
	e.elapsedTime += elapsedTime
	if e.elapsedTime >= e.totalTime {
		//到达消失时间
		var err error
		if e.Data.IsRangeAttack != 0 {
			entityes := e.scene.getEntitiesByRange(e.TargetPos.X, e.TargetPos.Y, shape.Coord(e.Data.AttackRange))
			if entityes != nil && len(entityes) > 0 {
				for _, entity := range entityes {
					if entity == e.caster {
						continue
					}
					canAttack := false
					switch val := e.caster.(type) {
					case *Hero:
						canAttack = val.CanAttackTarget(entity)
					case *Monster:
						canAttack = val.CanAttackTarget(entity)
					}
					if !canAttack {
						continue
					}
					err = e.processTargetHurt(entity)
					if err != nil {
						return err
					}
				}
			}
		} else {
			if e.target == nil {
				return errors.New("非单体技能但是没有指定释放对象？")
			}
			return e.processTargetHurt(e.target)
		}
		e.Destroy()
	}
	return err
}

func (e *SpellEntity) processTargetHurt(target IMovableEntity) error {
	if e.Data.Damage != 0 {
		var damage int64 = 0
		if e.Data.Damage > 0 {
			var defense int64 = 0
			switch val := target.(type) {
			case *Hero:
				defense = val.GetDefense()
			case *Monster:
				defense = val.GetDefense()
			}
			damage = e.Data.Damage - defense
			if damage < 1 { //至少有1点伤害
				damage = 1
			}
		} else { //加血不用计算防御力
			damage = e.Data.Damage
		}

		switch val := target.(type) {
		case *Hero:
			val.onBeenHurt(damage)
		case *Monster:
			val.onBeenHurt(damage)
		}
	}
	return e.processBufferState(target)
}

func (e *SpellEntity) processBufferState(target IMovableEntity) error {
	if e.Buf != nil {
		switch val := target.(type) {
		case *Hero:
			val.addBuffer(val, e.Buf)
		case *Monster:
			val.addBuffer(val, e.Buf)
		}
	}
	return nil
}

func (e *SpellEntity) casterManaCost() {
	e.PushTask(func() {
		switch val := e.caster.(type) {
		case *Hero:
			val.manaCost(e.Data.Mana)
		case *Monster:
			val.manaCost(e.Data.Mana)
		}
	})
}

func (e *SpellEntity) Destroy() {
	e.caster = nil
	e.target = nil
	if e.scene != nil {
		e.scene.removeSpell(e)
	}
	e.movableEntity.Destroy()
}
