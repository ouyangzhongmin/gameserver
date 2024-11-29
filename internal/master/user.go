package master

import (
	"github.com/ouyangzhongmin/gameserver/db/model"
	"github.com/ouyangzhongmin/nano/session"
)

type User struct {
	session  *session.Session
	data     *model.User
	Uid      int64
	heroData *model.Hero
}
