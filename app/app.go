package app

import (
	"github.com/tonkla/autotp/exchange"
	"github.com/tonkla/autotp/rdb"
	s "github.com/tonkla/autotp/strategy/common"
	"github.com/tonkla/autotp/types"
)

// Note: cannot move this to /types because of circular import of 'db'
type AppParams struct {
	EX exchange.Repository
	ST s.Repository
	DB *rdb.DB
	BP *types.BotParams
	TK *types.Ticker
	TO *types.TradeOrders
	QO types.QueryOrder
}
