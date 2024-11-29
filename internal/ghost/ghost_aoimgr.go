package ghost

// AOI管理器
import (
	"github.com/ouyangzhongmin/gameserver/pkg/aoi"
	"github.com/ouyangzhongmin/gameserver/pkg/shape"
)

type ghostAoiMgr struct {
	aoi aoi.AOI
}

func newAoiMgr(areaWidth int, gridCount int) *ghostAoiMgr {
	mgr := &ghostAoiMgr{}
	//基于大格子算法的AOI，找的一个基于九宫格的aoi库直接使用了，自己也可以定义二维数组的方式实现
	mgr.aoi = aoi.NewGridManager(0, 0, areaWidth, gridCount)
	return mgr
}

func (m *ghostAoiMgr) Enter(entity *ghostEntity) {
	m.aoi.Add(float64(entity.GetPos().X), float64(entity.GetPos().Y), entity.GetUUID(), entity)
}

// 防止泄露 对象销毁时一定要调用Leave
func (m *ghostAoiMgr) Leave(entity *ghostEntity) {
	m.aoi.Delete(float64(entity.GetPos().X), float64(entity.GetPos().Y), entity.GetUUID())
}

func (m *ghostAoiMgr) Moved(entity *ghostEntity, x, y, oldX, oldY shape.Coord) {
	m.aoi.Moved(float64(x), float64(y), float64(oldX), float64(oldY), entity.GetUUID(), entity)
}

func (m *ghostAoiMgr) Search(x, y shape.Coord) []interface{} {
	result := m.aoi.Search(float64(x), float64(y))
	return result
}
