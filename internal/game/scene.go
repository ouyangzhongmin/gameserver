package game

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/lonng/nano/scheduler"
	"github.com/ouyangzhongmin/gameserver/db"
	"github.com/ouyangzhongmin/gameserver/db/model"
	"github.com/ouyangzhongmin/gameserver/internal/game/constants"
	"github.com/ouyangzhongmin/gameserver/pkg/async"
	"github.com/ouyangzhongmin/gameserver/pkg/fileutil"
	"github.com/ouyangzhongmin/gameserver/pkg/path"
	"github.com/ouyangzhongmin/gameserver/pkg/shape"
	"github.com/ouyangzhongmin/gameserver/protocol"
	log "github.com/sirupsen/logrus"
	"math"
	"math/rand"
	"sync"
	"time"
)

const (
	SCENE_CHAN_BUFFER_SIZE = 2048
)

type SceneData struct {
	model.Scene
	DoorList          []model.SceneDoor
	MonsterConfigList []model.SceneMonsterConfig
}

type Scene struct {
	logger    *log.Entry
	sceneId   int
	sceneData *SceneData
	blockInfo *BlockInfo
	//这里注意是用的heroId做key
	heros           sync.Map //需要线程安全
	monsters        sync.Map
	spells          sync.Map
	chTasks         chan scheduler.Task
	chStop          chan struct{}
	toBuildViewList sync.Map
	aoiMgr          *aoiMgr

	rebornMonsters sync.Map

	//基于大格子算法的AOI
	//entityBlocks [][]sync.Map

	updateTicker *time.Ticker
	//每次更新的时间戳
	lastUpdateTimeStamp     int64
	refreshViewListDelatime int64
}

func NewScene(sceneData *SceneData) *Scene {
	s := &Scene{
		sceneId:   sceneData.Scene.Id,
		sceneData: sceneData,
		logger:    log.WithField(fieldDesk, sceneData.Scene.Id),
		chTasks:   make(chan scheduler.Task, SCENE_CHAN_BUFFER_SIZE),
		chStop:    make(chan struct{}),
	}
	s.blockInfo = NewBlockInfo()
	buf, err := fileutil.ReadFile(fileutil.FindResourcePth(fmt.Sprintf("blocks/%s.block", s.sceneData.MapFile)))
	if err != nil {
		panic(err)
	}
	err = s.blockInfo.ReadFrom(bytes.NewBuffer(buf))
	if err != nil {
		panic(err)
	}
	w := s.blockInfo.GetWidth()

	s.aoiMgr = newAoiMgr(int(w), int(w/constants.SCENE_AOI_GRID_SIZE))

	s.updateTicker = time.NewTicker(100 * time.Millisecond)
	go s._tasksFunc()
	s.initTimer()
	s.initMonsters()

	s.lastUpdateTimeStamp = time.Now().UnixMilli()
	return s
}

func (s *Scene) initTimer() {
	//// 每100MS调用一次场景刷新
	//scheduler.NewTimer(100*time.Millisecond, func() {
	//	if err := s.update(); err != nil {
	//		logger.Printf("scene:%d update error:%v\n", s.sceneId, err)
	//	}
	//})

	// 每5S保存一次用户数据
	scheduler.NewTimer(5*time.Second, func() {
		if err := s.save(); err != nil {
			logger.Printf("scene:%d save error:%v\n", s.sceneId, err)
		}
	})

}

func (s *Scene) _tasksFunc() {
	defer s.updateTicker.Stop()
	for {
		select {
		case <-s.chStop:
			logger.Printf("stop scene:%d\n", s.sceneId)
			return
		case task := <-s.chTasks:
			s._doTask(task)
		case <-s.updateTicker.C:
			if err := s.update(); err != nil {
				logger.Printf("scene:%d update error:%v\n", s.sceneId, err)
			}
		}
	}
}

func (s *Scene) _doTask(f func()) {
	defer func() {
		if err := recover(); err != nil {
			logger.Println(fmt.Sprintf("scene task err: %+v\n", err))
		}
	}()
	f()
}

func (s *Scene) PushTask(task scheduler.Task) {
	if len(s.chTasks) >= SCENE_CHAN_BUFFER_SIZE {
		logger.Errorln("Scene chTasks缓冲区已满, 开启携程执行 chTasks <- task 操作")
		async.Run(func() {
			s.chTasks <- task
		})
	} else {
		s.chTasks <- task
	}
}

