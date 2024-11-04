package db

import (
	"github.com/ouyangzhongmin/gameserver/db/model"
	"github.com/ouyangzhongmin/gameserver/pkg/errutil"
)

func QueryMonster(id int64) (*model.Monster, error) {
	m := &model.Monster{Id: id}
	has, err := database.Get(m)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errutil.ErrNotFound
	}
	return m, nil
}

func QueryAiConfig(mid int64) (*model.Aiconfig, error) {
	m := &model.Aiconfig{MonsterId: mid}
	has, err := database.Get(m)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errutil.ErrNotFound
	}
	return m, nil
}
