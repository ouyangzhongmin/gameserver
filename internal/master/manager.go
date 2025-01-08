package master

import (
	"errors"
	"fmt"
	"github.com/ouyangzhongmin/gameserver/constants"
	"github.com/ouyangzhongmin/gameserver/db"
	"github.com/ouyangzhongmin/gameserver/db/model"
	"github.com/ouyangzhongmin/gameserver/pkg/async"
	"github.com/ouyangzhongmin/gameserver/protocol"
	"github.com/ouyangzhongmin/nano/scheduler"
	"sync"
	"time"

	"github.com/ouyangzhongmin/nano"
	"github.com/ouyangzhongmin/nano/component"
	"github.com/ouyangzhongmin/nano/session"
	log "github.com/sirupsen/logrus"
)

const kickResetBacklog = 8

type (
	Manager struct {
		component.Base
		group *nano.Group // 广播channel

		//这个timer与handler在同一条线程,所以这里players不需要处理并发问题
		players    map[int64]*User   // 所有的玩家
		chKick     chan int64        // 退出队列
		chReset    chan int64        // 重置队列
		chRecharge chan RechargeInfo // 充值信息
		chScene    chan int

		scenesCount sync.Map
	}

	User struct {
		session  *session.Session
		data     *model.User
		Uid      int64
		heroData *model.Hero
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
		chKick:     make(chan int64, kickResetBacklog),
		chReset:    make(chan int64, kickResetBacklog),
		chRecharge: make(chan RechargeInfo, 32),
		chScene:    make(chan int, 32),
	}
}

func (m *Manager) AfterInit() {
	session.Lifetime.OnClosed(func(s *session.Session) {
		m.group.Leave(s)
		m.removePlayer(s.UID())
	})

	//这个timer与handler在同一条线程,所以这里players不需要处理并发问题
	scheduler.NewTimer(time.Second, func() {
	ctrl:
		for {
			select {
			case uid := <-m.chKick:
				p, ok := userManager.player(uid)
				if !ok || p.session == nil {
					logger.Errorf("玩家%d不在线", uid)
				}
				p.session.Close()
				logger.Infof("踢出玩家, UID=%d", uid)
			case <-m.chScene:
				m.reqSceneInfo()
			case uid := <-m.chReset:
				p, ok := userManager.player(uid)
				if !ok {
					return
				}
				if p.session != nil {
					logger.Errorf("玩家正在游戏中，不能重置: %d", uid)
					return
				}
				userManager.removePlayer(uid)
				logger.Infof("重置玩家, UID=%d", uid)

			default:
				break ctrl
			}
		}
	})
	// 每60S更新一次场景统计
	scheduler.NewTimer(60*time.Second, func() {
		m.chScene <- 1 //先去请求数据，等数据返回后再保存
		time.Sleep(time.Millisecond * 500)
		m.dumpSceneInfo()
	})
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
	// todo 切换场景时需要记录这个值
	s.Set("sceneId", sceneId)
	cell := cellManager.getSceneCell(user.heroData)
	if cell == nil {
		return errors.New("未匹配到任何cell")
	}
	s.Set("cellId", cell.CellID)
	s.Set("remoteAddr", cell.RemoteAddr)
	s.Router().Bind("SceneManager", cell.RemoteAddr)
	err = s.RPC("GateService.RecordScene", &protocol.UserSceneId{
		Uid:        uid,
		SceneId:    sceneId,
		CellId:     cell.CellID,
		RemoteAddr: cell.RemoteAddr,
	})
	if err != nil {
		logger.Errorf("rpc.Call(GateService.RecordScene) err: %v \n", err)
	}

	err = s.RPC("SceneManager.HeroEnterScene", &protocol.HeroEnterSceneRequest{
		SceneId:  sceneId,
		HeroData: heroData,
	})
	if err != nil {
		logger.Errorf("rpc.Call(SceneManager.HeroEnterScene) err: %v \n", err)
	}
	return err
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

	sceneId := constants.DEFAULT_SCENE
	heroData := createRandomHero(uid, sceneId, req.Name, req.Avatar, req.AttrType)
	id, err := db.InsertHero(heroData)
	if err != nil {
		return err
	}
	heroData.Id = id

	if user.session != nil && user.session != s {
		// 移除广播频道
		m.group.Leave(user.session)
		user.session.Clear()
		user.session.Close()
	}
	// 绑定新session
	user.session = s
	user.heroData = heroData
	// 添加到广播频道
	m.group.Add(s)

	//进入场景
	sceneId = heroData.SceneId
	if sceneId == 0 {
		sceneId = constants.DEFAULT_SCENE
	}
	res := &protocol.ChooseHeroResponse{
		Hero: *heroData,
	}
	s.Response(res)
	// todo 切换场景时需要记录这个值
	s.Set("sceneId", sceneId)
	cell := cellManager.getSceneCell(user.heroData)
	if cell == nil {
		return errors.New("未匹配到任何cell")
	}
	s.Set("cellId", cell.CellID)
	s.Set("remoteAddr", cell.RemoteAddr)
	s.Router().Bind("SceneManager", cell.RemoteAddr)
	err = s.RPC("GateService.RecordScene", &protocol.UserSceneId{
		Uid:        uid,
		SceneId:    sceneId,
		CellId:     cell.CellID,
		RemoteAddr: cell.RemoteAddr,
	})
	if err != nil {
		logger.Errorf("rpc.Call(GateService.RecordScene) err: %v \n", err)
	}

	err = s.RPC("SceneManager.HeroEnterScene", &protocol.HeroEnterSceneRequest{
		SceneId:  sceneId,
		HeroData: heroData,
	})
	if err != nil {
		logger.Errorf("rpc.Call(SceneManager.HeroEnterScene) err: %v \n", err)
	}
	return err
}

