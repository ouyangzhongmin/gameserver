package object

// 用于与前端通信的对象数据
import (
	"github.com/ouyangzhongmin/gameserver/constants"
	"github.com/ouyangzhongmin/gameserver/db/model"
	"github.com/ouyangzhongmin/gameserver/pkg/shape"
	"math/rand"
	"time"
)

type GameObject struct {
	Posx shape.Coord `json:"pos_x"`
	Posy shape.Coord `json:"pos_y"`
	Posz shape.Coord `json:"pos_z"`
	Uuid string      `json:"uuid"`
}

type HeroObject struct {
	model.Hero
	GameObject
	Life  int64 `json:"life" db:"life" ` //
	Mana  int64 `json:"mana" db:"mana" ` //
	State constants.ActionState
}

func NewHeroObject(data *model.Hero) *HeroObject {
	o := &HeroObject{
		Hero:       *data,
		GameObject: GameObject{},
	}
	o.Posx = shape.Coord(o.InitPosx)
	o.Posy = shape.Coord(o.InitPosy)
	o.Posz = shape.Coord(o.InitPosz)
	o.UpdateProperty()
	//初始化的时候满血满蓝
	o.Life = o.MaxLife
	o.Mana = o.MaxMana
	return o
}

func (h *HeroObject) UpdateProperty() {
	h.MaxLife = CaculateLife(h.BaseLife, h.Strength)
	h.MaxMana = CaculateLife(h.BaseMana, h.Intelligence)
	h.Attack = CaculateAttack(h.AttrType, h.BaseAttack, h.Strength, h.Agility, h.Intelligence)
	h.Defense = CaculateDefense(h.Defense, h.Agility)
}

func (h *HeroObject) IsAlive() bool {
	return h.Life > 0
}

type MonsterObject struct {
	GameObject
	Data        model.Monster         `json:"-"` //这个不用传递给前端，节省数据
	Name        string                `json:"name"`
	Id          int64                 `json:"id"`
	Avatar      string                `json:"avatar" db:"avatar" `             //模型
	MonsterType int                   `json:"monster_type" db:"monster_type" ` //0-怪物, 1-npc
	Level       int                   `json:"level" db:"level" `               //
	Grade       int                   `json:"grade" db:"grade" `               //级别：0"普通怪",1"小头目",2"精英怪",3"大BOSS",            4"变态怪", 5 "变态怪"
	MaxLife     int64                 `json:"max_life" db:"max_life" `         //
	MaxMana     int64                 `json:"max_mana" db:"max_mana" `         //
	Life        int64                 `json:"life" db:"life" `                 //
	Mana        int64                 `json:"mana" db:"mana" `                 //
	Defense     int64                 `json:"defense" db:"defense" `           //
	Attack      int64                 `json:"attack" db:"attack" `             //
	Dir         int                   `json:"dir"`
	State       constants.ActionState `json:"state"`
}

func NewMonsterObject(data *model.Monster, offset int) *MonsterObject {
	o := &MonsterObject{
		GameObject: GameObject{},
		Data:       *data,
	}
	o.UpdateProperty()
	//初始化的时候满血满蓝
	o.Life = o.MaxLife
	o.Mana = o.MaxMana
	//构造monster的id
	if offset <= 0 {
		offset = rand.Intn(100)
	}
	//注意这里的o.Data.Id不能超过10万，否则offsetId会被截取为0
	offsetId := (o.Data.Id * 10000) & 0xFFFFFFFF
	o.Id = time.Now().Unix()%1000000 + int64(offsetId) + int64(offset)
	o.Name = o.Data.Name
	o.Avatar = o.Data.Avatar
	o.Grade = o.Data.Grade
	o.Level = o.Data.Level
	o.MonsterType = o.Data.MonsterType
	return o
}

func (m *MonsterObject) UpdateProperty() {
	m.MaxLife = CaculateLife(m.Data.BaseLife, m.Data.Strength)
	m.MaxMana = CaculateLife(m.Data.BaseMana, m.Data.Intelligence)
	//精确攻击力，不附带随机攻击力
	m.Attack = CaculateAttack(m.Data.AttrType, m.Data.BaseAttack, m.Data.Strength, m.Data.Agility, m.Data.Intelligence)
	m.Defense = CaculateDefense(m.Defense, m.Data.Agility)
}

func (m *MonsterObject) IsAlive() bool {
	return m.Life > 0
}

// 附带了随机攻击力计算的函数
func (m *MonsterObject) GetAttack() int64 {
	var randAtt int64 = 0
	if m.Data.AttachAttackRandom > 0 {
		randAtt = int64(rand.Intn(m.Data.AttachAttackRandom))
	}
	return m.Attack + randAtt
}

func (m *MonsterObject) GetDefense() int64 {
	return m.Data.BaseDefense + m.Data.Agility*DEFENSE_AGILITY_PERM
}

type SpellObject struct {
	GameObject
	Data         model.Spell        `json:"-"`
	Id           int64              `json:"id"`
	SpellId      int                `json:"-"`
	Name         string             `json:"name" `          //
	Description  string             `json:"description" `   //
	FlyAnimation string             `json:"fly_animation" ` //
	FlyStepTime  int                `json:"fly_step_time" ` //飞行速度
	SpellType    int                `json:"spell_type" `    //飞行速度
	CdTime       int                `json:"cd_time" `       //cd间隔
	CurCdTime    int                `json:"cur_cd_time" `   //当前cd间隔
	Buf          *model.BufferState `json:"-"`
	TargetPos    shape.Vector3      `json:"target_pos"`
	CasterId     int64              `json:"caster_id"`
	CasterType   int                `json:"caster_type"`
	TargetId     int64              `json:"target_id"`
	TargetType   int                `json:"target_type"`
}

func NewSpellObject(data *model.Spell, buf *model.BufferState) *SpellObject {
	o := &SpellObject{
		GameObject: GameObject{},
		Data:       *data,
		Buf:        buf,
	}
	//构造一个id
	o.SpellId = data.Id
	o.Name = data.Name
	o.Description = data.Description
	o.FlyAnimation = data.FlyAnimation
	o.FlyStepTime = data.FlyStepTime
	o.SpellType = data.SpellType
	o.CdTime = data.CdTime
	return o
}

func (s *SpellObject) GenId() {
	s.Id = time.Now().UnixMilli()%1000000*100 + int64(rand.Intn(100))
}

func (s *SpellObject) ResetCDTime() {
	s.CurCdTime = s.CdTime
}

func (s *SpellObject) Update(elapsedTime int64) {
	s.CurCdTime -= int(elapsedTime)
	if s.CurCdTime <= 0 {
		s.CurCdTime = 0
	}
}

type BufferObject struct {
	model.BufferState
	CurCnt      int   `json:"cur_cnt"`      //当前第几次伤害
	ElapsedTime int64 `json:"elapsed_time"` //已经过的时间
	TotalTime   int64 `json:"total_time"`   //总持续时间
}

func NewBufferObject(data *model.BufferState) *BufferObject {
	o := &BufferObject{
		BufferState: *data,
	}
	return o
}
