package master

import (
	"github.com/ouyangzhongmin/gameserver/db/model"
	"github.com/ouyangzhongmin/gameserver/internal/game/constants"
	"github.com/ouyangzhongmin/gameserver/pkg/utils"
	"github.com/ouyangzhongmin/gameserver/protocol"
)

func BroadcastSystemMessage(message string) {
	defaultManager.group.Broadcast("onBroadcast", &protocol.StringMessage{Message: message})
}

func Kick(uid int64) error {
	defaultManager.chKick <- uid
	return nil
}

func Reset(uid int64) {
	defaultManager.chReset <- uid
}

func Recharge(uid, coin int64) {
	defaultManager.chRecharge <- RechargeInfo{uid, coin}
}

func createRandomHero(uid int64, sceneId int, name, avatar string, attrType int) *model.Hero {
	if name == "" {
		name = utils.GetRandomHeroName()
	}
	h := &model.Hero{
		Name:       name,
		Avatar:     avatar,
		AttrType:   0,
		Uid:        uid,
		Experience: 0,
		Level:      1,
		BaseLife:   100,
		BaseMana:   100,

		StepTime:    300,
		SceneId:     sceneId,
		AttackRange: 10,
	}
	if attrType == constants.ATTR_TYPE_STRENGTH {
		h.BaseDefense = 5
		h.BaseAttack = 12
		h.Strength = 8
		h.Agility = 5
		h.Intelligence = 6
	} else if attrType == constants.ATTR_TYPE_AGILITY {
		h.BaseDefense = 5
		h.BaseAttack = 12
		h.Strength = 6
		h.Agility = 10
		h.Intelligence = 5
	} else {
		h.BaseDefense = 5
		h.BaseAttack = 12
		h.Strength = 5
		h.Agility = 6
		h.Intelligence = 10
	}
	h.MaxLife = constants.CaculateLife(h.BaseLife, h.Strength)
	h.MaxMana = constants.CaculateLife(h.BaseMana, h.Intelligence)
	h.Attack = constants.CaculateAttack(h.AttrType, h.BaseAttack, h.Strength, h.Agility, h.Intelligence)
	h.Defense = constants.CaculateDefense(h.Defense, h.Agility)
	return h
}