func (s *Scene) initMonsters() {
	if s.sceneData.MonsterConfigList != nil {
		for _, cfg := range s.sceneData.MonsterConfigList {
			err := s.initMonsterByConfig(cfg)
			if err != nil {
				panic(err)
			}
		}
	}
	logger.Println("初始怪物数量:", s.totalMonsterCount())
}

func (s *Scene) initMonsterByConfig(cfg model.SceneMonsterConfig) error {
	monsterData, err := db.QueryMonster(cfg.MonsterId)
	if err != nil {
		logger.Errorln("initMonsters err::" + err.Error())
		return err
	}
	aidata, err := db.QueryAiConfig(cfg.MonsterId)
	if err != nil {
		logger.Warningln("monster:%d 没有配置aiconfig", cfg.MonsterId)
	}
	rect := shape.Rect{
		X:      int64(cfg.Bornx - cfg.ARange),
		Y:      int64(cfg.Borny - cfg.ARange),
		Width:  int64(cfg.ARange * 2),
		Height: int64(cfg.ARange * 2),
	}
	if rect.X < 0 {
		rect.X = 0
	}
	if rect.Y < 0 {
		rect.Y = 0
	}
	fpath := fmt.Sprintf("blocks/%s_%d,%d,%d.paths", s.sceneData.MapFile, cfg.Bornx, cfg.Borny, cfg.ARange)
	buf, err := fileutil.ReadFile(fileutil.FindResourcePth(fpath))
	var spaths []*path.SerialPaths
	if err != nil {
		fmt.Printf("%s未配置:%v\n", fpath, err)
	} else {
		err = json.Unmarshal(buf, &spaths)
		if err != nil {
			logger.Warningln("Unmarshal paths file err:", err)
		}
	}

	for i := 0; i < cfg.Total; i++ {
		m := NewMonster(monsterData, i+1)
		m.SetSceneMonsterConfig(&cfg)
		if spaths != nil && len(spaths) > 0 {
			//预制路径
			sindex := rand.Intn(len(spaths))
			sp := spaths[sindex]
			m.SetPreparePaths(sp)
			m.SetPos(shape.Coord(sp.Paths[0].Sx), shape.Coord(sp.Paths[0].Sy), shape.Coord(cfg.Bornz))
		} else {
			//随机坐标
			rx, ry, err := s.GetRandomXY(rect, 100)
			if err != nil {
				logger.Warningln("monster init pos err:", err)
				rx, ry = shape.Coord(cfg.Bornx), shape.Coord(cfg.Borny)
			}
			m.SetPos(rx, ry, shape.Coord(cfg.Bornz))
		}
		m.bornPos.Copy(m.GetPos())
		m.SetMovableRect(rect)
		if aidata != nil {
			m.SetAiData(newMonsterAi(m, aidata))
		}
		logger.Debugf("newmonster:%d,%d,%d \n", m.GetID(), m.GetPos().X, m.GetPos().Y)
		s.addMonster(m)
	}
	return nil
}

// 复活一个monster
func (s *Scene) rebornOneMonster(rm *rebornMonster) {
	m := NewMonster(rm.Data, 0)
	m.SetSceneMonsterConfig(rm.Cfg)
	m.SetMovableRect(rm.MovableRect)
	if rm.PreparePaths != nil {
		m.SetPreparePaths(rm.PreparePaths)
		m.SetPos(shape.Coord(rm.PreparePaths.Paths[0].Sx), shape.Coord(rm.PreparePaths.Paths[0].Sy), shape.Coord(rm.Cfg.Bornz))
	} else {
		//随机坐标
		rect := rm.MovableRect
		rx, ry, err := s.GetRandomXY(rect, 100)
		if err != nil {
			logger.Warningln("monster init pos err:", err)
			rx, ry = shape.Coord(rm.Cfg.Bornx), shape.Coord(rm.Cfg.Borny)
		}
		m.SetPos(rx, ry, shape.Coord(rm.Cfg.Bornz))
	}
	if rm.aidata != nil {
		m.SetAiData(newMonsterAi(m, rm.aidata.(*model.Aiconfig)))
	}
	m.bornPos.Copy(m.GetPos())
	s.addMonster(m)

}

func (s *Scene) GetSceneId() int {
	return s.sceneId
}

func (s *Scene) GetSceneData() *SceneData {
	return s.sceneData
}

func (s *Scene) GetWidth() uint32 {
	return s.blockInfo.GetWidth()
}

func (s *Scene) GetHeight() uint32 {
	return s.blockInfo.GetHeight()
}

func (s *Scene) Stop() {
	s.chStop <- struct{}{}
}

