package object

const (
	LIFE_STRENGTH_PERM     = 15
	MANA_INTELLIGENCE_PERM = 10
	ATTACK_ATTR_PERM       = 2
	DEFENSE_AGILITY_PERM   = 2
)

func CaculateLife(baseLife, strength int64) int64 {
	return baseLife + strength*LIFE_STRENGTH_PERM
}

func CaculateMana(baseMana, intelligence int64) int64 {
	return baseMana + intelligence*MANA_INTELLIGENCE_PERM
}

func CaculateAttack(attrType int, baseAttack, strength, agility, intelligence int64) int64 {
	if attrType == 1 {
		return baseAttack + agility*ATTACK_ATTR_PERM
	} else if attrType == 2 {
		return baseAttack + intelligence*ATTACK_ATTR_PERM
	}
	return baseAttack + strength*ATTACK_ATTR_PERM
}

func CaculateDefense(baseDefense, agility int64) int64 {
	return baseDefense + agility*DEFENSE_AGILITY_PERM
}
