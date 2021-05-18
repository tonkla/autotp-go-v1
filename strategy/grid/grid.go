package grid

import (
	"github.com/tonkla/autotp/db"
	"github.com/tonkla/autotp/helper"
	"github.com/tonkla/autotp/types"
)

func OnTick(ticker types.Ticker, p types.GridParams) []types.Order {
	buyPrice, sellPrice, gridWidth := helper.GetGridRange(ticker.Price, p.LowerPrice, p.UpperPrice, float64(p.Grids))

	var orders []types.Order

	// Has already bought at this price?
	record := db.GetRecordByPrice(buyPrice, types.SIDE_BUY)
	if record == nil {
		orders = append(orders, types.Order{
			Symbol: ticker.Symbol,
			Side:   types.SIDE_BUY,
			Price:  buyPrice,
			Qty:    p.Qty,
			TP:     buyPrice + gridWidth*2,
		})
	}

	// Has already sold at this price?
	record = db.GetRecordByPrice(sellPrice, types.SIDE_SELL)
	if record == nil {
		orders = append(orders, types.Order{
			Symbol: ticker.Symbol,
			Side:   types.SIDE_SELL,
			Price:  sellPrice,
			Qty:    p.Qty,
			TP:     sellPrice - gridWidth*2,
		})
	}

	return orders
}
