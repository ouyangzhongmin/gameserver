package aoi

import (
	"fmt"
	"sync"
)

var (
	// Define direction vectors for the eight directions by populating dx and dy arrays.
	dx = []int{-1, -1, -1, 0, 0, 1, 1, 1}
	dy = []int{-1, 0, 1, -1, 1, -1, 0, 1}
)

// Grid represents a grid with a unique identifier and entities within it.
type Grid struct {
	GID      int      // Grid ID
	Entities sync.Map // Entities within the current grid
}

// GridManager implements AOI (Area of Interest) using a rectangular grid.
type GridManager struct {
	StartX    int // X-coordinate of the left boundary of the AOI
	StartY    int // Y-coordinate of the upper boundary of the AOI
	AreaWidth int // Width of each grid (assuming square grids)
	GridCount int // Number of grids in each row/column
	grids     map[int]*Grid
	pool      sync.Pool
}

// NewGrid creates a new grid with the specified ID.
func NewGrid(gid int) *Grid {
	return &Grid{
		GID: gid,
	}
}

// NewGridManager initializes a new GridManager with the specified parameters.
func NewGridManager(startX, startY, areaWidth, gridCount int) AOI {
	manager := &GridManager{
		StartX:    startX,
		StartY:    startY,
		AreaWidth: areaWidth,
		GridCount: gridCount,
		grids:     make(map[int]*Grid),
	}
	manager.pool.New = func() interface{} {
		return make([]*Grid, 0, 9)
	}

	// Initialize grids with unique IDs
	for y := 0; y < gridCount; y++ {
		for x := 0; x < gridCount; x++ {
			// Grid ID calculation: ID = IDy * nx + IDx (using grid coordinates to obtain grid ID)
			gID := y*gridCount + x
			manager.grids[gID] = NewGrid(gID)
		}
	}

	return manager
}

// gridWidth calculates the width of each grid.
func (g *GridManager) gridWidth() int {
	return g.AreaWidth / g.GridCount
}

// getGIDByPos calculates the grid ID based on the given coordinates.
func (g *GridManager) getGIDByPos(x, y float64) int {
	gx := (int(x) - g.StartX) / g.gridWidth()
	gy := (int(y) - g.StartY) / g.gridWidth()

	return gy*g.GridCount + gx
}

// getSurroundGrids retrieves information about the surrounding nine grids based on the given grid ID.
func (g *GridManager) getSurroundGrids(gID int) []*Grid {
	grids := g.pool.Get().([]*Grid)
	defer func() {
		grids = grids[:0]
		g.pool.Put(grids)
	}()

	if _, ok := g.grids[gID]; !ok {
		return grids
	}
	grids = append(grids, g.grids[gID])

	// Calculate the coordinates of the grid based on the grid ID
	x, y := gID%g.GridCount, gID/g.GridCount

	// Add information about the eight neighboring grids
	for i := 0; i < 8; i++ {
		newX := x + dx[i]
		newY := y + dy[i]

		if newX >= 0 && newX < g.GridCount && newY >= 0 && newY < g.GridCount {
			grids = append(grids, g.grids[newY*g.GridCount+newX])
		}
	}

	return grids
}

// Add adds an entity to the appropriate grid based on its coordinates.
func (g *GridManager) Add(x, y float64, key string, data interface{}) {
	entity := entityPool.Get().(*Entity)
	entity.X = x
	entity.Y = y
	entity.Key = key
	entity.Data = data

	ID := g.getGIDByPos(x, y)
	grid := g.grids[ID]
	grid.Entities.Store(key, entity)
	//fmt.Println("GridManager add:", key, x, y)
}

// Delete removes an entity from the grid based on its coordinates.
func (g *GridManager) Delete(x, y float64, key string) {
	ID := g.getGIDByPos(x, y)
	grid := g.grids[ID]

	if e, ok := grid.Entities.Load(key); ok {
		entity := e.(*Entity)
		grid.Entities.Delete(key)
		entity.Data = nil //要清理掉引用
		entityPool.Put(entity)
		//fmt.Println("GridManager delete:", key, x, y)
	} else {
		fmt.Println("GridManager delete not find:", key, x, y)
	}
}

func (g *GridManager) Moved(x, y, oldx, oldy float64, key string, data interface{}) {
	oldGid := g.getGIDByPos(float64(oldx), float64(oldy))
	newgid := g.getGIDByPos(float64(x), float64(y))
	//fmt.Println("GridManager Moved :", key, x, y, oldx, oldy)
	if oldGid != newgid {
		if oldx > 0 && oldy > 0 {
			g.Delete(oldx, oldy, key)
		}
		g.Add(x, y, key, data)
	}
}

// Search retrieves a list of entity keys within the specified coordinates' range.
func (g *GridManager) Search(x, y float64) []interface{} {
	result := resultPool.Get().([]interface{})
	defer func() {
		result = result[:0]
		resultPool.Put(result)
	}()

	ID := g.getGIDByPos(x, y)
	grids := g.getSurroundGrids(ID)

	// Collect entity keys from the surrounding grids
	for _, grid := range grids {
		grid.Entities.Range(func(_, value interface{}) bool {
			result = append(result, value.(*Entity).Data)
			return true
		})
	}

	return result
}
