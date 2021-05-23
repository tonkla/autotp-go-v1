package grid

import (
	"strings"

	"github.com/tonkla/autotp/db"
	"github.com/tonkla/autotp/helper"
	"github.com/tonkla/autotp/types"
)

func OnTick(ticker types.Ticker, p types.GridParams) []types.Order {
	buyPrice, sellPrice, gridWidth := helper.GetGridRange(ticker.Price, p.LowerPrice, p.UpperPrice, float64(p.Grids))

	var orders []types.Order

	_view := strings.ToLower(p.View)

	// Has already bought at this price?
	if _view == "long" || _view == "l" || _view == "neutral" || _view == "n" {
		order := types.Order{
			Exchange: ticker.Exchange,
			Symbol:   ticker.Symbol,
			Price:    buyPrice,
			TP:       buyPrice + gridWidth*2,
			Qty:      p.Qty,
			Side:     types.SIDE_BUY,
			Status:   types.ORDER_STATUS_LIMIT,
		}
		if !db.DoesOrderExists(&order) {
			orders = append(orders, order)
		}
	}

	// Has already sold at this price?
	if _view == "short" || _view == "s" || _view == "neutral" || _view == "n" {
		order := types.Order{
			Exchange: ticker.Exchange,
			Symbol:   ticker.Symbol,
			Price:    sellPrice,
			TP:       sellPrice - gridWidth*2,
			Qty:      p.Qty,
			Side:     types.SIDE_SELL,
			Status:   types.ORDER_STATUS_LIMIT,
		}
		if !db.DoesOrderExists(&order) {
			orders = append(orders, order)
		}
	}

	return orders
}
