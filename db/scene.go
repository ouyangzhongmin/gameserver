package db

import (
	"github.com/ouyangzhongmin/gameserver/db/model"
	"github.com/ouyangzhongmin/gameserver/pkg/errutil"
)

func QueryScene(id int) (*model.Scene, error) {
	h := &model.Scene{Id: id}
	has, err := database.Get(h)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errutil.ErrNotFound
	}
	return h, nil
}

func SceneList() ([]model.Scene, error) {
	bean := &model.Scene{}

	list := []model.Scene{}
	if err := database.Find(&list, bean); err != nil {
		return nil, err
	}

	if len(list) < 1 {
		return []model.Scene{}, nil
	}

	return list, nil
}

func SceneDoorList(sceneId int) ([]model.SceneDoor, error) {
	result := make([]model.SceneDoor, 0)
	if err := database.Where("scene_id=?", sceneId).Find(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func SceneMonsterConfigList(sceneId int) ([]model.SceneMonsterConfig, error) {
	result := make([]model.SceneMonsterConfig, 0)
	if err := database.Where("scene_id=?", sceneId).Find(&result); err != nil {
		return nil, err
	}
	return result, nil
}
