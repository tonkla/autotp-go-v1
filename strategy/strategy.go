package strategy

import (
	"errors"

	"github.com/tonkla/autotp/exchange"
	"github.com/tonkla/autotp/rdb"
	"github.com/tonkla/autotp/strategy/daily"
	"github.com/tonkla/autotp/strategy/grid"
	"github.com/tonkla/autotp/strategy/trend"
	t "github.com/tonkla/autotp/types"
)

type Repository interface {
	OnTick(t.Ticker) *t.TradeOrders
}

func New(db *rdb.DB, bp *t.BotParams, ex exchange.Repository) (Repository, error) {
	if bp.Strategy == t.StrategyGrid {
		return grid.New(db, bp, ex), nil
	} else if bp.Strategy == t.StrategyTrend {
		return trend.New(db, bp), nil
	} else if bp.Strategy == t.StrategyDaily {
		return daily.New(db, bp), nil
	}
	return nil, errors.New("strategy not found")
}
