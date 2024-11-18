package game

import (
	"github.com/lonng/nano/session"
	"github.com/ouyangzhongmin/gameserver/constants"
	"github.com/ouyangzhongmin/gameserver/pkg/errutil"
)

func heroWithSession(s *session.Session) (*Hero, error) {
	p, ok := s.Value(constants.KCurHero).(*Hero)
	if !ok {
		return nil, errutil.ErrHeroNotFound
	}
	return p, nil
}
