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
	model.Monster
	MaxLife int64 `json:"max_life" db:"max_life" ` //
	MaxMana int64 `json:"max_mana" db:"max_mana" ` //
	Life    int64 `json:"life" db:"life" `         //
	Mana    int64 `json:"mana" db:"mana" `         //
	Defense int64 `json:"defense" db:"defense" `   //
	Attack  int64 `json:"attack" db:"attack" `     //
	State   constants.ActionState
}

func NewMonsterObject(data *model.Monster) *MonsterObject {
	o := &MonsterObject{
		GameObject: GameObject{},
		Monster:    *data,
	}
	o.UpdateProperty()
	//初始化的时候满血满蓝
	o.Life = o.MaxLife
	o.Mana = o.MaxMana
	//构造monster的id,通过时间戳去后5位再乘100+一个随机数
	o.Id = int64(int(time.Now().UnixMilli()%10000)*100 + rand.Intn(100))
	return o
}

func (m *MonsterObject) UpdateProperty() {
	m.MaxLife = constants.CaculateLife(m.BaseLife, m.Strength)
	m.MaxMana = constants.CaculateLife(m.BaseMana, m.Intelligence)
	m.Attack = constants.CaculateAttack(m.AttrType, m.BaseAttack, m.Strength, m.Agility, m.Intelligence)
	m.Defense = constants.CaculateDefense(m.Defense, m.Agility)
}

func (m *MonsterObject) IsAlive() bool {
	return m.Life > 0
}

type SpellObject struct {
	GameObject
	model.Spell
	Buf       *model.BufferState `json:"-"`
	TargetPos shape.Vector3      `json:"target_pos"`
}

func NewSpellObject(data *model.Spell, buf *model.BufferState) *SpellObject {
	o := &SpellObject{
		GameObject: GameObject{},
		Spell:      *data,
		Buf:        buf,
	}
	//构造一个id
	o.Id = int(time.Now().UnixMilli()%10000)*10 + rand.Intn(10)
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
