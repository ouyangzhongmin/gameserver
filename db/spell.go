package db

import (
	"github.com/ouyangzhongmin/gameserver/db/model"
	"github.com/ouyangzhongmin/gameserver/pkg/errutil"
)

func QuerySpell(id int) (*model.Spell, error) {
	h := &model.Spell{Id: id}
	has, err := database.Get(h)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errutil.ErrNotFound
	}
	return h, nil
}

func QueryBufferState(id int) (*model.BufferState, error) {
	h := &model.BufferState{Id: id}
	has, err := database.Get(h)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errutil.ErrNotFound
	}
	return h, nil
}
