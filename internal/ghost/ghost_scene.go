package ghost

import (
	"bytes"
	"fmt"
	"github.com/ouyangzhongmin/gameserver/constants"
	"github.com/ouyangzhongmin/gameserver/db/model"
	"github.com/ouyangzhongmin/gameserver/internal/game"
	"github.com/ouyangzhongmin/gameserver/pkg/fileutil"
	"github.com/ouyangzhongmin/gameserver/pkg/logger"
	"github.com/ouyangzhongmin/gameserver/protocol"
	"github.com/ouyangzhongmin/nano/scheduler"
	"github.com/ouyangzhongmin/nano/session"
	"time"
)

type ghostScene struct {
	data            *model.Scene
	blockInfo       *game.BlockInfo
	entities        map[string]*ghostEntity
	toBuildViewList map[string]*ghostEntity
	aoiMgr          *ghostAoiMgr
	//每次更新的时间戳
	lastUpdateTimeStamp     int64
	refreshViewListDelatime int64
}

func newGhostScene(data *model.Scene) *ghostScene {
	s := &ghostScene{
		data:            data,
		entities:        make(map[string]*ghostEntity),
		toBuildViewList: make(map[string]*ghostEntity),
	}
	s.blockInfo = game.NewBlockInfo()
	buf, err := fileutil.ReadFile(fileutil.FindResourcePth(fmt.Sprintf("blocks/%s.block", data.MapFile)))
	if err != nil {
		panic(err)
	}
	err = s.blockInfo.ReadFrom(bytes.NewBuffer(buf))
	if err != nil {
		panic(err)
	}
	w := s.blockInfo.GetWidth()
	s.aoiMgr = newAoiMgr(int(w), int(w/constants.SCENE_AOI_GRID_SIZE))
	return s
}

func (s *ghostScene) initTimer() {
	//这个timer与handler在同一条线程,所以这里不需要处理并发问题
	// 每100MS计算一次
	scheduler.NewTimer(100*time.Millisecond, func() {
		if err := s.update(); err != nil {
			logger.Printf("scene:%d update error:%v\n", s.data.Id, err)
		}
	})
}

func (s *ghostScene) updateEntity(session *session.Session, e *protocol.GhostEntity) {
	le := s.getEntity(e.Uuid)
	if le == nil {
		//插入数据
		le = &ghostEntity{
			scene:       s,
			session:     session,
			GhostEntity: *e,
		}
		s.addEntity(le)
	} else {
		oldx, oldy, oldz := le.GetPos().X, le.GetPos().Y, le.GetPos().Z
		//更新数据
		le.GhostEntity = *e
		le.session = session
		if oldx != e.Pos.X || oldy != e.Pos.Y || oldz != e.Pos.Z {
			s.aoiMgr.Moved(le, e.Pos.X, e.Pos.Y, oldx, oldy)
		}
	}
	session.Bind(e.Id)
	session.Set("GhostHero", le)
}

func (s *ghostScene) addEntity(e *ghostEntity) {
	s.entities[e.Uuid] = e
	s.aoiMgr.Enter(e)
}

func (s *ghostScene) removeEntity(uuid string) {
	e := s.getEntity(uuid)
	if e == nil {
		return
	}
	if e.session != nil {
		e.session.Clear()
		e.session = nil
	}
	s.aoiMgr.Leave(e)
	delete(s.entities, uuid)
}

func (s *ghostScene) getEntity(uuid string) *ghostEntity {
	return s.entities[uuid]
}

func (s *ghostScene) addToBuildViewList(e *ghostEntity) {
	s.toBuildViewList[e.Uuid] = e
}

func (s *ghostScene) update() error {
	ts := time.Now().UnixMilli()
	//每帧的时间间隔
	elapsedTime := ts - s.lastUpdateTimeStamp

	s.refreshViewListDelatime += elapsedTime
	if s.refreshViewListDelatime >= 500 {
		//刷新所有对象的视野
		s.refreshViewList()
		s.refreshViewListDelatime = 0
	}

	s.lastUpdateTimeStamp = ts
	return nil
}

func (s *ghostScene) refreshViewList() {
	for key, v := range s.toBuildViewList {
		s._refreshEntityViewList(v)
		delete(s.toBuildViewList, key)
	}
}

func (s *ghostScene) _refreshEntityViewList(entity *ghostEntity) {
	if entity.EntityType == constants.ENTITY_TYPE_SPELL {
		return
	}
	s.updateEntityViewList(entity)

	entites := s.aoiMgr.Search(entity.GetPos().X, entity.GetPos().Y)
	for _, e0 := range entites {
		if e0 == nil {
			continue
		}
		e := e0.(*ghostEntity)
		if e != entity {
			if entity.GetEntityType() == constants.ENTITY_TYPE_HERO { //&&
				//如果我是英雄， 判定我能不能看见对方
				if entity.CanSee(e) && !entity.IsInViewList(e) {
					//原来不在视野内，现在看见了
					entity.onEnterView(e)      //进入他的视野
					e.onEnterOtherView(entity) //记录我进入了谁的视野
				}
			}
			if e.GetEntityType() == constants.ENTITY_TYPE_HERO {
				//循环的是英雄, 检查这个英雄是否能看见我
				if e.CanSee(entity) && !e.IsInViewList(entity) {
					//原来不在视野内，现在看见了
					e.onEnterView(entity)      //进入他的视野
					entity.onEnterOtherView(e) //记录我进入了谁的视野
				}
			}
		}
	}
}

// todo 这里的刷新视野频度会跟随同屏数量增加而成倍数增加，比如同屏一万人，那么每个人都需要遍历2万次去判断是否离开视野，这里需要重新评估是否有更好的方案
func (s *ghostScene) updateEntityViewList(entity *ghostEntity) {
	//检查当前视野内的对象是否已离开
	if entity.GetEntityType() == constants.ENTITY_TYPE_HERO {
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
	entity.canSeeMeViewList.Range(func(key, value interface{}) bool {
		target := value.(IMovableEntity)
		if target.GetScene() != entity.GetScene() || !target.CanSee(em) {
			//原来在视野内，现在看不见了
			target.onExitView(entity)      //自己离开了target的视野
			entity.onExitOtherView(target) //自己记录的m能看见我
		}
		return true
	})
}
