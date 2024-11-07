package game

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/lonng/nano/scheduler"
	"github.com/ouyangzhongmin/gameserver/pkg/shape"
	"sync/atomic"
)

type Entity struct {
	scene *Scene
	// nano的handler都在一条go scheduler.Sched()线程中执行的，
	// 所以这里定义每个对象都需要有自己的独立运行的携程
	_chTasks        chan scheduler.Task
	_chTasksBufSize int
	_chDestroy      chan struct{}
	_uuid           string // 不存储在数据库，只作为运行对象的唯一值
	_id             int64
	_name           string
	_entityType     int
	_pos            shape.Vector3
	_destroyed      atomic.Bool
}

func (e *Entity) initEntity(id int64, name string, entityType int, bufSize int) {
	e._chTasksBufSize = bufSize
	e._chTasks = make(chan scheduler.Task, bufSize)
	e._chDestroy = make(chan struct{})
	e._uuid = uuid.New().String()
	e._id = id
	e._name = name
	e._entityType = entityType
	go e._tasksFunc()
}

func (e *Entity) _tasksFunc() {
	for {
		select {
		case <-e._chDestroy:
			logger.Printf("stop entity:%d\n", e.GetID())
			e._destroyed.Store(true)
			return
		case task := <-e._chTasks:
			e._doTask(task)
		}
	}
}

func (e *Entity) _doTask(f func()) {
	defer func() {
		if err := recover(); err != nil {
			logger.Println(fmt.Sprintf("entity task err: %+v\n", err))
		}
	}()
	f()
}

func (e *Entity) PushTask(task scheduler.Task) {
	//todo 这里的task内可能会再有PushTask的调用，如果缓存区满后可能会导致死锁住，所以这里开携程了
	if e._destroyed.Load() {
		return
	}
	if len(e._chTasks) >= e._chTasksBufSize {
		logger.Errorf("Entity:%d-%s task buffer is full 开启携程", e._id, e._name)
		go func() {
			e._chTasks <- task
		}()
		return
	}
	e._chTasks <- task
}

func (e *Entity) Destroy() {
	e._id = 0
	e._name += "_destroyed"
	e.scene = nil
	close(e._chDestroy)
}

func (e *Entity) IsDestroyed() bool {
	return e._destroyed.Load()
}

func (e *Entity) onEnterScene(scene *Scene) {
	e.scene = scene
}

func (e *Entity) onExitScene(scene *Scene) {
	e.scene = nil
}

func (e *Entity) GetScene() *Scene {
	return e.scene
}

func (e *Entity) SetPos(x, y, z shape.Coord) {
	e._pos.X = x
	e._pos.Y = y
	e._pos.Z = z
}

func (e *Entity) GetPos() shape.Vector3 {
	return e._pos
}

func (e *Entity) GetID() int64 {
	return e._id
}

func (e *Entity) GetUUID() string {
	return e._uuid
}

func (e *Entity) GetEntityType() int {
	return e._entityType
}

func (e *Entity) ToString() string {
	return fmt.Sprintf("id:%d,uuid:%s, posX:%d, posY:%d, posZ:%d", e.GetID(), e.GetUUID(), e._pos.X, e._pos.Y, e._pos.Z)
}
