package master

import (
	"errors"
	"fmt"
	"github.com/ouyangzhongmin/gameserver/db/model"
	"github.com/ouyangzhongmin/gameserver/pkg/shape"
	"github.com/ouyangzhongmin/gameserver/protocol"
	"github.com/ouyangzhongmin/nano"
	"github.com/ouyangzhongmin/nano/component"
	"github.com/ouyangzhongmin/nano/session"
)

type sceneCell struct {
	SceneId int
	Cells   []*protocol.Cell
	EnterX  int
	EnterY  int
	Width   int
	Height  int
}

type CellManager struct {
	component.Base
	scenesCells map[int]*sceneCell
}

func NewCellManager() *CellManager {
	return &CellManager{scenesCells: make(map[int]*sceneCell)}
}

func (m *CellManager) AfterInit() {

}

func (m *CellManager) RegisterSceneCell(s *session.Session, req *protocol.RegisterSceneCellRequest) error {
	sceneId := req.SceneId
	//停止的时候要考虑怎么清理掉
	scell, ok := m.scenesCells[sceneId]
	bounds := shape.Rect{0, 0, int64(req.Width), int64(req.Height)}
	if !ok {
		scell = &sceneCell{
			SceneId: sceneId,
			Cells:   make([]*protocol.Cell, 0),
			EnterX:  req.Enterx,
			EnterY:  req.EnterY,
			Width:   req.Width,
			Height:  req.Height,
		}
		m.scenesCells[sceneId] = scell
	} else {
		m.checkSceneCells(scell, req.RemoteAddr)
		cellsLen := len(scell.Cells)
		if cellsLen > 0 {
			// todo 第一个测试方便
			firstW := 60
			tmp := scell.Cells[0]
			tmp.IsNew = false
			tmp.Bounds = shape.Rect{0, 0, int64(firstW), int64(req.Height)}
			//这里只做竖向一刀切，不做横竖九宫格类似的
			cutW := (req.Width - firstW) / (cellsLen)
			for i := 1; i < cellsLen; i++ {
				//旧的也要变更
				tmp := scell.Cells[i]
				tmp.IsNew = false
				offsetW := firstW + (i-1)*cutW
				tmp.Bounds = shape.Rect{int64(offsetW), 0, int64(cutW), int64(req.Height)}
			}
			bounds = shape.Rect{int64(firstW + (cellsLen-1)*cutW), 0, int64(cutW), int64(req.Height)}
		}
	}
	cellId := sceneId*10000 + len(scell.Cells) + 1
	//newsession, err := nano.NewRpcSession(req.GateAddr)
	//if err != nil {
	//	return err
	//}
	//newsession.Set("cellId", cellId)
	//newsession.Set("sceneId", sceneId)
	//newsession.Set("remoteAddr", req.RemoteAddr)
	//newsession.Router().Bind("SceneManager", req.RemoteAddr)
	c := &protocol.Cell{
		CellID:      cellId,
		SceneId:     sceneId,
		Bounds:      bounds,
		EdgeSize:    30,
		RemoteAddr:  req.RemoteAddr,
		GateAddr:    req.GateAddr,
		IsFirstCell: len(scell.Cells) == 0,
		IsNew:       true,
		//Session:     newsession,
	}

	scell.Cells = append(scell.Cells, c)
	for i := 0; i < len(scell.Cells); i++ {
		//通知到scene的具体的cell信息
		tmp := scell.Cells[i]
		logger.Printf("当前场景cell:%d, remoteAddr:%s \n", tmp.CellID, tmp.RemoteAddr)
		err := nano.RPCWithAddr("SceneManager.SceneCells", &protocol.SceneCelllsRequest{
			SceneId: tmp.SceneId,
			CellId:  tmp.CellID,
			Cells:   scell.Cells,
		}, tmp.RemoteAddr)
		if err != nil {
			// todo 这里需要确保能把cell信息通知到
			logger.Errorln("cell.SceneManager.SceneCells err:", err)
			continue
		}
	}
	return nil
}

func (m *CellManager) checkSceneCells(scell *sceneCell, remoteAddr string) {
	for i := len(scell.Cells) - 1; i >= 0; i-- {
		//旧的也要变更
		tmp := scell.Cells[i]
		if tmp.RemoteAddr == "" {
			scell.Cells = append(scell.Cells[:i], scell.Cells[i+1:]...)
			logger.Errorln("cell.Session == nil")
			continue
		}
		err := nano.RPCWithAddr("SceneManager.CheckCellHealth", &protocol.CheckCellHealthRequest{V: 1}, tmp.RemoteAddr)
		if err != nil {
			scell.Cells = append(scell.Cells[:i], scell.Cells[i+1:]...)
			logger.Errorln("cell.SceneManager.CheckHealth err:", err)
			continue
		}
		if remoteAddr == tmp.RemoteAddr {
			scell.Cells = append(scell.Cells[:i], scell.Cells[i+1:]...)
			logger.Errorln("cell 已存在:", remoteAddr)
			continue
		}
	}
}

