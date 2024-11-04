package master

import (
	"errors"
	"github.com/lonng/nano/scheduler"
	"github.com/ouyangzhongmin/gameserver/db"
	"github.com/ouyangzhongmin/gameserver/internal/game/constants"
	"github.com/ouyangzhongmin/gameserver/pkg/async"
	"github.com/ouyangzhongmin/gameserver/protocol"
	"sync"
	"time"

	"github.com/lonng/nano"
	"github.com/lonng/nano/component"
	"github.com/lonng/nano/session"
	log "github.com/sirupsen/logrus"
)

const kickResetBacklog = 8

var defaultManager = NewManager()

type (
	Manager struct {
		component.Base
		group *nano.Group // 广播channel

		//这里注意是用的userId做key ,scene里使用的是heroId做key
		players    map[int64]*User   // 所有的玩家
		chAdd      chan *User        // 添加队列
		chRemove   chan int64        // 删除队列
		chKick     chan int64        // 退出队列
		chReset    chan int64        // 重置队列
		chRecharge chan RechargeInfo // 充值信息
		chScene    chan int

		scenesCount sync.Map
	}

	RechargeInfo struct {
		Uid  int64 // 用户ID
		Coin int64 // 房卡数量
	}
)

func NewManager() *Manager {
	return &Manager{
		group:      nano.NewGroup("_SYSTEM_MESSAGE_BROADCAST"),
		players:    map[int64]*User{},
		chAdd:      make(chan *User, 1024),
		chRemove:   make(chan int64, 1024),
		chKick:     make(chan int64, kickResetBacklog),
		chReset:    make(chan int64, kickResetBacklog),
		chRecharge: make(chan RechargeInfo, 32),
		chScene:    make(chan int, 32),
	}
}

func (m *Manager) AfterInit() {
	session.Lifetime.OnClosed(func(s *session.Session) {
		m.group.Leave(s)
		if s.UID() > 0 {
			m.removePlayer(s.UID())
		}
	})
	go m.doChanFunc()

	// 每60S更新一次场景统计
	scheduler.NewTimer(60*time.Second, func() {
		m.chScene <- 1 //先去请求数据，等数据返回后再保存
		time.Sleep(time.Millisecond * 500)
		m.dumpSceneInfo()
	})
}

func (m *Manager) doChanFunc() {
	for {
		select {
		case user := <-m.chAdd:
			uid := user.Uid
			_, ok := defaultManager.player(uid)
			if ok {
				logger.Errorf("玩家%d已在线，正在覆盖", uid)
			}
			m.players[uid] = user
			logger.Infof("玩家上线, UID=%d", uid)

		case uid := <-m.chRemove:
			_, ok := defaultManager.player(uid)
			if !ok {
				return
			}
			delete(m.players, uid)
			log.Infof("玩家: %d从在线列表中删除, 剩余：%d", uid, len(m.players))
			logger.Infof("删除玩家, UID=%d", uid)
		case uid := <-m.chKick:
			p, ok := defaultManager.player(uid)
			if !ok || p.session == nil {
				logger.Errorf("玩家%d不在线", uid)
			}
			p.session.Close()
			logger.Infof("踢出玩家, UID=%d", uid)
		case <-m.chScene:
			m.reqSceneInfo()
		case uid := <-m.chReset:
			p, ok := defaultManager.player(uid)
			if !ok {
				return
			}
			if p.session != nil {
				logger.Errorf("玩家正在游戏中，不能重置: %d", uid)
				return
			}
			defaultManager.removePlayer(uid)
			logger.Infof("重置玩家, UID=%d", uid)
			//case ri := <-m.chRecharge:
			//	player, ok := m.player(ri.Uid)
			//	// 如果玩家在线
			//	if s := player.session; ok && s != nil {
			//		s.Push("onCoinChange", &protocol.CoinChangeInformation{Coin: ri.Coin})
			//	}
		}
	}
}

