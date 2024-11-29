package protocol

import (
	"github.com/ouyangzhongmin/gameserver/db/model"
	"github.com/ouyangzhongmin/gameserver/internal/game/object"
	"github.com/ouyangzhongmin/gameserver/pkg/path"
	"github.com/ouyangzhongmin/gameserver/pkg/shape"
	"github.com/ouyangzhongmin/nano/session"
)

type ThirdUserLoginRequest struct {
	Platform    string `json:"platform"`     //三方平台/渠道
	AppID       string `json:"appId"`        //用户来自于哪一个应用
	ChannelID   string `json:"channelId"`    //用户来自于哪一个渠道
	Device      Device `json:"device"`       //设备信息
	Name        string `json:"name"`         //微信平台名
	OpenID      string `json:"openid"`       //微信平台openid
	AccessToken string `json:"access_token"` //微信AccessToken
}

type LoginInfo struct {
	// 三方登录字段
	Platform     string `json:"platform"`      //三方平台
	ThirdAccount string `json:"third_account"` //三方平台唯一ID
	ThirdName    string `json:"account"`       //三方平台账号名

	Token      string `json:"token"`       //用户Token
	ExpireTime int64  `json:"expire_time"` //Token过期时间

	AccountID int64 `json:"acId"` //用户的uuid,即user表的pk

	GameServerIP   string `json:"ip"` //游戏服的ip&port
	GameServerPort int    `json:"port"`
}

type UserLoginResponse struct {
	Code int32     `json:"code"` //状态码
	Data LoginInfo `json:"data"`
}

type LoginRequest struct {
	AppID     string `json:"appId"`     //用户来自于哪一个应用
	ChannelID string `json:"channelId"` //用户来自于哪一个渠道
	IMEI      string `json:"imei"`
	Device    Device `json:"device"`
}

type ClientConfig struct {
	Version     string `json:"version"`
	Android     string `json:"android"`
	IOS         string `json:"ios"`
	Heartbeat   int    `json:"heartbeat"`
	ForceUpdate bool   `json:"forceUpdate"`

	Title string `json:"title"` // 分享标题
	Desc  string `json:"desc"`  // 分享描述

	Daili1 string `json:"daili1"`
	Daili2 string `json:"daili2"`
	Kefu1  string `json:"kefu1"`

	AppId  string `json:"appId"`
	AppKey string `json:"appKey"`
}

type LoginResponse struct {
	Code     int          `json:"code"`
	Name     string       `json:"name"`
	Uid      int64        `json:"uid"`
	HeadUrl  string       `json:"head_url"`
	FangKa   int          `json:"fangka"`
	Sex      int          `json:"sex"` //[0]未知 [1]男 [2]女
	IP       string       `json:"ip"`
	Port     int          `json:"port"`
	PlayerIP string       `json:"playerIp"`
	Config   ClientConfig `json:"config"`
	Messages []string     `json:"messages"`
	HeroList []model.Hero `json:"hero_list"`
	Debug    int          `json:"debug"`
	IsGuest  int          `json:"is_guest"`
}

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
	SceneId int `json:"scene_id"`
	Width   int `json:"width"`
	Height  int `json:"height"`
	Enterx  int `json:"enter_x"`
	EnterY  int `json:"enter_y"`
}

type Cell struct {
	CellID      int              `json:"cell_id"`
	SceneId     int              `json:"scene_id"`
	Bounds      shape.Rect       `json:"bounds"`    // 边界划分
	EdgeSize    int              `json:"edge_size"` // 边缘宽度
	RemoteAddr  string           `json:"remote_addr"`
	IsFirstCell bool             `json:"is_first_cell"`
	IsNew       bool             `json:"is_new"`
	Session     *session.Session `json:"-"`
}

type SceneCelllsRequest struct {
	SceneId int `json:"scene_id"`
	CellId  int `json:"cell_id"`
	Cells   []*Cell
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

type EncryptTest struct {
	Payload string `json:"payload"`
	Key     string `json:"key"`
}

type EncryptTestTest struct {
	Result string `json:"result"`
}
