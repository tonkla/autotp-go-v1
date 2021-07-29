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
	DB        db.DB
}

const triggerBefore = 0.002

func OnTick(params OnTickParams) *t.TradeOrders {
	ticker := params.Ticker
	p := params.BotParams
	db := params.DB

	var orders []t.Order

	buyPrice, sellPrice, gridWidth := strategy.GetGridRange(ticker.Price, p.LowerPrice, p.UpperPrice, p.Grids)

	order := t.Order{
		BotID:    p.BotID,
		Exchange: ticker.Exchange,
		Symbol:   ticker.Symbol,
		Qty:      p.Qty,
		Status:   t.OrderStatusNew,
		Type:     t.OrderTypeLimit,
	}

	view := strings.ToUpper(p.View)

	if view == t.ViewLong || view == "L" || view == t.ViewNeutral || view == "N" {
		order.Side = t.OrderSideBuy
		order.OpenPrice = buyPrice
		o := db.GetActiveOrder(order, p.Slippage)
		if o != nil {
			if p.SL > 0 {
				closePrice := buyPrice - gridWidth*p.SL
				order.SL = closePrice
				order.StopPrice = closePrice + closePrice*triggerBefore
			}
			if p.TP > 0 {
				closePrice := buyPrice + gridWidth*p.TP
				order.TP = closePrice
				order.StopPrice = closePrice - closePrice*triggerBefore
			}
			orders = append(orders, order)
		}
	}

	if view == t.ViewShort || view == "S" || view == t.ViewNeutral || view == "N" {
		order.Side = t.OrderSideSell
		order.OpenPrice = sellPrice
		o := db.GetActiveOrder(order, p.Slippage)
		if o != nil {
			if p.SL > 0 {
				closePrice := sellPrice + gridWidth*p.SL
				order.SL = closePrice
				order.StopPrice = closePrice - closePrice*triggerBefore
			}
			if p.TP > 0 {
				closePrice := sellPrice - gridWidth*p.TP
				order.TP = closePrice
				order.StopPrice = closePrice + closePrice*triggerBefore
			}
			orders = append(orders, order)
		}
	}

	return &t.TradeOrders{
		OpenOrders: orders,
	}
}
