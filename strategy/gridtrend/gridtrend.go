package grid

import (
	"strings"

	"github.com/tonkla/autotp/db"
	"github.com/tonkla/autotp/strategy"
	t "github.com/tonkla/autotp/types"
)

type OnTickParams struct {
	Ticker    t.Ticker
	BotParams t.BotParams
	HPrices   []t.HistoricalPrice
	DB        db.DB
}

const openGaps = 2

func OnTick(params OnTickParams) *t.TradeOrders {
	ticker := params.Ticker
	p := params.BotParams
	db := params.DB

	var orders []t.Order

	buyPrice, sellPrice, gridWidth := strategy.GetGridRange(ticker.Price, p.LowerPrice, p.UpperPrice, p.Grids)
	trend := strategy.GetTrend(params.HPrices, int(p.MAPeriod))

	order := t.Order{
		BotID:    p.BotID,
		Exchange: ticker.Exchange,
		Symbol:   ticker.Symbol,
		Qty:      p.Qty,
		Status:   t.ORDER_STATUS_LIMIT,
	}

	view := strings.ToUpper(p.View)

	if view == t.VIEW_NEUTRAL || view == "N" || view == t.VIEW_LONG || view == "L" {
		order.OpenPrice = buyPrice
		if trend <= t.TREND_DOWN_1 {
			order.OpenPrice = buyPrice - gridWidth*openGaps
		}
		order.Side = t.SIDE_BUY
		if db.GetActiveOrder(order, p.Slippage) == nil {
			if p.SL > 0 {
				order.SL = buyPrice - gridWidth*p.SL
			}
			if p.TP > 0 {
				order.TP = buyPrice + gridWidth*p.TP
			}
		}
		orders = append(orders, order)
	}

	if view == t.VIEW_NEUTRAL || view == "N" || view == t.VIEW_SHORT || view == "S" {
		order.OpenPrice = sellPrice
		if trend >= t.TREND_UP_1 {
			order.OpenPrice = sellPrice + gridWidth*openGaps
		}
		order.Side = t.SIDE_SELL
		if db.GetActiveOrder(order, p.Slippage) == nil {
			if p.SL > 0 {
				order.SL = sellPrice + gridWidth*p.SL
			}
			if p.TP > 0 {
				order.TP = sellPrice - gridWidth*p.TP
			}
			orders = append(orders, order)
		}
	}

	return &t.TradeOrders{
		OpenOrders: orders,
	}
}
