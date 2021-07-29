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
		order.OpenPrice = buyPrice
		order.Side = t.OrderSideBuy
		_o := db.GetActiveOrder(order, p.Slippage)
		if _o.OpenPrice == 0 {
			if p.SL > 0 {
				order.SL = buyPrice - gridWidth*p.SL
			}
			if p.TP > 0 {
				order.TP = buyPrice + gridWidth*p.TP
			}
			orders = append(orders, order)
		}
	}

	if view == t.ViewShort || view == "S" || view == t.ViewNeutral || view == "N" {
		order.OpenPrice = sellPrice
		order.Side = t.OrderSideSell
		_o := db.GetActiveOrder(order, p.Slippage)
		if _o.OpenPrice == 0 {
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