// nano的handler都在一条go scheduler.Sched()线程中执行的，所有handler如果耗时高需要开异步执行
func (m *Manager) ChooseHero(s *session.Session, req *protocol.ChooseHeroRequest) error {
	uid := req.Uid
	heroId := req.HeroId
	s.Bind(uid)
	log.Infof("玩家: %d选择英雄进入游戏: %+v", req.Uid, req)
	user, ok := m.player(uid)
	if !ok {
		log.Infof("玩家: %d不在线", uid)
		userData, err := db.QueryUser(uid)
		if err != nil {
			return err
		}
		user = &User{
			Uid:     uid,
			session: s,
			data:    userData,
		}
		m.addPlayer(user)
	} else {
		log.Infof("玩家: %d已经在线", uid)
		// 重置之前的session
		if user.session != nil && user.session != s {
			// 移除广播频道
			m.group.Leave(user.session)
			user.session.Clear()
			user.session.Close()
		}
	}
	heroData, err := db.QueryHero(heroId)
	if err != nil {
		return errors.New("英雄不存在")
	}
	// 绑定新session
	user.session = s
	user.heroData = heroData
	// 添加到广播频道
	m.group.Add(s)

	//进入场景
	sceneId := heroData.SceneId
	if sceneId == 0 {
		sceneId = constants.DEFAULT_SCENE
	}

	return s.RPC("SceneManager.HeroEnterScene", &protocol.HeroEnterSceneRequest{
		SceneId:  sceneId,
		HeroData: heroData,
	})
}

func (m *Manager) CreateHero(s *session.Session, req *protocol.CreateHeroRequest) error {
	uid := req.Uid
	s.Bind(uid)
	log.Infof("玩家: %d创建英雄进入游戏: %+v", req.Uid, req)
	user, ok := m.player(uid)
	if !ok {
		log.Infof("玩家: %d不在线", uid)
		userData, err := db.QueryUser(uid)
		if err != nil {
			return err
		}
		user = &User{
			Uid:     uid,
			session: s,
			data:    userData,
		}
		m.addPlayer(user)
	} else {
		log.Infof("玩家: %d已经在线", uid)
		// 重置之前的session
		if user.session != nil && user.session != s {
			// 移除广播频道
			m.group.Leave(user.session)
			user.session.Clear()
			user.session.Close()
		}
	}
	heroData := createRandomHero(uid, constants.DEFAULT_SCENE, req.Name, req.Avatar, req.AttrType)
	id, err := db.InsertHero(heroData)
	if err != nil {
		return err
	}
	heroData.Id = id

	// 绑定新session
	user.session = s
	user.heroData = heroData
	// 添加到广播频道
	m.group.Add(s)

	//进入场景
	sceneId := heroData.SceneId
	if sceneId == 0 {
		sceneId = constants.DEFAULT_SCENE
	}
	res := &protocol.ChooseHeroResponse{
		Hero: *heroData,
	}
	s.Response(res)
	err = s.RPC("SceneManager.HeroEnterScene", &protocol.HeroEnterSceneRequest{
		SceneId:  sceneId,
		HeroData: heroData,
	})
	return err
}

func (m *Manager) player(uid int64) (*User, bool) {
	p, ok := m.players[uid]

	return p, ok
}

func (m *Manager) addPlayer(h *User) {
	m.chAdd <- h
}

func (m *Manager) removePlayer(uid int64) {
	m.chRemove <- uid
}

func (m *Manager) sessionCount() int {
	return len(m.players)
}

func (m *Manager) SceneInfoCallBack(s *session.Session, req *protocol.SceneInfoResponse) error {
	for _, scene := range req.Scenes {
		m.scenesCount.Store(scene.SceneId, scene)
	}
	return nil
}

func (m *Manager) reqSceneInfo() {
	for _, s := range m.players {
		err := s.session.RPC("SceneManager.SceneInfo", &protocol.SceneInfoRequest{})
		if err != nil {
			logger.Errorln(err)
			return
		}
		break
	}
}

func (m *Manager) dumpSceneInfo() {
	logger.Infof("在线人数: %d  当前时间: %s", m.sessionCount(), time.Now().Format("2006-01-02 15:04:05"))
	sCount := defaultManager.sessionCount()

	async.Run(func() {
		// 统计结果异步写入数据库
		scenesCount := make(map[int]interface{})
		m.scenesCount.Range(func(k, v interface{}) bool {
			scenesCount[k.(int)] = v
			return true
		})
		db.InsertOnline(sCount, scenesCount, time.Now().UnixMilli())
	})
}
