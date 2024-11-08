package object

// 用于与前端通信的对象数据
import (
	"github.com/ouyangzhongmin/gameserver/db/model"
	"github.com/ouyangzhongmin/gameserver/internal/game/constants"
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
	h.MaxLife = constants.CaculateLife(h.BaseLife, h.Strength)
	h.MaxMana = constants.CaculateLife(h.BaseMana, h.Intelligence)
	h.Attack = constants.CaculateAttack(h.AttrType, h.BaseAttack, h.Strength, h.Agility, h.Intelligence)
	h.Defense = constants.CaculateDefense(h.Defense, h.Agility)
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
	o.Avatar = o.Data.Avatar
	o.Grade = o.Data.Grade
	o.Level = o.Data.Level
	o.MonsterType = o.Data.MonsterType
	return o
}

func (m *MonsterObject) UpdateProperty() {
	m.MaxLife = constants.CaculateLife(m.Data.BaseLife, m.Data.Strength)
	m.MaxMana = constants.CaculateLife(m.Data.BaseMana, m.Data.Intelligence)
	//精确攻击力，不附带随机攻击力
	m.Attack = constants.CaculateAttack(m.Data.AttrType, m.Data.BaseAttack, m.Data.Strength, m.Data.Agility, m.Data.Intelligence)
	m.Defense = constants.CaculateDefense(m.Defense, m.Data.Agility)
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
	return m.Data.BaseDefense + m.Data.Agility*constants.DEFENSE_AGILITY_PERM
}

type SpellObject struct {
	GameObject
	Data      model.Spell        `json:"-"`
	Id        int64              `json:"id"`
	Animation string             `json:"animation" ` //
	StepTime  int                `json:"step_time" ` //飞行速度
	CdTime    int                `json:"cd_time" `   //cd间隔
	Buf       *model.BufferState `json:"-"`
	TargetPos shape.Vector3      `json:"target_pos"`
}

func NewSpellObject(data *model.Spell, buf *model.BufferState) *SpellObject {
	o := &SpellObject{
		GameObject: GameObject{},
		Data:       *data,
		Buf:        buf,
	}
	//构造一个id
	o.Id = time.Now().UnixMilli()%1000000*100 + int64(rand.Intn(100))
	return o
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