func (s *Scene) totalPlayerCount() int {
	l := 0
	s.heros.Range(func(k, v interface{}) bool {
		l++
		return true
	})
	return l
}

func (s *Scene) totalMonsterCount() int {
	l := 0
	s.monsters.Range(func(k, v interface{}) bool {
		l++
		return true
	})
	return l
}

func (s *Scene) addHero(h *Hero) {
	//这个要在前面执行，并发的update内可能会取到空的scene
	if s.sceneData.Enterx > 0 && s.sceneData.Entery > 0 {
		//使用场景的出生点
		h.SetPos(shape.Coord(s.sceneData.Enterx), shape.Coord(s.sceneData.Entery), shape.Coord(s.sceneData.Enterz))
	}

	h.onEnterScene(s)
	s.heros.Store(h.GetID(), h)
	s.aoiMgr.Enter(h)

	h.SendMsg(protocol.OnEnterScene, &protocol.EnterSceneResponse{
		Scene:    s.sceneData.Scene,
		Doors:    s.sceneData.DoorList,
		HeroData: *h.GetData(),
	})
}

func (s *Scene) removeHero(h *Hero) {
	s.aoiMgr.Leave(h)
	s.heros.Delete(h.GetID())
	h.onExitScene(s)
}

func (s *Scene) addMonster(m *Monster) {
	//这个要在前面执行，并发的update内可能会取到空的scene
	m.onEnterScene(s)
	if tm, ok := s.monsters.Load(m.GetID()); ok {
		tm1 := tm.(*Monster)
		logger.Warningln("已经存在怪物:", tm1.GetID(), tm1._name)
	}
	s.monsters.Store(m.GetID(), m)

	s.aoiMgr.Enter(m)
}

func (s *Scene) removeMonster(m *Monster) {
	s.aoiMgr.Leave(m)

	s.monsters.Delete(m.GetID())
	m.onExitScene(s)
}

func (s *Scene) addRebornMonster(m *rebornMonster) {
	s.rebornMonsters.Store(m.Uid, m)
}

func (s *Scene) addSpell(m *SpellEntity) {
	//这个要在前面执行，并发的update内可能会取到空的scene
	m.onEnterScene(s)
	s.spells.Store(m.GetUUID(), m)

	s.aoiMgr.Enter(m)
}

func (s *Scene) removeSpell(m *SpellEntity) {
	s.aoiMgr.Leave(m)

	s.spells.Delete(m.GetUUID())
	m.onExitScene(s)
}

func (s *Scene) entityMoved(e IMovableEntity, oldX, oldY shape.Coord) {
	s.PushTask(func() {
		s.aoiMgr.Moved(e, oldX, oldY)
		s.addToBuildViewList(e)
	})
}

func (s *Scene) update() error {
	ts := time.Now().UnixMilli()
	//每帧的时间间隔
	elapsedTime := ts - s.lastUpdateTimeStamp
	//heroCnt := 0
	s.heros.Range(func(key, value any) bool {
		h := value.(*Hero)
		//heroCnt += 1
		if h.session == nil {
			logger.Errorln("hero.session is nil", h.GetID(), h._name)
			s.removeHero(h)
		} else {
			h.PushTask(func() {
				s.updateEntityViewList(h)
				err := h.update(ts, elapsedTime)
				if err != nil {
					logger.Errorln("hero.update err:", err)
				}
			})
		}

		return true
	})
	s.monsters.Range(func(key, value any) bool {
		m := value.(*Monster)
		m.PushTask(func() {
			s.updateEntityViewList(m)
			err := m.update(ts, elapsedTime)
			if err != nil {
				logger.Errorln("monster.update err:", err)
			}
		})
		return true
	})
	s.spells.Range(func(key, value any) bool {
		e := value.(*SpellEntity)
		e.PushTask(func() {
			s.updateEntityViewList(e)
			err := e.update(ts, elapsedTime)
			if err != nil {
				logger.Errorln("spellEntity.update err:", err)
			}
		})
		return true
	})
	s.refreshViewListDelatime += elapsedTime
	if s.refreshViewListDelatime >= 500 {
		//刷新所有对象的视野
		s.refreshViewList()
		s.refreshViewListDelatime = 0

		// 复活monster
		s.rebornMonsters.Range(func(key, value any) bool {
			m := value.(*rebornMonster)
			if m.RebornTimestamp <= ts {
				s.rebornOneMonster(m)
				s.rebornMonsters.Delete(key)
			}
			return true
		})
	}

	s.lastUpdateTimeStamp = ts

	return nil
}