func (m *CellManager) getSceneCell(hero *model.Hero) *protocol.Cell {
	scell, ok := m.scenesCells[hero.SceneId]
	if !ok {
		return nil
	}
	posx, posy := hero.InitPosx, hero.InitPosy
	if scell.EnterX > 0 && scell.EnterY > 0 {
		posx, posy = scell.EnterX, scell.EnterY
	}
	for _, cell := range scell.Cells {
		if cell.Bounds.Contains(int64(posx), int64(posy)) {
			return cell
		}
	}
	return nil
}

func (m *CellManager) getSceneCellById(sceneId, cellId int) *protocol.Cell {
	scell, ok := m.scenesCells[sceneId]
	if !ok {
		return nil
	}
	for _, cell := range scell.Cells {
		if cell.CellID == cellId {
			return cell
		}
	}
	return nil
}

// 迁移用户数据
func (m *CellManager) MigrateHero(s *session.Session, req *protocol.MigrateHeroRequest) error {
	user, _ := userManager.player(req.HeroData.Uid)
	if user == nil {
		return errors.New(fmt.Sprintf("MigrateHero没有找到用户:%d", req.HeroData.Id))
	}
	if user.heroData.Id != req.HeroData.Id {
		return errors.New(fmt.Sprintf("MigrateHero用户数据错误:%d", req.HeroData.Id))
	}

	// 目标的cell
	cell := m.getSceneCellById(req.SceneId, req.CellId)
	if cell == nil {
		return errors.New("未匹配到任何cell")
	}

	// 更改用户session的route绑定
	// 这里需要统一使用user内记录的session
	userSession := user.session
	userSession.Router().Delete("SceneManager")
	// todo 服务区切换需要更新
	userSession.Set("sceneId", req.SceneId)
	userSession.Set("cellId", cell.CellID) //迁移后需要把请求的标记更改到新的node服务器上
	userSession.Set("remoteAddr", cell.RemoteAddr)
	userSession.Router().Bind("SceneManager", cell.RemoteAddr)

	// 向目标服务器创建迁移对象
	logger.Debugf("向cell:%d创建migrate Hero:%d \n", cell.CellID, req.HeroData.Id)
	// 必须使用userSession
	err := userSession.RPC("SceneManager.CreateMigrateHero", req)
	if err != nil {
		logger.Errorf("rpc.Call(SceneManager.CreateMigrateHero) err: %v \n", err)
		return err
	}

	//通知gate服务器也更改 必须使用userSession
	err = userSession.RPC("GateService.RecordScene", &protocol.UserSceneId{
		Uid:        user.Uid,
		SceneId:    req.SceneId,
		CellId:     cell.CellID,
		RemoteAddr: cell.RemoteAddr,
	})
	if err != nil {
		logger.Errorf("rpc.Call(GateService.RecordScene) err: %v \n", err)
		return err
	}
	return nil
}

func (m *CellManager) updateCellWithMemberShutdown(remoteAddr string) {
	logger.Println("cell节点已断开:", remoteAddr)
	for _, scell := range m.scenesCells {
		for i := len(scell.Cells) - 1; i >= 0; i-- {
			//旧的也要变更
			tmp := scell.Cells[i]
			if tmp.RemoteAddr == "" {
				scell.Cells = append(scell.Cells[:i], scell.Cells[i+1:]...)
				logger.Errorln("cell.RemoteAddr is empty")
				continue
			}
			err := nano.RPCWithAddr("SceneManager.CheckHealth", &protocol.CheckCellHealthRequest{V: 1}, tmp.RemoteAddr)
			if err != nil {
				scell.Cells = append(scell.Cells[:i], scell.Cells[i+1:]...)
				logger.Errorln("cell.SceneManager.CheckHealth err:", err)
				continue
			}
			if remoteAddr == tmp.RemoteAddr {
				scell.Cells = append(scell.Cells[:i], scell.Cells[i+1:]...)
				logger.Errorln("cell 节点已关闭:", remoteAddr)
				continue
			}
		}
	}

}
