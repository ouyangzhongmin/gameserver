package db

import (
	"github.com/ouyangzhongmin/gameserver/db/model"
	"github.com/ouyangzhongmin/gameserver/pkg/errutil"
)

func InsertHero(h *model.Hero) (int64, error) {
	if h == nil {
		return 0, errutil.ErrInvalidParameter
	}
	row, err := database.Exec(`insert into hero(name,avatar,attr_type,uid,experience,level,max_life,max_mana,
                 defense,attack,base_life,base_mana,base_defense,base_attack,strength,agility,
                 intelligence,step_time,scene_id,attack_range) 
					values (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		h.Name, h.Avatar, h.AttrType, h.Uid, h.Experience, h.Level, h.MaxLife, h.MaxMana,
		h.Defense, h.Attack, h.BaseLife, h.BaseMana, h.BaseDefense, h.BaseAttack, h.Strength, h.Agility,
		h.Intelligence, h.StepTime, h.SceneId, h.AttackRange)
	if err != nil {
		return 0, err
	}
	return row.LastInsertId()
}

func UpdateHero(d *model.Hero) error {
	if d == nil {
		return nil
	}
	_, err := database.Where("id=?", d.Id).AllCols().Update(d)
	return err
}

func QueryHero(id int64) (*model.Hero, error) {
	h := &model.Hero{Id: id}
	has, err := database.Get(h)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errutil.ErrNotFound
	}
	return h, nil
}

func DeleteHero(id int64) error {
	_, err := database.Exec("UPDATE `hero` SET `status` = 0 WHERE `id`= ? ", id)
	if err != nil {
		return err
	}
	return nil
}

func HeroList(uid int64) ([]model.Hero, error) {
	result := make([]model.Hero, 0)
	err := database.Where("uid=?", uid).Desc("id").Find(&result)

	if err != nil {
		return nil, errutil.ErrDBOperation
	}
	return result, nil
}