func (s *Scene) updateEntityViewList(entity IMovableEntity) {
	var em *movableEntity = nil
	switch val := entity.(type) {
	case *Hero:
		em = &val.movableEntity
	case *Monster:
		em = &val.movableEntity
	case *SpellEntity:
		em = &val.movableEntity
	}
	//检查当前视野内的对象是否已离开
	if em.GetEntityType() == constants.ENTITY_TYPE_HERO {
		em.viewList.Range(func(key, value interface{}) bool {
			target := value.(IMovableEntity)
			if target.GetScene() != entity.GetScene() || !entity.CanSee(target) {
				//原来在视野内，现在看不见了
				entity.onExitView(target)      //target离开了m的视野
				target.onExitOtherView(entity) //清除target记录的m能看见他
			}
			return true
		})
	}
	em.canSeeMeViewList.Range(func(key, value interface{}) bool {
		target := value.(IMovableEntity)
		if target.GetScene() != entity.GetScene() || !target.CanSee(em) {
			//原来在视野内，现在看不见了
			target.onExitView(em)      //自己离开了target的视野
			em.onExitOtherView(target) //自己记录的m能看见我
		}
		return true
	})
}

func (s *Scene) save() error {
	s.heros.Range(func(key, value any) bool {
		h := value.(*Hero)
		h.save()
		return true
	})
	return nil
}

func (s *Scene) addToBuildViewList(e IMovableEntity) {
	s.PushTask(func() {
		s.toBuildViewList.Store(e.GetUUID(), e)
	})
}

func (s *Scene) refreshViewList() {
	//改为再同一条线程操作block数组，去掉加锁
	//s.blockmutx.RLock()
	//defer s.blockmutx.RUnlock()
	s.PushTask(func() {
		s.toBuildViewList.Range(func(key, value any) bool {
			s._refreshEntityViewList(value.(IMovableEntity))
			s.toBuildViewList.Delete(key)
			return true
		})
	})
}

func (s *Scene) _refreshEntityViewList(entity IMovableEntity) {
	entites := s.aoiMgr.Search(entity.GetPos().X, entity.GetPos().Y)
	for _, e0 := range entites {
		e := e0.(IMovableEntity)
		if e != entity && entity.GetEntityType() == constants.ENTITY_TYPE_HERO {
			//如果我是英雄， 判定我能不能看见对方
			if entity.CanSee(e) && !entity.IsInViewList(e) {
				//原来不在视野内，现在看见了
				entity.onEnterView(e)      //进入他的视野
				e.onEnterOtherView(entity) //记录我进入了谁的视野
			}
		}
		if e != entity && e.GetEntityType() == constants.ENTITY_TYPE_HERO {
			//循环的是英雄, 检查这个英雄是否能看见我
			if e.CanSee(entity) && !e.IsInViewList(entity) {
				//原来不在视野内，现在看见了
				e.onEnterView(entity)      //进入他的视野
				entity.onEnterOtherView(e) //记录我进入了谁的视野
			}
		}
	}
}

// 这个是线程安全的，可并发调用, 注意区别PathFinder
func (s *Scene) FindPath(sx, sy, ex, ey shape.Coord) ([][]int32, error) {
	path, _, _, err := s.blockInfo.FindPath(int32(sx), int32(sy), int32(ex), int32(ey))
	return path, err
}

func (s *Scene) IsWalkable(x, y shape.Coord) bool {
	if s.blockInfo != nil {
		return s.blockInfo.IsWalkable(int32(x), int32(y))
	}
	return false
}

// 通过圆范围查找对象
func (s *Scene) getEntitiesByRange(cx, cy, arange shape.Coord) map[string]IMovableEntity {
	result := make(map[string]IMovableEntity)
	entites := s.aoiMgr.Search(cx, cy)
	for _, e0 := range entites {
		e := e0.(IMovableEntity)
		//if shape.IsInsideCircle(float64(cx), float64(cy), float64(arange), float64(e.GetPos().X), float64(e.GetPos().Y)) {
		if shape.Coord(math.Abs(float64(cx-e.GetPos().X))) <= arange && shape.Coord(math.Abs(float64(cy-e.GetPos().Y))) <= arange {
			//判定是否在警戒范围内
			result[e.GetUUID()] = e
		}
	}
	return result
}

func (s *Scene) GetRandomXY(rect shape.Rect, cnt int) (shape.Coord, shape.Coord, error) {
	return s.blockInfo.GetRandomXY(rect, cnt)
}
