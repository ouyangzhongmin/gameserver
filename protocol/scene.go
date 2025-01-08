package protocol

import (
	"github.com/ouyangzhongmin/gameserver/db/model"
	"github.com/ouyangzhongmin/gameserver/internal/game/object"
	"github.com/ouyangzhongmin/gameserver/pkg/coord"
)

type UserSceneId struct {
	Uid     int64 `json:"uid"`
	SceneId int   `json:"scene_id"`
}

type HeroEnterSceneRequest struct {
	SceneId  int `json:"scene_id"`
	HeroData *model.Hero
}

type HeroLeaveSceneRequest struct {
	SceneId int   `json:"scene_id"`
	HeroId  int64 `json:"hero_id"`
}

type SceneInfoRequest struct {
}

type SceneInfoResponse struct {
	Scenes []SceneInfoItem `json:"scenes"`
}

type SceneInfoItem struct {
	SceneId    int `json:"scene_id"`
	MonsterCnt int `json:"monster_cnt"`
	HeroCnt    int `json:"hero_cnt"`
}

type EnterSceneResponse struct {
	Scene    model.Scene       `json:"scene"`
	Doors    []model.SceneDoor `json:"doors"`
	HeroData object.HeroObject `json:"hero_data"`
}

type HeroSetViewRangeRequest struct {
	HeroID int64 `json:"hero_id"`
	Width  int   `json:"width"`
	Height int   `json:"height"`
}

type TargetEnterViewResponse struct {
	EntityType int                    `json:"entity_type"`
	Data       interface{}            `json:"data"`
	Buffers    []*object.BufferObject `json:"buffers"`
}

type TargetExitViewResponse struct {
	EntityType int   `json:"entity_type"`
	ID         int64 `json:"id"`
}

type HeroMoveRequest struct {
	Uid        int64     `json:"uid"`
	TracePaths [][]int32 `json:"trace_paths"` //前端需要定时同步一小段路
}

type HeroMoveStopRequest struct {
	Uid  int64       `json:"uid"`
	PosX coord.Coord `json:"pos_x"`
	PosY coord.Coord `json:"pos_y"`
	PosZ coord.Coord `json:"pos_z"`
}

type HeroMoveTraceResponse struct {
	ID         int64       `json:"id"`
	TracePaths [][]int32   `json:"trace_paths"` //trace_paths[0][0] 前面的是Y轴的数据，后面的是X轴的数据，trace_paths[y][x],前端要注意
	StepTime int         `json:"step_time"`
	PosX     coord.Coord `json:"pos_x"`
	PosY     coord.Coord `json:"pos_y"`
	PosZ     coord.Coord `json:"pos_z"`
}

type HeroMoveStopResponse struct {
	ID   int64       `json:"id"`
	PosX coord.Coord `json:"pos_x"`
	PosY coord.Coord `json:"pos_y"`
	PosZ coord.Coord `json:"pos_z"`
}

type MonsterMoveTraceResponse struct {
	ID         int64       `json:"id"`
	TracePaths [][]int32   `json:"trace_paths"`
	StepTime int         `json:"step_time"`
	PosX     coord.Coord `json:"pos_x"`
	PosY     coord.Coord `json:"pos_y"`
	PosZ     coord.Coord `json:"pos_z"`
}

type MonsterMoveStopResponse struct {
	ID   int64       `json:"id"`
	PosX coord.Coord `json:"pos_x"`
	PosY coord.Coord `json:"pos_y"`
	PosZ coord.Coord `json:"pos_z"`
}

type AttackRequest struct {
	AttackerId int64  `json:"attacker_id"`
	TargetId   int64  `json:"target_id"`
	TargetType int    `json:"target_type"` // 0 monster , 1 hero
	Action     string `json:"action"`      // 0 monster , 1 hero
}

type ReleaseSpellResponse struct {
	SpellObject *object.SpellObject `json:"spell_object"`
}

type LifeChangedResponse struct {
	ID         int64 `json:"id"`
	EntityType int   `json:"entity_type"`
	Damage     int64 `json:"damage"`
	Life       int64 `json:"life"`
	MaxLife    int64 `json:"max_life"`
}

type ManaChangedResponse struct {
	ID         int64 `json:"id"`
	EntityType int   `json:"entity_type"`
	Cost       int64 `json:"cost"`
	Mana       int64 `json:"mana"`
	MaxMana    int64 `json:"max_mana"`
}

type EntityDieResponse struct {
	ID         int64 `json:"id"`
	EntityType int   `json:"entity_type"`
}

// 前端需要自行判定id相同的覆盖旧的
type EntityBufferAddResponse struct {
	ID         int64 `json:"id"`
	EntityType int   `json:"entity_type"`
	Buf        *object.BufferObject
}

type EntitBufferRemoveResponse struct {
	ID         int64 `json:"id"`
	EntityType int   `json:"entity_type"`
	BufID      int   `json:"buf_id"`
}

type MonsterAttackResponse struct {
	ID         int64       `json:"id"`
	Action     string      `json:"action"`
	Damage     int64       `json:"damage"`
	TargetId   int64       `json:"target_id"`
	EntityType int         `json:"entity_type"`
	PosX       coord.Coord `json:"pos_x"`
	PosY       coord.Coord `json:"pos_y"`
	PosZ       coord.Coord `json:"pos_z"`
}

type TextMessageRequest struct {
	HeroId int64  `json:"hero_id"`
	Msg    string `json:"msg"`
}

type TextMessageResponse struct {
	HeroId int64  `json:"hero_id"`
	Msg    string `json:"msg"`
}

type RecordingVoice struct {
	FileId string `json:"fileId"`
}

type PlayRecordingVoice struct {
	Uid    int64  `json:"uid"`
	FileId string `json:"fileId"`
}

type DynamicResetMonstersRequest struct {
	Configs []model.SceneMonsterConfig `json:"configs"`
}

type ClientInitCompletedRequest struct {
	IsReEnter bool `json:"isReenter"`
}
