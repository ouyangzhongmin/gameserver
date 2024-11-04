package api

import (
	"github.com/gorilla/mux"
	"github.com/lonng/nex"
	"github.com/ouyangzhongmin/gameserver/db"
	"github.com/ouyangzhongmin/gameserver/db/model"
	"github.com/ouyangzhongmin/gameserver/pkg/errutil"
	"github.com/ouyangzhongmin/gameserver/pkg/whitelist"
	"net/http"
	"strconv"
)

func MakeSceneService() http.Handler {
	router := mux.NewRouter()
	router.Handle("/v1/scene/", nex.Handler(sceneList)).Methods("GET")     //获取列表(lite)
	router.Handle("/v1/scene/{id}", nex.Handler(sceneByID)).Methods("GET") //获取记录
	return router
}

func SceneByID(id int) (*model.Scene, error) {
	p, err := db.QueryScene(id)
	if err != nil {
		return nil, err
	}
	return p, nil

}

func sceneList(r *http.Request) ([]model.Scene, error) {
	if !whitelist.VerifyIP(r.RemoteAddr) {
		return nil, errutil.ErrPermissionDenied
	}
	list, err := db.SceneList()
	if err != nil {
		return nil, err
	}
	return list, nil
}

func sceneByID(r *http.Request) (*model.Scene, error) {
	if !whitelist.VerifyIP(r.RemoteAddr) {
		return nil, errutil.ErrPermissionDenied
	}
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok || idStr == "" {
		return nil, errutil.ErrInvalidParameter
	}

	id, err := strconv.ParseInt(idStr, 10, 0)
	if err != nil {
		return nil, errutil.ErrInvalidParameter
	}

	scene, err := SceneByID(int(id))
	if err != nil {
		return nil, err
	}
	return scene, nil
}
