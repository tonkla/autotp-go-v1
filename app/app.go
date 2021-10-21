package app

import (
	"github.com/tonkla/autotp/exchange"
	"github.com/tonkla/autotp/rdb"
	"github.com/tonkla/autotp/strategy"
	t "github.com/tonkla/autotp/types"
)

type AppParams struct {
	EX exchange.Repository
	ST strategy.Repository
	DB *rdb.DB
	BP *t.BotParams
	TK *t.Ticker
	TO t.TradeOrders
	QO t.QueryOrder
}
