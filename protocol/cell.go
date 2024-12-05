package protocol

import (
	"github.com/ouyangzhongmin/gameserver/db/model"
	"github.com/ouyangzhongmin/gameserver/internal/game/object"
	"github.com/ouyangzhongmin/gameserver/pkg/path"
	"github.com/ouyangzhongmin/gameserver/pkg/shape"
)

type ChooseHeroResponse struct {
	model.Hero
}

type ChooseHeroRequest struct {
	Uid    int64  `json:"uid"`
	HeroId int64  `json:"hero_id"`
	IP     string `json:"ip"`
}

type CreateHeroRequest struct {
	Uid      int64  `json:"uid"`
	Avatar   string `json:"avatar"`
	Name     string `json:"name"`
	AttrType int    `json:"attr_type"`
}

type HeroChangeSceneRequest struct {
	Uid     int64 `json:"uid"`
	HeroId  int64 `json:"hero_id"`
	SceneId int   `json:"scene_id"`
}

type RegisterSceneCellRequest struct {
	SceneId    int    `json:"scene_id"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Enterx     int    `json:"enter_x"`
	EnterY     int    `json:"enter_y"`
	RemoteAddr string `json:"remote_addr"`
	GateAddr   string `json:"gate_addr"`
}

type Cell struct {
	CellID      int        `json:"cell_id"`
	SceneId     int        `json:"scene_id"`
	Bounds      shape.Rect `json:"bounds"`    // 边界划分
	EdgeSize    int        `json:"edge_size"` // 边缘宽度
	RemoteAddr  string     `json:"remote_addr"`
	GateAddr    string     `json:"gate_addr"`
	IsFirstCell bool       `json:"is_first_cell"`
	IsNew       bool       `json:"is_new"`
	//Session     *session.Session `json:"-"`
}

type SceneCelllsRequest struct {
	SceneId int `json:"scene_id"`
	CellId  int `json:"cell_id"`
	Cells   []*Cell
}

type CheckCellHealthRequest struct {
	V int `json:"v"`
}

type MigrateMonsterRequest struct {
	SceneId       int                       `json:"scene_id"`
	CellId        int                       `json:"cell_id"`
	FromCellId    int                       `json:"from_cell_id"`
	MonsterObject *object.MonsterObject     `json:"monster_object"`
	MonsterData   *model.Monster            `json:"monster_data"`
	Cfg           *model.SceneMonsterConfig `json:"cfg"`
	AiData        *model.Aiconfig           `json:"ai_data"`
	MovableRect   shape.Rect                `json:"movable_rect"`
	PreparePaths  *path.SerialPaths         //预制的移动路径
	XViewRange    int                       `json:"x_view_range"`
	YViewRange    int                       `json:"y_view_range"`
}

type CreateGhostMonsterReq struct {
	SceneId       int                       `json:"scene_id"`
	CellId        int                       `json:"cell_id"`
	FromCellId    int                       `json:"from_cell_id"`
	MonsterObject *object.MonsterObject     `json:"monster_object"`
	MonsterData   *model.Monster            `json:"monster_data"`
	Cfg           *model.SceneMonsterConfig `json:"cfg"`
	MovableRect   shape.Rect                `json:"movable_rect"`
	XViewRange    int                       `json:"x_view_range"`
	YViewRange    int                       `json:"y_view_range"`
}

type RemoveGhostMonsterReq struct {
	SceneId   int   `json:"scene_id"`
	CellId    int   `json:"cell_id"`
	MonsterId int64 `json:"monster_id"`
}

type SyncGhostMonsterReq struct {
	SceneId       int                   `json:"scene_id"`
	CellId        int                   `json:"cell_id"`
	MonsterId     int64                 `json:"monster_id"`
	MonsterObject *object.MonsterObject `json:"monster_object"`
}

type SendMsgFromGhostReq struct {
	SceneId int         `json:"scene_id"`
	CellId  int         `json:"cell_id"`
	HeroId  int64       `json:"hero_id"`
	Route   string      `json:"route"`
	Msg     interface{} `json:"msg"`
}

type BroadcastToGhostReq struct {
	SceneId    int         `json:"scene_id"`
	CellId     int         `json:"cell_id"`
	EntityId   int64       `json:"entity_id"`
	EntityType int         `json:"entity_type"`
	Route      string      `json:"route"`
	Msg        interface{} `json:"msg"`
}

type GhostMonsterBeenHurtedReq struct {
	SceneId   int   `json:"scene_id"`
	CellId    int   `json:"cell_id"`
	MonsterId int64 `json:"monster_id"`
	Damage    int64 `json:"damage"`
}

type GhostMonsterBeenAttacedReq struct {
	SceneId      int   `json:"scene_id"`
	CellId       int   `json:"cell_id"`
	MonsterId    int64 `json:"monster_id"`
	AttackerId   int64 `json:"attacker_id"`
	AttackerType int   `json:"attacker_type"`
}

type MigrateHeroRequest struct {
	SceneId        int                `json:"scene_id"`
	CellId         int                `json:"cell_id"`
	FromCellId     int                `json:"from_cell_id"`
	HeroObject     *object.HeroObject `json:"hero_object"`
	HeroData       *model.Hero        `json:"hero_data"`
	XViewRange     int                `json:"x_view_range"`
	YViewRange     int                `json:"y_view_range"`
	ViewListIds    []string           `json:"view_list_ids"`
	CanSeemeIds    []string           `json:"can_seeme_ids"`
	TracePath      [][]int32          `json:"trace_path"`       //当前移动路径
	TraceIndex     int                `json:"trace_index"`      //当前已移动到第几步
	TraceTotalTime int64              `json:"trace_total_time"` //当前移动的总时间
	TargetX        int                `json:"target_x"`
	TargetY        int                `json:"target_y"`
	TargetZ        int                `json:"target_z"`
}
