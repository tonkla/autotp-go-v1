package gridtrend

import (
	"github.com/tonkla/autotp/db"
	"github.com/tonkla/autotp/helper"
	"github.com/tonkla/autotp/talib"
	"github.com/tonkla/autotp/types"
)

func OnTick(ticker types.Ticker, p types.BotParams, hprices []types.HistoricalPrice, db db.DB) []types.Order {
	var orders []types.Order

	bar := hprices[len(hprices)-1]
	if bar.Open == 0 || bar.High == 0 || bar.Low == 0 || bar.Close == 0 {
		return nil
	}

	close := bar.Close

	var highes, lows, closes []float64
	for _, b := range hprices {
		highes = append(highes, b.High)
		lows = append(lows, b.Low)
		closes = append(closes, b.Close)
	}
	hwma := talib.WMA(highes, int(p.MAPeriod))
	hma_0 := hwma[len(hwma)-1]

	lwma := talib.WMA(lows, int(p.MAPeriod))
	lma_0 := lwma[len(lwma)-1]

	cwma := talib.WMA(closes, int(p.MAPeriod))
	cma_0 := cwma[len(cwma)-1]
	cma_1 := cwma[len(cwma)-2]

	atr := hma_0 - lma_0

	buyPrice, sellPrice, gridWidth := helper.GetGridRange(ticker.Price, p.LowerPrice, p.UpperPrice, p.Grids)

	order := types.Order{
		BotID:    p.BotID,
		Exchange: ticker.Exchange,
		Symbol:   ticker.Symbol,
		Qty:      p.Qty,
		Status:   types.ORDER_STATUS_LIMIT,
	}

	// Uptrend or Oversold
	if (cma_1 < cma_0 && close < hma_0+0.5*atr) || close < lma_0-0.5*atr {
		order.OpenPrice = buyPrice
		order.Side = types.SIDE_BUY
		if p.SL > 0 {
			order.SL = buyPrice - gridWidth*p.SL
		}
		if p.TP > 0 {
			order.TP = buyPrice + gridWidth*p.TP
		}
		if !db.IsOrderActive(order, p.Slippage) {
			orders = append(orders, order)
		}
	}

	// Downtrend or Overbought
	if (cma_1 > cma_0 && close > lma_0-0.5*atr) || close > hma_0+0.5*atr {
		order.OpenPrice = sellPrice
		order.Side = types.SIDE_SELL
		if p.SL > 0 {
			order.SL = sellPrice + gridWidth*p.SL
		}
		if p.TP > 0 {
			order.TP = sellPrice - gridWidth*p.TP
		}
		if !db.IsOrderActive(order, p.Slippage) {
			orders = append(orders, order)
		}
	}

	return orders
}
