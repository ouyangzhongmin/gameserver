package game

import (
	"errors"
	"github.com/ouyangzhongmin/gameserver/db"
	"github.com/ouyangzhongmin/gameserver/protocol"
	"time"

	"github.com/lonng/nano/component"
	"github.com/lonng/nano/session"
)

const (
	fieldDesk = "desk"
)

type (
	SceneManager struct {
		component.Base
		scenes map[int]*Scene
	}
)

var defaultSceneManager = NewSceneManager()

func NewSceneManager() *SceneManager {
	return &SceneManager{
		scenes: make(map[int]*Scene),
	}
}

func (manager *SceneManager) AfterInit() {
	session.Lifetime.OnClosed(func(s *session.Session) {
		// Fixed: 玩家WIFI切换到4G网络不断开, 重连时，将UID设置为illegalSessionUid
		if s.UID() > 0 {
			if err := manager.onPlayerDisconnect(s); err != nil {
				logger.Errorf("玩家退出: UID=%d, Error=%s", s.UID, err.Error())
			}
		}
	})

	scenes, err := db.SceneList()
	if err != nil {
		panic(err)
	}
	for _, sceneData := range scenes {
		doorList, err := db.SceneDoorList(sceneData.Id)
		if err != nil {
			panic(err)
		}
		configList, err := db.SceneMonsterConfigList(sceneData.Id)
		if err != nil {
			panic(err)
		}
		scene := NewScene(&SceneData{
			Scene:             sceneData,
			DoorList:          doorList,
			MonsterConfigList: configList,
		})
		manager.scenes[sceneData.Id] = scene
	}
}

func (manager *SceneManager) GetScene(sceneId int) *Scene {
	return manager.scenes[sceneId]
}

func (manager *SceneManager) onPlayerDisconnect(s *session.Session) error {
	p, err := heroWithSession(s)
	if err != nil {
		return err
	}
	logger.Println("SceneManager.onPlayerDisconnect: 玩家网络断开", p.scene)
	// 移除session
	p.bindSession(nil)
	p.Destroy()
	return nil
}

func (manager *SceneManager) HeroEnterScene(s *session.Session, req *protocol.HeroEnterSceneRequest) error {
	if req.HeroData == nil {
		return errors.New("hero_data is nil")
	}
	scene := manager.scenes[req.SceneId]
	if scene == nil {
		return errors.New("scene not found")
	}
	hero := NewHero(s, req.HeroData)
	s.Bind(req.HeroData.Uid)
	hero.bindSession(s)
	scene.addHero(hero)
	return nil
}

func (manager *SceneManager) HeroSetViewRange(s *session.Session, req *protocol.HeroSetViewRangeRequest) error {
	p, err := heroWithSession(s)
	if err != nil {
		return err
	}
	logger.Println("hero设置屏幕范围:", req.HeroID, req.Width, req.Height)
	p.SetViewRange(req.Width/2, req.Height/2)
	return nil
}

func (manager *SceneManager) Attack(s *session.Session, req *protocol.AttackRequest) error {
	return nil
}

func (manager *SceneManager) HeroMove(s *session.Session, req *protocol.HeroMoveRequest) error {
	p, err := heroWithSession(s)
	if err != nil {
		return err
	}
	lastPoint := req.TracePaths[len(req.TracePaths)-1]
	targetX := lastPoint[1]
	targety := lastPoint[0]
	return p.MoveByPaths(int(targetX), int(targety), 0, req.TracePaths)
}

func (manager *SceneManager) HeroMoveStop(s *session.Session, req *protocol.HeroMoveStopRequest) error {
	p, err := heroWithSession(s)
	if err != nil {
		return err
	}
	return p.MoveStop(req.PosX, req.PosY, req.PosZ)
}

func (manager *SceneManager) SceneInfo(s *session.Session, req *protocol.SceneInfoRequest) error {
	items := make([]protocol.SceneInfoItem, 0)
	for _, scene := range manager.scenes {
		logger.Debugf("scenInfo: scene_id: %d,  当前人数: %d, 怪物数量:%d",
			scene.GetSceneId(), scene.totalPlayerCount(), scene.totalMonsterCount())
		items = append(items, protocol.SceneInfoItem{
			SceneId:    scene.GetSceneId(),
			MonsterCnt: scene.totalMonsterCount(),
			HeroCnt:    scene.totalPlayerCount(),
		})
	}
	return s.RPC("Manager.SceneInfoCallBack", &protocol.SceneInfoResponse{Scenes: items})
}

// 玩家文字消息
func (manager *SceneManager) TextMessage(s *session.Session, msg *protocol.TextMessageRequest) error {
	p, err := heroWithSession(s)
	if err != nil {
		return err
	}
	p.PushTask(func() {
		p.Broadcast(protocol.OnTextMessage, msg, false)
	})
	return nil
}

// 玩家语音消息
func (manager *SceneManager) VoiceMessage(s *session.Session, msg []byte) error {
	//p, err := heroWithSession(s)
	//if err != nil {
	//	return err
	//}

	//d := p.scene
	//if d != nil && d.group != nil {
	//	return d.group.Broadcast("onVoiceMessage", msg)
	//}

	return nil
}

// 玩家录制完语音
func (manager *SceneManager) RecordingVoice(s *session.Session, msg *protocol.RecordingVoice) error {
	//p, err := heroWithSession(s)
	//if err != nil {
	//	return err
	//}
	//
	//d := p.scene
	//resp := &protocol.PlayRecordingVoice{
	//	Uid:    s.UID(),
	//	FileId: msg.FileId,
	//}
	//
	//if d != nil && d.group != nil {
	//	return d.group.Broadcast("onRecordingVoice", resp)
	//}
	return nil
}

// 动态重置怪物
func (manager *SceneManager) DynamicResetMonsters(s *session.Session, req *protocol.DynamicResetMonstersRequest) error {
	sceneIds := make(map[int]int)
	for _, c := range req.Configs {
		sceneIds[c.SceneId] = 1
	}
	//先全部清除了
	for _, sid := range sceneIds {
		manager.scenes[sid].monsters.Range(func(key, value any) bool {
			manager.scenes[sid].removeMonster(value.(*Monster))
			time.Sleep(10 * time.Millisecond)
			return true
		})
	}
	time.Sleep(200 * time.Millisecond)
	for _, c := range req.Configs {
		err := manager.scenes[c.SceneId].initMonsterByConfig(c)
		if err != nil {
			return err
		}
		time.Sleep(300 * time.Millisecond)
	}
	return nil
}
