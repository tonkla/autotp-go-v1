package strategy

import (
	"github.com/tonkla/autotp/rdb"
	s "github.com/tonkla/autotp/strategy/common"
	"github.com/tonkla/autotp/strategy/grid"
	t "github.com/tonkla/autotp/types"
)

func New(db rdb.DB, bp t.BotParams) (s.Repository, error) {
	if bp.Strategy == t.StrategyGrid {
		return grid.New(db, bp), nil
	}

	return nil, nil
}
