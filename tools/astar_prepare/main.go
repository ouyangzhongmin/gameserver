package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ouyangzhongmin/gameserver/internal/game"
	"github.com/ouyangzhongmin/gameserver/pkg/fileutil"
	"github.com/ouyangzhongmin/gameserver/pkg/path"
	"github.com/ouyangzhongmin/gameserver/pkg/shape"
	"github.com/urfave/cli"
	"os"
	"strconv"
	"strings"
)

func main() {
	app := cli.NewApp()

	// base application info
	app.Name = "astar prepare tool"
	app.Author = "Oyzm"
	app.Version = "0.0.1"
	app.Copyright = "oyzm team reserved"
	app.Usage = "astar prepare tool"

	// flags
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "mapfile,m",
			Value: "xinshoucun",
			Usage: "指定的map",
		},
		cli.StringFlag{
			Name:  "rect,r",
			Value: "15,140,50",
			Usage: "指定预生成路径的范围",
		},
		cli.IntFlag{
			Name:  "count,c",
			Value: 100,
			Usage: "数量",
		},
	}

	app.Action = serve
	app.Run(os.Args)
}

func serve(c *cli.Context) error {
	mapFile := c.String("mapfile")
	rectstr := c.String("rect")
	count := c.Int("count")
	rectArr := strings.Split(rectstr, ",")
	if len(rectArr) != 3 {
		panic("rect 格式错误")
	}
	bornx, err := strconv.Atoi(rectArr[0])
	borny, err := strconv.Atoi(rectArr[1])
	arange, err := strconv.Atoi(rectArr[2])
	if err != nil {
		panic(err)
	}
	rect := shape.Rect{
		X:      int64(bornx - arange),
		Y:      int64(borny - arange),
		Width:  int64(arange * 2),
		Height: int64(arange * 2),
	}
	if rect.X < 0 {
		rect.X = 0
	}
	if rect.Y < 0 {
		rect.Y = 0
	}
	blockInfo := game.NewBlockInfo()
	buf, err := fileutil.ReadFile(fileutil.FindResourcePth(fmt.Sprintf("cmd/game/blocks/%s.block", mapFile)))
	if err != nil {
		panic(err)
	}
	err = blockInfo.ReadFrom(bytes.NewBuffer(buf))
	if err != nil {
		panic(err)
	}
	fmt.Printf("开始处理map:%s, rect:%s \n", mapFile, rectstr)
	spaths := make([]*path.SerialPaths, 0)
	for id := 1; id <= count; id++ {
		sp := &path.SerialPaths{
			Id:    id,
			Paths: make([]path.PointPath, 0),
		}
		spaths = append(spaths, sp)
		var sx shape.Coord
		var sy shape.Coord
		var err error
		sx, sy, err = blockInfo.GetRandomXY(rect, 1000)
		if err != nil {
			panic("没有找到可以出生的点")
		}
		for i := 0; i < 20; i++ {
			arange2 := 5 //每条路径5个格子，这样不会走很远
			rect2 := shape.Rect{
				X:      int64(int(sx) - arange2),
				Y:      int64(int(sy) - arange2),
				Width:  int64(arange2 * 2),
				Height: int64(arange2 * 2),
			}
			if rect2.X < rect.X {
				rect2.X = rect.X
			}
			if rect2.Y < rect.Y {
				rect2.Y = rect.Y
			}
			if rect2.X+rect2.Width > rect.X+rect.Width {
				rect2.X = rect.X + rect.Width - rect2.Width
			}
			if rect2.Y+rect2.Height > rect.Y+rect.Height {
				rect2.Y = rect.Y + rect.Height - rect2.Height
			}
			ex, ey, err := blockInfo.GetRandomXY(rect2, 100)
			if err != nil {
				fmt.Printf("%v范围找不到可以随机的点", rect2)
				continue
			}
			if ex == sx && ey == sy {
				ex, ey, err = blockInfo.GetRandomXY(rect2, 100)
				if err != nil {
					panic("没有找到可以结束的点")
				}
			}

			tmpPaths, _, _, err := blockInfo.FindPath(int32(sx), int32(sy), int32(ex), int32(ey))
			if err != nil {
				continue
			}
			fmt.Printf("生成路径id:%d,index:%d,sx:%d,sy:%d,ex:%d,ey:%d ,paths::%v \n", id, i, sx, sy, ex, ey, tmpPaths)
			sp.Paths = append(sp.Paths, path.PointPath{
				Sx:    int32(sx),
				Sy:    int32(sy),
				Ex:    int32(ex),
				Ey:    int32(ey),
				Paths: tmpPaths,
			})
			//需要连续的路径
			sx = ex
			sy = ey
		}
	}
	content, err := json.Marshal(spaths)
	if err != nil {
		panic(err)
	}
	err = fileutil.WriteFile(content, fmt.Sprintf("./%s_%s.paths", mapFile, rectstr))
	if err != nil {
		panic(err)
	}
	fmt.Println("结束")
	return nil
}
