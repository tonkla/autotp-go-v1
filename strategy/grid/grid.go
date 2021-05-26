package grid

import (
	"strings"

	"github.com/tonkla/autotp/helper"
	"github.com/tonkla/autotp/types"
)

func OnTick(ticker *types.Ticker, p types.GridParams, h types.Helper) []types.Order {
	buyPrice, sellPrice, gridWidth := helper.GetGridRange(ticker.Price, p.LowerPrice, p.UpperPrice, float64(p.Grids))

	view := strings.ToLower(p.View)

	order := types.Order{
		Exchange: ticker.Exchange,
		Symbol:   ticker.Symbol,
		Qty:      p.Qty,
		Status:   types.ORDER_STATUS_LIMIT,
	}

	var orders []types.Order

	// Has already bought at this price?
	if view == "long" || view == "l" || view == "neutral" || view == "n" {
		order.Price = buyPrice
		order.Side = types.SIDE_BUY
		if !h.DoesOrderExist(&order) {
			if p.SL > 0 {
				order.SL = buyPrice - gridWidth*p.SL
			}
			if p.TP > 0 {
				order.TP = buyPrice + gridWidth*p.TP
			}
			orders = append(orders, order)
		}
	}

	// Has already sold at this price?
	if view == "short" || view == "s" || view == "neutral" || view == "n" {
		order.Price = sellPrice
		order.Side = types.SIDE_SELL
		if !h.DoesOrderExist(&order) {
			if p.SL > 0 {
				order.SL = sellPrice + gridWidth*p.SL
			}
			if p.TP > 0 {
				order.TP = sellPrice - gridWidth*p.TP
			}
			orders = append(orders, order)
		}
	}

	return orders
}
