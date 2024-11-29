package game

import (
	"github.com/ouyangzhongmin/gameserver/constants"
	"github.com/ouyangzhongmin/gameserver/pkg/errutil"
	"github.com/ouyangzhongmin/nano/session"
)

func heroWithSession(s *session.Session) (*Hero, error) {
	p, ok := s.Value(constants.KCurHero).(*Hero)
	if !ok {
		return nil, errutil.ErrHeroNotFound
	}
	return p, nil
}
