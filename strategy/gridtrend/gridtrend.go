package grid

import (
	"strings"

	"github.com/tonkla/autotp/db"
	"github.com/tonkla/autotp/strategy"
	t "github.com/tonkla/autotp/types"
)

type OnTickParams struct {
	Ticker    t.Ticker
	OrderBook t.OrderBook
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

	buyPrice, sellPrice, gridWidth := strategy.GetGridRange(ticker.Price, p.LowerPrice, p.UpperPrice, p.GridSize)
	trend := strategy.GetTrend(params.HPrices, int(p.MAPeriod))

	order := t.Order{
		BotID:    p.BotID,
		Exchange: ticker.Exchange,
		Symbol:   ticker.Symbol,
		Qty:      p.Qty,
		Status:   t.OrderStatusNew,
		Type:     t.OrderTypeLimit,
	}

	view := strings.ToUpper(p.View)

	if view == t.ViewNeutral || view == "N" || view == t.ViewLong || view == "L" {
		order.Side = t.OrderSideBuy
		order.OpenPrice = buyPrice
		if trend <= t.TrendDown1 {
			order.OpenPrice = buyPrice - gridWidth*openGaps
		}

		o := db.GetActiveOrder(order, p.Slippage)
		if o == nil {
			if p.GridTP > 0 {
				order.TPPrice = buyPrice + gridWidth*p.GridTP
			}
			orders = append(orders, order)
		}
	}

	if view == t.ViewNeutral || view == "N" || view == t.ViewShort || view == "S" {
		order.Side = t.OrderSideSell
		order.OpenPrice = sellPrice
		if trend >= t.TrendUp1 {
			order.OpenPrice = sellPrice + gridWidth*openGaps
		}

		o := db.GetActiveOrder(order, p.Slippage)
		if o == nil {
			if p.GridTP > 0 {
				order.TPPrice = sellPrice - gridWidth*p.GridTP
			}
			orders = append(orders, order)
		}
	}

	return &t.TradeOrders{
		OpenOrders: orders,
	}
}
