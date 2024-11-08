package game

import (
	"errors"
	"github.com/ouyangzhongmin/gameserver/internal/game/constants"
	"github.com/ouyangzhongmin/gameserver/internal/game/object"
	"github.com/ouyangzhongmin/gameserver/pkg/shape"
)

type SpellEntity struct {
	*object.SpellObject
	movableEntity
	caster      IEntity
	target      IEntity
	elapsedTime int64
	totalTime   int64
}

func NewSpellEntity(spellObject *object.SpellObject, caster IEntity) *SpellEntity {
	e := &SpellEntity{
		SpellObject: spellObject,
	}
	e.initEntity(int64(e.SpellObject.Id), "spell", constants.ENTITY_TYPE_SPELL, 64)
	e.GameObject.Uuid = e.GetUUID()
	return e
}

func (e *SpellEntity) onEnterScene(scene *Scene) {
	e.movableEntity.onEnterScene(scene)
	//更新block数据
	e.scene.addToBuildViewList(e)
}

func (e *SpellEntity) onExitScene(scene *Scene) {
	e.movableEntity.onExitScene(scene)
}

func (e *SpellEntity) SetPos(x, y, z shape.Coord) {
	oldx, oldy := e.GetPos().X, e.GetPos().Y
	e.Posx = x
	e.Posy = y
	e.Posz = z
	e.movableEntity.SetPos(x, y, z)
	if e.scene != nil {
		//更新block数据 go的继承关系是组合关系，这个逻辑如果写在movableEntity会导致存储的对象是*moveableEntity，并不是*Hero
		e.scene.entityMoved(e, oldx, oldy)
	}
}

// 需要在添加到场景内之前执行
func (e *SpellEntity) SetTarget(target IEntity) {
	e.target = target
	e.SetTargetPos(target.GetPos())
}

// 需要在添加到场景内之前执行
func (e *SpellEntity) SetTargetPos(target shape.Vector3) {
	e.TargetPos = target

	// 计算两点距离 再算出需要移动的总时间
	dist := int(shape.CalculateDistance(float64(e.caster.GetPos().X), float64(e.caster.GetPos().Y), float64(e.TargetPos.X), float64(e.TargetPos.Y)))
	e.totalTime = int64(dist * e.StepTime)
}

func (e *SpellEntity) update(curMilliSecond int64, elapsedTime int64) error {
	err := e.movableEntity.update(curMilliSecond, elapsedTime)
	e.elapsedTime += elapsedTime
	if e.elapsedTime >= e.totalTime {
		//到达消失时间
		var err error
		if e.Data.IsRangeAttack != 0 {
			entityes := e.scene.getEntitiesByRange(e.TargetPos.X, e.TargetPos.Y, shape.Coord(e.Data.AttackRange))
			if entityes != nil && len(entityes) > 0 {
				for _, entity := range entityes {
					switch val := e.caster.(type) {
					case *Hero:
						if !val.CanAttackTarget(entity) {
							continue
						}
					case *Monster:
						if !val.CanAttackTarget(entity) {
							continue
						}
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
		e.casterManaCost()
		e.Destroy()
	}
	return err
}

func (e *SpellEntity) processTargetHurt(target IEntity) error {
	if e.Data.Damage != 0 {
		switch val := target.(type) {
		case *Hero:
			val.onBeenHurt(e.Data.Damage)
		case *Monster:
			val.onBeenHurt(e.Data.Damage)
		}
	}
	return e.processBufferState(target)
}

func (e *SpellEntity) processBufferState(target IEntity) error {
	if e.Buf != nil {

	}
	return nil
}

func (e *SpellEntity) casterManaCost() {
	switch val := e.caster.(type) {
	case *Hero:
		val.manaCost(e.Data.Mana)
	case *Monster:
		val.manaCost(e.Data.Mana)
	}
}

func (e *SpellEntity) Destroy() {
	e.caster = nil
	e.target = nil
	if e.scene != nil {
		e.scene.removeSpell(e)
	}
	e.viewList.Range(func(key, value interface{}) bool {
		value.(IMovableEntity).onExitOtherView(e)
		return true
	})
	e.canSeeMeViewList.Range(func(key, value interface{}) bool {
		value.(IMovableEntity).onExitView(e)
		return true
	})
	e.movableEntity.Destroy()
}
