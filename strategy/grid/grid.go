package grid

import (
	"strings"

	"github.com/tonkla/autotp/helper"
	"github.com/tonkla/autotp/types"
)

func OnTick(ticker *types.Ticker, p *types.BotParams) []types.Order {
	var orders []types.Order

	buyPrice, sellPrice, gridWidth := helper.GetGridRange(ticker.Price, p.LowerPrice, p.UpperPrice, p.Grids)

	view := strings.ToUpper(p.View)

	order := types.Order{
		Exchange: ticker.Exchange,
		Symbol:   ticker.Symbol,
		Qty:      p.Qty,
		Status:   types.ORDER_STATUS_LIMIT,
	}

	if view == "LONG" || view == "L" || view == "NEUTRAL" || view == "N" {
		order.Price = buyPrice
		order.Side = types.SIDE_BUY
		if p.SL > 0 {
			order.SL = buyPrice - gridWidth*p.SL
		}
		if p.TP > 0 {
			order.TP = buyPrice + gridWidth*p.TP
		}
		orders = append(orders, order)
	}

	if view == "SHORT" || view == "S" || view == "NEUTRAL" || view == "N" {
		order.Price = sellPrice
		order.Side = types.SIDE_SELL
		if p.SL > 0 {
			order.SL = sellPrice + gridWidth*p.SL
		}
		if p.TP > 0 {
			order.TP = sellPrice - gridWidth*p.TP
		}
		orders = append(orders, order)
	}

	return orders
}
