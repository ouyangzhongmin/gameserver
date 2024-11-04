package game

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/ouyangzhongmin/gameserver/pkg/astar"
	"github.com/ouyangzhongmin/gameserver/pkg/shape"
	"io"
	"math/rand"
	"sync"
)

// 二维格子数据
type BlockInfo struct {
	blockTable [][]int32
	colCount   uint16
	rowCount   uint16
	pool       sync.Pool
}

func NewBlockInfo() *BlockInfo {
	b := &BlockInfo{}
	b.pool = sync.Pool{New: func() interface{} {
		return astar.NewAstar(b.blockTable)
	}}
	return b
}

func (b *BlockInfo) GetHeight() uint32 {
	return uint32(b.rowCount)
}

func (b *BlockInfo) GetWidth() uint32 {
	return uint32(b.colCount)
}

func (b *BlockInfo) GetBlockTable() [][]int32 {
	return b.blockTable
}

func (b *BlockInfo) IsWalkable(i, j int32) bool {
	if j < int32(b.rowCount) && j >= 0 && i >= 0 && i < int32(b.colCount) {
		return b.blockTable[j][i] == 0
	}
	return false
}

func (b *BlockInfo) ReadFrom(bytebuffer *bytes.Buffer) error {
	err := binary.Read(bytebuffer, binary.BigEndian, &b.colCount)
	if err != nil {
		if err != io.EOF {
			fmt.Println("Error reading uint16:", err)
			return err
		}
	}
	err = binary.Read(bytebuffer, binary.BigEndian, &b.rowCount)
	if err != nil {
		if err != io.EOF {
			fmt.Println("Error reading uint16:", err)
			return err
		}
	}
	b.blockTable = make([][]int32, b.rowCount)
	for i := range b.blockTable {
		b.blockTable[i] = make([]int32, b.colCount)
	}
	for i := 0; i < int(b.rowCount); i++ {
		for j := 0; j < int(b.colCount); j++ {
			bt, err := bytebuffer.ReadByte()
			if err != nil {
				fmt.Println("Error reading byte:", err)
				return err
			}
			if bt == 0 { //编辑器是127为可以走，0为不能走
				b.blockTable[i][j] = 1
			} else {
				b.blockTable[i][j] = 0
			}
		}
	}
	return nil
}

func (b *BlockInfo) FindPath(sx, sy, ex, ey int32) (path [][]int32, block, turn int, err error) {
	a := b.pool.Get().(*astar.AStar)
	defer func() {
		a.Clean()
		b.pool.Put(a)
	}()
	path, block, turn, err = a.FindPath([]int32{sy, sx}, []int32{ey, ex})
	return path, block, turn, err
}

func (b *BlockInfo) GetRandomXY(rect shape.Rect, cnt int) (shape.Coord, shape.Coord, error) {
	if cnt <= 0 {
		return 0, 0, errors.New("没有找到可用的随机坐标")
	}
	rx := rand.Intn(int(rect.Width)) + int(rect.X)
	ry := rand.Intn(int(rect.Height)) + int(rect.Y)
	if rx < 0 {
		rx = 0
	}
	if ry < 0 {
		ry = 0
	}
	if b.IsWalkable(int32(rx), int32(ry)) {
		return shape.Coord(rx), shape.Coord(ry), nil
	}
	return b.GetRandomXY(rect, cnt-1)
}