func (m *Manager) HeroChangeScene(s *session.Session, req *protocol.HeroChangeSceneRequest) error {
	uid := req.Uid
	log.Infof("玩家: %d切换场景: %+v", req.Uid, req)
	sceneId := req.SceneId
	user, ok := m.player(uid)
	if !ok {
		str := fmt.Sprintf("玩家: %d不在线", uid)
		return errors.New(str)
	}
	if user.session != nil && user.session != s {
		// 移除广播频道
		m.group.Leave(user.session)
		user.session.Clear()
		user.session.Close()
	}
	user.session = s

	oldSceneId := user.heroData.SceneId
	if sceneId == oldSceneId {
		return errors.New("已在当前场景")
	}
	// 离开上一个场景
	err := s.RPC("SceneManager.HeroLeaveScene", &protocol.HeroLeaveSceneRequest{
		SceneId: oldSceneId,
		HeroId:  user.heroData.Id,
	})
	if err != nil {
		return err
	}
	//进入新场景
	user.heroData.SceneId = sceneId
	s.Router().Delete("SceneManager")
	// todo 切换场景时需要记录这个值
	s.Set("sceneId", sceneId)
	cell := cellManager.getSceneCell(user.heroData)
	if cell == nil {
		return errors.New("未匹配到任何cell")
	}
	s.Set("cellId", cell.CellID)
	s.Set("remoteAddr", cell.RemoteAddr)
	s.Router().Bind("SceneManager", cell.RemoteAddr)
	err = s.RPC("GateService.RecordScene", &protocol.UserSceneId{
		Uid:        uid,
		SceneId:    sceneId,
		CellId:     cell.CellID,
		RemoteAddr: cell.RemoteAddr,
	})
	if err != nil {
		logger.Errorf("rpc.Call(GateService.RecordScene) err: %v \n", err)
	}

	err = s.RPC("SceneManager.HeroEnterScene", &protocol.HeroEnterSceneRequest{
		SceneId:  sceneId,
		HeroData: user.heroData,
	})
	if err != nil {
		logger.Errorf("rpc.Call(SceneManager.HeroEnterScene) err: %v \n", err)
	}
	err = db.UpdateHero(&model.Hero{Id: user.heroData.Id, SceneId: user.heroData.SceneId})
	return err
}

func (m *Manager) player(uid int64) (*User, bool) {
	p, ok := m.players[uid]
	return p, ok
}

func (m *Manager) addPlayer(user *User) {
	uid := user.Uid
	_, ok := userManager.player(uid)
	if ok {
		logger.Errorf("玩家%d已在线，正在覆盖", uid)
	}
	m.players[uid] = user
	logger.Infof("玩家上线, UID=%d", uid)
}

func (m *Manager) removePlayer(uid int64) {
	_, ok := userManager.player(uid)
	if !ok {
		return
	}
	delete(m.players, uid)
	log.Infof("玩家: %d从在线列表中删除, 剩余：%d", uid, len(m.players))
	logger.Infof("删除玩家, UID=%d", uid)
}

func (m *Manager) sessionCount() int {
	return len(m.players)
}

func (m *Manager) SceneInfoCallBack(s *session.Session, req *protocol.SceneInfoResponse) error {
	for _, scene := range req.Scenes {
		m.scenesCount.Store(fmt.Sprintf("%d-%d", scene.SceneId, scene.CellId), scene)
	}
	return nil
}

func (m *Manager) reqSceneInfo() {
	for _, scene := range cellManager.scenesCells {
		for _, scell := range scene.Cells {
			err := nano.RPCWithAddr("SceneManager.SceneInfo", &protocol.SceneInfoRequest{}, scell.RemoteAddr)
			if err != nil {
				logger.Errorln(err)
				return
			}
		}
	}
}

func (m *Manager) dumpSceneInfo() {
	logger.Infof("在线人数: %d  当前时间: %s", m.sessionCount(), time.Now().Format("2006-01-02 15:04:05"))
	sCount := userManager.sessionCount()

	async.Run(func() {
		// 统计结果异步写入数据库
		scenesCount := make(map[string]interface{})
		m.scenesCount.Range(func(k, v interface{}) bool {
			scenesCount[k.(string)] = v
			return true
		})
		db.InsertOnline(sCount, scenesCount, time.Now().UnixMilli())
	})
}
