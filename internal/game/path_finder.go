package game

import (
	"fmt"
	"github.com/ouyangzhongmin/gameserver/pkg/astar"
)

// 这个非线程安全，需要单线程一个执行
type PathFinder struct {
	cachedPaths map[string][][]int32
	a           *astar.AStar
}

func NewPathFinder(grids [][]int32) *PathFinder {
	f := &PathFinder{}
	f.cachedPaths = make(map[string][][]int32)
	f.a = astar.NewAstar(grids)
	return f
}

func (f *PathFinder) FindPath(sx, sy, ex, ey int) (path [][]int32, err error) {
	key := fmt.Sprintf("%d_%d_%d_%d", sy, sx, ey, ex)
	if paths, ok := f.cachedPaths[key]; ok {
		return paths, nil
	} else {
		path, _, _, err = f.a.FindPath([]int32{int32(sy), int32(sx)}, []int32{int32(ey), int32(ex)})
		if path != nil {
			f.cachedPaths[key] = paths
		}
		return path, err
	}
}
