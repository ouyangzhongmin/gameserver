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
		Name:        name,
		Avatar:      avatar,
		AttrType:    0,
		Uid:         uid,
		Experience:  0,
		Level:       1,
		BaseLife:    100000,
		BaseMana:    1000,
		StepTime:    300,
		SceneId:     sceneId,
		AttackRange: 3,
	}
	if attrType == constants.ATTR_TYPE_STRENGTH {
		h.BaseDefense = 5
		h.BaseAttack = 22
		h.Strength = 28
		h.Agility = 22
		h.Intelligence = 20
	} else if attrType == constants.ATTR_TYPE_AGILITY {
		h.BaseDefense = 5
		h.BaseAttack = 22
		h.Strength = 22
		h.Agility = 30
		h.Intelligence = 20
	} else {
		h.BaseDefense = 5
		h.BaseAttack = 22
		h.Strength = 18
		h.Agility = 18
		h.Intelligence = 35
	}
	h.MaxLife = constants.CaculateLife(h.BaseLife, h.Strength)
	h.MaxMana = constants.CaculateLife(h.BaseMana, h.Intelligence)
	h.Attack = constants.CaculateAttack(h.AttrType, h.BaseAttack, h.Strength, h.Agility, h.Intelligence)
	h.Defense = constants.CaculateDefense(h.Defense, h.Agility)
	return h
}
