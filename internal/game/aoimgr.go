package game

// AOI管理器
import (
	"github.com/ouyangzhongmin/gameserver/internal/game/constants"
	"github.com/ouyangzhongmin/gameserver/pkg/aoi"
	"github.com/ouyangzhongmin/gameserver/pkg/shape"
)

type aoiMgr struct {
	aoi aoi.AOI
}

func newAoiMgr(areaWidth int, gridCount int) *aoiMgr {
	mgr := &aoiMgr{}
	//基于大格子算法的AOI，找的一个基于九宫格的aoi库直接使用了，自己也可以定义二维数组的方式实现
	mgr.aoi = aoi.NewGridManager(0, 0, areaWidth, gridCount)
	return mgr
}

func (m *aoiMgr) Enter(entity IMovableEntity) {
	m.aoi.Add(float64(entity.GetPos().X), float64(entity.GetPos().Y), entity.GetUUID(), entity)
}

// 防止泄露 对象销毁时一定要调用Leave
func (m *aoiMgr) Leave(entity IMovableEntity) {
	m.aoi.Delete(float64(entity.GetPos().X), float64(entity.GetPos().Y), entity.GetUUID())
}

func (m *aoiMgr) Moved(entity IMovableEntity, oldX, oldY shape.Coord) {
	oldbx := oldX / constants.SCENE_BLOCK_TO_SHIFT
	oldby := oldY / constants.SCENE_BLOCK_TO_SHIFT
	bx := entity.GetPos().X / constants.SCENE_BLOCK_TO_SHIFT
	by := entity.GetPos().Y / constants.SCENE_BLOCK_TO_SHIFT

	if oldbx != bx || oldby != by {
		if oldbx > 0 && oldby > 0 {
			m.aoi.Delete(float64(oldX), float64(oldY), entity.GetUUID())
		}
		m.aoi.Add(float64(entity.GetPos().X), float64(entity.GetPos().Y), entity.GetUUID(), entity)
	}
}

func (m *aoiMgr) Search(x, y shape.Coord) []interface{} {
	result := m.aoi.Search(float64(x), float64(y))
	return result
}
