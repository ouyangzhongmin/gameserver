package game

// AOI管理器
import (
	"github.com/ouyangzhongmin/gameserver/pkg/aoi"
	"github.com/ouyangzhongmin/gameserver/pkg/coord"
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

func (m *aoiMgr) Moved(entity IMovableEntity, x, y, oldX, oldY coord.Coord) {
	m.aoi.Moved(float64(x), float64(y), float64(oldX), float64(oldY), entity.GetUUID(), entity)
}

func (m *aoiMgr) Search(x, y coord.Coord) []interface{} {
	result := m.aoi.Search(float64(x), float64(y))
	return result
}
