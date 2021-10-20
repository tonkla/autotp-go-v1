package app

import (
	"github.com/tonkla/autotp/db"
	"github.com/tonkla/autotp/exchange"
	s "github.com/tonkla/autotp/strategy/common"
	"github.com/tonkla/autotp/types"
)

// Note: cannot move this to /types because of circular import of 'db'
type AppParams struct {
	EX exchange.Repository
	ST s.Repository
	DB *db.DB
	BP *types.BotParams
	TK *types.Ticker
	TO *types.TradeOrders
	QO types.QueryOrder
}
