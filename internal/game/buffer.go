package game

import (
	"github.com/ouyangzhongmin/gameserver/db/model"
	"github.com/ouyangzhongmin/gameserver/internal/game/constants"
	"github.com/ouyangzhongmin/gameserver/internal/game/object"
	"github.com/ouyangzhongmin/gameserver/protocol"
)

type Buffer struct {
	*object.BufferObject
	target IEntity
}

func NewBuffer(target IEntity, state *model.BufferState) *Buffer {
	buf := &Buffer{
		BufferObject: object.NewBufferObject(state),
		target:       target,
	}
	buf.initTotalTime()
	buf.broadcastAdd()
	return buf
}

func (buf *Buffer) Add(state *model.BufferState) {
	//已经存在相同的
	if state.Stackable != 0 {
		//可叠加
		buf.EffectCnt += state.EffectCnt
		buf.initTotalTime()
	} else {
		//重置为新的
		buf.CurCnt = 0
		buf.ElapsedTime = 0
	}
	buf.broadcastAdd()
}

func (buf *Buffer) initTotalTime() {
	buf.TotalTime = int64(buf.EffectCnt*buf.EffectDurationTime + (buf.EffectCnt-1)*buf.EffectDisappearTime)
}

func (buf *Buffer) update(curMilliSecond int64, elapsedTime int64) {
	if buf.target == nil {
		return
	}
	buf.ElapsedTime += elapsedTime
	cnt := int(buf.ElapsedTime/int64(buf.EffectDurationTime+buf.EffectDisappearTime)) + 1
	if buf.CurCnt != cnt {
		//新的一次效果执行
		buf.CurCnt = cnt
		buf.doOnceHurt()
	}
	if buf.ElapsedTime > buf.TotalTime {
		//到时间了，清除这个buf
		buf.Remove()
		return
	}
}

func (buf *Buffer) doOnceHurt() {
	switch val := buf.target.(type) {
	case *Hero:
		val.onBeenHurt(int64(buf.Damage))
	case *Monster:
		val.onBeenHurt(int64(buf.Damage))
	}
}

func (buf *Buffer) Remove() {
	target := buf.target.(*movableEntity)
	target.removeBuffer(buf.Id)

	switch val := buf.target.(type) {
	case *Hero:
		val.Broadcast(protocol.OnBufferRemove, &protocol.EntitBufferRemoveResponse{
			ID:         val.GetID(),
			EntityType: constants.ENTITY_TYPE_HERO,
			BufID:      buf.Id,
		})
	case *Monster:
		val.Broadcast(protocol.OnBufferRemove, &protocol.EntitBufferRemoveResponse{
			ID:         val.GetID(),
			EntityType: constants.ENTITY_TYPE_MONSTER,
			BufID:      buf.Id,
		})
	}

	buf.target = nil
}

func (buf *Buffer) broadcastAdd() {
	switch val := buf.target.(type) {
	case *Hero:
		val.Broadcast(protocol.OnBufferAdd, &protocol.EntityBufferAddResponse{
			ID:         val.GetID(),
			EntityType: constants.ENTITY_TYPE_HERO,
			Buf:        buf.BufferObject,
		})
	case *Monster:
		val.Broadcast(protocol.OnBufferAdd, &protocol.EntityBufferAddResponse{
			ID:         val.GetID(),
			EntityType: constants.ENTITY_TYPE_MONSTER,
			Buf:        buf.BufferObject,
		})
	}
}
