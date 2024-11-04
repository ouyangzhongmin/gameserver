package db

import (
	"encoding/json"
	"github.com/ouyangzhongmin/gameserver/db/model"
	"github.com/ouyangzhongmin/gameserver/pkg/errutil"
	log "github.com/sirupsen/logrus"
)

func InsertOnline(count int, scenes map[int]interface{}, ts int64) {
	jsonstr, err := json.Marshal(scenes)
	if err != nil {
		log.Errorf("统计在线人数失败: %s", err.Error())
		return
	}
	_, err = database.Exec("insert into online(user_count,scenes,time) values (?,?,?)", count, string(jsonstr), ts)
	if err != nil {
		log.Errorf("统计在线人数失败: %s", err.Error())
	}
}

func OnlineStats(begin, end int64) ([]model.Online, error) {
	if begin > end {
		return nil, errutil.ErrIllegalParameter
	}

	list := make([]model.Online, 0)

	return list, database.Where("`time` BETWEEN ? AND ?", begin, end).Find(&list)
}
