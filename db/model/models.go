package model

import (
	"time"
)

type Aiconfig struct {
	Id           int       `json:"id" db:"id" `                       //
	MonsterId    int64     `json:"monster_id" db:"monster_id" `       //
	PatrolRange  int       `json:"patrol_range" db:"patrol_range" `   //
	ChaseRange   int       `json:"chase_range" db:"chase_range" `     //
	AlertRange   int       `json:"alert_range" db:"alert_range" `     //警戒范围
	AutoBeatback int       `json:"auto_beatback" db:"auto_beatback" ` //1自动反击，0不反击
	Spells       string    `json:"spells" db:"spells" `               //拥有哪些技能
	Conds        string    `json:"conds" db:"conds" `                 //# 条件节点，满足条件时，会执行后面的act        'conds':{            # 配置方式：'节点名': [判断类型(>, <, =, %), 数值, 时间， 概率](未使用填-1)            'once': [['=', -1, 'act_1', -1, -1]],  # 立即执行act_1            'blood':[['<', '0.5', 'act_2', -1, -1]], # 血量小于0.5时执行act2            'time':[['%' ,'10', 'act_3', -1, 80]],  # 每隔10秒有0.8的概率执行一次act_3            ...        }
	CreateAt     time.Time `json:"-" db:"create_at" `                 //
	UpdateAt     time.Time `json:"-" db:"update_at" `                 //
}
type BufferState struct {
	Id                  int       `json:"id" db:"id" `                                       //
	Name                string    `json:"name" db:"name" `                                   //
	Animation           string    `json:"animation" db:"animation" `                         //
	BufType             int       `json:"buf_type" db:"buf_type" `                           //0 伤害类型，1治疗类型
	Damage              int       `json:"damage" db:"damage" `                               //负数就是治疗，正数为伤害
	EffectDurationTime  int       `json:"effect_duration_time" db:"effect_duration_time" `   //每次持续时间
	EffectDisappearTime int       `json:"effect_disappear_time" db:"effect_disappear_time" ` //每次结束后的消失时间
	EffectCnt           int       `json:"effect_cnt" db:"effect_cnt" `                       //生效次数
	CdTime              int       `json:"cd_time" db:"cd_time" `                             //cd时间
	Stackable           int       `json:"stackable" db:"stackable" `                         //是否可叠加
	CreateAt            time.Time `json:"-" db:"create_at" `                                 //
	UpdateAt            time.Time `json:"-" db:"update_at" `                                 //
}
type Hero struct {
	Id           int64     `json:"id" db:"id" `                     //
	Name         string    `json:"name" db:"name" `                 //
	Avatar       string    `json:"avatar" db:"avatar" `             //资源
	AttrType     int       `json:"attr_type" db:"attr_type" `       //属性类型:0 力量，1敏捷, 2智慧
	Uid          int64     `json:"uid" db:"uid" `                   //
	Experience   int64     `json:"experience" db:"experience" `     //
	Level        int       `json:"level" db:"level" `               //
	MaxLife      int64     `json:"max_life" db:"max_life" `         //
	MaxMana      int64     `json:"max_mana" db:"max_mana" `         //
	Defense      int64     `json:"defense" db:"defense" `           //
	Attack       int64     `json:"attack" db:"attack" `             //
	BaseLife     int64     `json:"base_life" db:"base_life" `       //
	BaseMana     int64     `json:"base_mana" db:"base_mana" `       //
	BaseDefense  int64     `json:"base_defense" db:"base_defense" ` //
	BaseAttack   int64     `json:"base_attack" db:"base_attack" `   //
	Strength     int64     `json:"strength" db:"strength" `         //
	Agility      int64     `json:"agility" db:"agility" `           //
	Intelligence int64     `json:"intelligence" db:"intelligence" ` //
	StepTime     int       `json:"step_time" db:"step_time" `       //移动速度
	SceneId      int       `json:"scene_id" db:"scene_id" `         //
	InitPosx     int       `json:"init_posx" db:"init_posx" `       //
	InitPosy     int       `json:"init_posy" db:"init_posy" `       //
	InitPosz     int       `json:"init_posz" db:"init_posz" `       //
	AttackRange  int       `json:"attack_range" db:"attack_range" ` //
	CreateAt     time.Time `json:"-" db:"create_at" `               //
	UpdateAt     time.Time `json:"-" db:"update_at" `               //
}
type Login struct {
	Id        int64     `json:"id" db:"id" `                 //
	Uid       int64     `json:"uid" db:"uid" `               //
	Remote    string    `json:"remote" db:"remote" `         //
	Ip        string    `json:"ip" db:"ip" `                 //
	Model     string    `json:"model" db:"model" `           //
	Imei      string    `json:"imei" db:"imei" `             //
	Os        string    `json:"os" db:"os" `                 //
	Appid     string    `json:"appid" db:"appid" `           //
	ChannelId string    `json:"channel_id" db:"channel_id" ` //
	LoginAt   int64     `json:"login_at" db:"login_at" `     //
	LogoutAt  int64     `json:"logout_at" db:"logout_at" `   //
	CreateAt  time.Time `json:"-" db:"create_at" `           //
}
type Monster struct {
	Id                 int64     `json:"id" db:"id" `                                     //
	Name               string    `json:"name" db:"name" `                                 //
	Avatar             string    `json:"avatar" db:"avatar" `                             //模型
	MonsterType        int       `json:"monster_type" db:"monster_type" `                 //0-怪物, 1-npc
	Level              int       `json:"level" db:"level" `                               //
	Grade              int       `json:"grade" db:"grade" `                               //级别：0"普通怪",1"小头目",2"精英怪",3"大BOSS",            4"变态怪", 5 "变态怪"
	AttrType           int       `json:"attr_type" db:"attr_type" `                       //属性类型:0 力量，1敏捷, 2智慧
	BaseLife           int64     `json:"base_life" db:"base_life" `                       //
	BaseMana           int64     `json:"base_mana" db:"base_mana" `                       //
	BaseDefense        int64     `json:"base_defense" db:"base_defense" `                 //
	BaseAttack         int64     `json:"base_attack" db:"base_attack" `                   //
	AttachAttackRandom int       `json:"attach_attack_random" db:"attach_attack_random" ` //附带的随机值
	Strength           int64     `json:"strength" db:"strength" `                         //
	Agility            int64     `json:"agility" db:"agility" `                           //
	Intelligence       int64     `json:"intelligence" db:"intelligence" `                 //
	RunStepTime        int       `json:"run_step_time" db:"run_step_time" `               //跑速度
	IdleStepTime       int       `json:"idle_step_time" db:"idle_step_time" `             //正常速度
	ChaseStepTime      int       `json:"chase_step_time" db:"chase_step_time" `           //追击速度
	EscapeStepTime     int       `json:"escape_step_time" db:"escape_step_time" `         //逃跑速度
	AttackRange        int       `json:"attack_range" db:"attack_range" `                 //攻击范围
	AttackDuration     int       `json:"attack_duration" db:"attack_duration" `           //攻击间隔
	Description        string    `json:"description" db:"description" `                   //简介
	CreateAt           time.Time `json:"-" db:"create_at" `                               //
	UpdateAt           time.Time `json:"-" db:"update_at" `                               //
}
type Online struct {
	Id        int       `json:"id" db:"id" `                 //
	UserCount int       `json:"user_count" db:"user_count" ` //
	Scenes    string    `json:"scenes" db:"scenes" `         //
	Time      int64     `json:"time" db:"time" `             //
	CreateAt  time.Time `json:"-" db:"create_at" `           //
	UpdateAt  time.Time `json:"-" db:"update_at" `           //
}
type Register struct {
	Id           int       `json:"id" db:"id" `                       //
	Uid          int64     `json:"uid" db:"uid" `                     //
	Remote       string    `json:"remote" db:"remote" `               //
	Ip           string    `json:"ip" db:"ip" `                       //
	Imei         string    `json:"imei" db:"imei" `                   //
	Os           string    `json:"os" db:"os" `                       //
	Model        string    `json:"model" db:"model" `                 //
	Appid        string    `json:"appid" db:"appid" `                 //
	ChannelId    string    `json:"channel_id" db:"channel_id" `       //
	RegisterType int       `json:"register_type" db:"register_type" ` //
	CreateAt     time.Time `json:"-" db:"create_at" `                 //
	UpdateAt     time.Time `json:"-" db:"update_at" `                 //
}
type Scene struct {
	Id        int       `json:"id" db:"id" `                 //
	Name      string    `json:"name" db:"name" `             //场景名称
	SceneType string    `json:"scene_type" db:"scene_type" ` //场景类型
	MapFile   string    `json:"map_file" db:"map_file" `     //场景资源名称
	Enterx    int       `json:"enterx" db:"enterx" `         //
	Entery    int       `json:"entery" db:"entery" `         //
	Enterz    int       `json:"enterz" db:"enterz" `         //
	CreateAt  time.Time `json:"-" db:"create_at" `           //
	UpdateAt  time.Time `json:"-" db:"update_at" `           //
}
type SceneDoor struct {
	Id            int       `json:"id" db:"id" `                           //
	Name          string    `json:"name" db:"name" `                       //
	SceneId       int       `json:"scene_id" db:"scene_id" `               //所在场景
	Posx          int       `json:"posx" db:"posx" `                       //
	Posy          int       `json:"posy" db:"posy" `                       //
	Posz          int       `json:"posz" db:"posz" `                       //
	TargetSceneId int       `json:"target_scene_id" db:"target_scene_id" ` //目标场景
	DestPosx      int       `json:"dest_posx" db:"dest_posx" `             //
	DestPosy      int       `json:"dest_posy" db:"dest_posy" `             //
	DestPosz      int       `json:"dest_posz" db:"dest_posz" `             //
	CreateAt      time.Time `json:"-" db:"create_at" `                     //
	UpdateAt      time.Time `json:"-" db:"update_at" `                     //
}
type SceneMonsterConfig struct {
	Id        int       `json:"id" db:"id" `                 //
	SceneId   int       `json:"scene_id" db:"scene_id" `     //
	MonsterId int64     `json:"monster_id" db:"monster_id" ` //
	Total     int       `json:"total" db:"total" `           //
	Reborn    int       `json:"reborn" db:"reborn" `         //重生间隔
	Bornx     int       `json:"bornx" db:"bornx" `           //
	Borny     int       `json:"borny" db:"borny" `           //
	Bornz     int       `json:"bornz" db:"bornz" `           //
	ARange    int       `json:"a_range" db:"a_range" `       //活动范围
	CreateAt  time.Time `json:"-" db:"create_at" `           //
	UpdateAt  time.Time `json:"-" db:"update_at" `           //
}
type SceneNpcConfig struct {
	Id       int       `json:"id" db:"id" `             //
	SceneId  int       `json:"scene_id" db:"scene_id" ` //
	NpcId    int64     `json:"npc_id" db:"npc_id" `     //
	Bornx    int       `json:"bornx" db:"bornx" `       //
	Borny    int       `json:"borny" db:"borny" `       //
	Bornz    int       `json:"bornz" db:"bornz" `       //
	ARange   int       `json:"a_range" db:"a_range" `   //活动范围
	CreateAt time.Time `json:"-" db:"create_at" `       //
	UpdateAt time.Time `json:"-" db:"update_at" `       //
}
type Spell struct {
	Id            int       `json:"id" db:"id" `                           //
	Name          string    `json:"name" db:"name" `                       //
	FlyAnimation  string    `json:"fly_animation" db:"fly_animation" `     //
	Damage        int64     `json:"damage" db:"damage" `                   //
	Mana          int64     `json:"mana" db:"mana" `                       //消耗
	FlyStepTime   int       `json:"fly_step_time" db:"fly_step_time" `     //飞行速度
	CdTime        int       `json:"cd_time" db:"cd_time" `                 //cd间隔
	BufId         int       `json:"buf_id" db:"buf_id" `                   //
	IsRangeAttack int       `json:"is_range_attack" db:"is_range_attack" ` //是否范围攻击
	AttackRange   int       `json:"attack_range" db:"attack_range" `       //攻击范围，单体为0
	SpellType     int       `json:"spell_type" db:"spell_type" `           //技能类型: 0 对敌人，1，对自己, 2，对友军
	Description   string    `json:"description" db:"description" `         //
	CreateAt      time.Time `json:"-" db:"create_at" `                     //
	UpdateAt      time.Time `json:"-" db:"update_at" `                     //
}
type ThirdAccount struct {
	Id           int64     `json:"id" db:"id" `                       //
	ThirdAccount string    `json:"third_account" db:"third_account" ` //wx的openid
	Uid          int64     `json:"uid" db:"uid" `                     //
	Platform     string    `json:"platform" db:"platform" `           //
	ThirdName    string    `json:"third_name" db:"third_name" `       //
	HeadUrl      string    `json:"head_url" db:"head_url" `           //
	Sex          int       `json:"sex" db:"sex" `                     //
	CreateAt     time.Time `json:"-" db:"create_at" `                 //
	UpdateAt     time.Time `json:"-" db:"update_at" `                 //
}
type User struct {
	Id          int64     `json:"id" db:"id" `                     //
	Algo        string    `json:"algo" db:"algo" `                 //
	Hash        string    `json:"hash" db:"hash" `                 //
	Role        int       `json:"role" db:"role" `                 //账号类型: 1-管理员账号，2- 第三方平台账号
	Coin        int64     `json:"coin" db:"coin" `                 //金币
	IsOnline    int       `json:"is_online" db:"is_online" `       //
	Salt        string    `json:"salt" db:"salt" `                 //盐值
	LastLoginat int64     `json:"last_loginat" db:"last_loginat" ` //
	PrivKeyey   string    `json:"priv_keyey" db:"priv_keyey" `     //
	PubKey      string    `json:"pub_key" db:"pub_key" `           //
	Debug       int       `json:"debug" db:"debug" `               //
	Status      int       `json:"status" db:"status" `             //0禁用，1启用
	IsGuest     int       `json:"is_guest" db:"is_guest" `         //
	CreateAt    time.Time `json:"-" db:"create_at" `               //
	UpdateAt    time.Time `json:"-" db:"update_at" `               //
}
