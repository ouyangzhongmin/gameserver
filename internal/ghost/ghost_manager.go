package ghost

import (
	"errors"
	"fmt"
	"github.com/ouyangzhongmin/gameserver/db"
	"github.com/ouyangzhongmin/gameserver/protocol"
	"github.com/ouyangzhongmin/nano/component"
	"github.com/ouyangzhongmin/nano/session"
)

type GhostManager struct {
	component.Base
	scenes   map[int]*ghostScene
	sceneIds []int
}

var defaultGhostManager = newGhostManager()

func newGhostManager() *GhostManager {
	return &GhostManager{
		scenes:   make(map[int]*ghostScene),
		sceneIds: make([]int, 0),
	}
}

func (manager *GhostManager) setSceneIds(sceneIds []int) {
	manager.sceneIds = append(manager.sceneIds, sceneIds...)
}

func (manager *GhostManager) AfterInit() {
	scenes, err := db.SceneList(manager.sceneIds)
	if err != nil {
		panic(err)
	}
	for _, sceneData := range scenes {
		scene := newGhostScene(&sceneData)
		manager.scenes[sceneData.Id] = scene
	}
}

// 同步Ghost对象
func (manager *GhostManager) SyncGhostEntity(s *session.Session, e *protocol.GhostEntity) error {
	scene := manager.scenes[e.SceneID]
	if scene == nil {
		return errors.New(fmt.Sprintf("ghost secene:%d is not exit!", e.SceneID))
	}
	scene.updateEntity(e)
	return nil
}

func (manager *GhostManager) RemoveGhostEntity(s *session.Session, e *protocol.GhostEntitySimple) error {
	scene := manager.scenes[e.SceneID]
	if scene == nil {
		return errors.New(fmt.Sprintf("ghost secene:%d is not exit!", e.SceneID))
	}
	le := scene.getEntity(e.Uuid)
	if le == nil {
		return errors.New(fmt.Sprintf("ghost entity:%d is not exit!", e.SceneID))
	} else {
		scene.removeEntity(e.Uuid)
	}
	return nil
}

func (manager *GhostManager) AddToBuildViewList(s *session.Session, e *protocol.GhostEntitySimple) error {
	scene := manager.scenes[e.SceneID]
	if scene == nil {
		return errors.New(fmt.Sprintf("ghost secene:%d is not exit!", e.SceneID))
	}
	le := scene.getEntity(e.Uuid)
	if le == nil {
		return errors.New(fmt.Sprintf("ghost entity:%d is not exit!", e.SceneID))
	} else {
		scene.addToBuildViewList(le)
	}
	return nil
}
