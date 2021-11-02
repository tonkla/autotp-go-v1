package common

import (
	"math"

	h "github.com/tonkla/autotp/helper"
	"github.com/tonkla/autotp/rdb"
	"github.com/tonkla/autotp/talib"
	t "github.com/tonkla/autotp/types"
)

// IsUpperMA returns true if the price is upper than the current WMA price
func IsUpperMA(price float64, prices []t.HistoricalPrice, period int64, gap float64) bool {
	cwma := talib.WMA(GetCloses(prices), int(period))
	if len(cwma) == 0 {
		return false
	}
	return price-gap > cwma[len(cwma)-1]
}

// IsLowerMA returns true if the price is lower than the current WMA price
func IsLowerMA(price float64, prices []t.HistoricalPrice, period int64, gap float64) bool {
	cwma := talib.WMA(GetCloses(prices), int(period))
	if len(cwma) == 0 {
		return false
	}
	return price+gap < cwma[len(cwma)-1]
}

// GetTrend returns a stupid trend, do not trust him
func GetTrend(hprices []t.HistoricalPrice, period int) int {
	trend := t.TrendNo

	if len(hprices) < 10 || hprices[len(hprices)-1].Open == 0 || period <= 0 {
		return trend
	}

	p_0 := hprices[len(hprices)-1]
	p_1 := hprices[len(hprices)-2]

	o_0 := p_0.Open

	h_0 := p_0.High
	h_1 := p_1.High

	l_0 := p_0.Low
	l_1 := p_1.Low

	c_0 := p_0.Close

	var h, l, c []float64
	for _, p := range hprices {
		h = append(h, p.High)
		l = append(l, p.Low)
		c = append(c, p.Close)
	}

	hwma := talib.WMA(h, period)
	hma_0 := hwma[len(hwma)-1]

	lwma := talib.WMA(l, period)
	lma_0 := lwma[len(lwma)-1]

	cwma := talib.WMA(c, period)
	cma_0 := cwma[len(cwma)-1]
	cma_1 := cwma[len(cwma)-2]
	cma_2 := cwma[len(cwma)-3]

	// Not the J. Welles Wilder Jr.'s ATR
	atr := hma_0 - lma_0

	// Positive slope
	if cma_1 < cma_0 {
		trend = t.TrendUp1
		// Higher low, and continued positive slope
		if l_1 < l_0 && cma_2 < cma_1 {
			trend = t.TrendUp2
			// Green bar, or moving to top
			if o_0 < c_0 || h_0-c_0 < (c_0-l_0)*0.5 {
				trend = t.TrendUp3
				// Low is greater than average close, or long green bar, or narrow upper band
				if l_0 > cma_0 || h_0-l_0 > atr || hma_0-cma_0 < (cma_0-lma_0)*0.6 {
					trend = t.TrendUp4
					// Low is greater than average high, or very long green bar
					if l_0 > hma_0 || h_0-l_0 > 1.25*atr {
						trend = t.TrendUp5
					}
				}
			}
		}
	}
	// Negative slope
	if cma_1 > cma_0 {
		trend = t.TrendDown1
		// Lower high, and continued negative slope
		if h_1 > h_0 && cma_2 > cma_1 {
			trend = t.TrendDown2
			// Red bar, or moving to bottom
			if o_0 > c_0 || (h_0-c_0)*0.5 > c_0-l_0 {
				trend = t.TrendDown3
				// High is less than average close, or long red bar, or narrow lower band
				if h_0 < cma_0 || h_0-l_0 > atr || (hma_0-cma_0)*0.6 > cma_0-lma_0 {
					trend = t.TrendDown4
					// High is less than average low, or very long red bar
					if h_0 < lma_0 || h_0-l_0 > 1.25*atr {
						trend = t.TrendDown5
					}
				}
			}
		}
	}
	return trend
}

// GetATR returns @tonkla's ATR, that is not the "J. Welles Wilder Jr."'s ATR :-P
func GetATR(prices []t.HistoricalPrice, period int) *float64 {
	if len(prices) == 0 || len(prices) <= period {
		return nil
	}

	var h, l = GetHighsLows(prices)
	hwma := talib.WMA(h, period)
	hma_0 := hwma[len(hwma)-1]
	lwma := talib.WMA(l, period)
	lma_0 := lwma[len(lwma)-1]
	atr := hma_0 - lma_0
	return &atr
}

// GetGridRange returns the lower number and the upper number that closed to the target number
func GetGridRange(target float64, lowerNum float64, upperNum float64, gridSize float64) (float64, float64, float64) {
	if target <= lowerNum || lowerNum >= upperNum || gridSize < 2 {
		return 0, 0, 0
	}

	lower := lowerNum
	upper := upperNum
	zone := upper - lower
	gridWidth := zone / gridSize

	if math.Mod(gridSize, 5) == 0 {
		div := zone / 5

		if lower+div*4 < target {
			lower += div * 4
		} else if lower+div*3 < target {
			lower += div * 3
		} else if lower+div*2 < target {
			lower += div * 2
		} else if lower+div < target {
			lower += div
		}

		if upper-div*4 > target {
			upper -= div * 4
		} else if upper-div*3 > target {
			upper -= div * 3
		} else if upper-div*2 > target {
			upper -= div * 2
		} else if upper-div > target {
			upper -= div
		}
	} else if math.Mod(gridSize, 4) == 0 {
		div := zone / 4

		if lower+div*3 < target {
			lower += div * 3
		} else if lower+div*2 < target {
			lower += div * 2
		} else if lower+div < target {
			lower += div
		}

		if upper-div*3 > target {
			upper -= div * 3
		} else if upper-div*2 > target {
			upper -= div * 2
		} else if upper-div > target {
			upper -= div
		}
	} else if math.Mod(gridSize, 3) == 0 {
		div := zone / 3

		if lower+div*2 < target {
			lower += div * 2
		} else if lower+div < target {
			lower += div
		}

		if upper-div*2 > target {
			upper -= div * 2
		} else if upper-div > target {
			upper -= div
		}
	} else if math.Mod(gridSize, 2) == 0 {
		div := zone / 2

		if lower+div < target {
			lower += div
		}

		if upper-div > target {
			upper -= div
		}
	}

	for i := 0; i < int(gridSize); i++ {
		if lower+gridWidth < target {
			lower += gridWidth
		} else {
			break
		}
	}

	for i := 0; i < int(gridSize); i++ {
		if upper-gridWidth > target {
			upper -= gridWidth
		} else {
			break
		}
	}

	return lower, upper, gridWidth
}

// GetGridZones returns all buyable zones of the grid
func GetGridZones(target float64, lowerNum float64, upperNum float64, gridSize float64) ([]float64, float64) {
	if target <= lowerNum || lowerNum >= upperNum || gridSize < 2 {
		return nil, 0
	}

	start, _, gridWidth := GetGridRange(target, lowerNum, upperNum, gridSize)

	var zones []float64
	for i := 0.0; i < gridSize; i++ {
		num := start + i*gridWidth
		if num >= upperNum {
			break
		}
		zones = append(zones, num)
	}
	return zones, gridWidth
}

// GetCloses returns CLOSE prices of the historical prices
func GetCloses(prices []t.HistoricalPrice) []float64 {
	var c []float64
	for _, p := range prices {
		c = append(c, p.Close)
	}
	return c
}

// GetHighsLows returns HIGH,LOW prices of the historical prices
func GetHighsLows(prices []t.HistoricalPrice) ([]float64, []float64) {
	var h, l []float64
	for _, p := range prices {
		h = append(h, p.High)
		l = append(l, p.Low)
	}
	return h, l
}

// GetHighsLowsCloses returns HIGH,LOW,CLOSE prices of the historical prices
func GetHighsLowsCloses(prices []t.HistoricalPrice) ([]float64, []float64, []float64) {
	var h, l, c []float64
	for _, p := range prices {
		h = append(h, p.High)
		l = append(l, p.Low)
		c = append(l, p.Close)
	}
	return h, l, c
}

// SLLong creates SL orders of active LONG orders
func SLLong(db *rdb.DB, bp *t.BotParams, qo t.QueryOrder, ticker t.Ticker, atr float64) []t.Order {
	var closeOrders []t.Order

	slStop := float64(bp.SLim.SLStop)
	isFutures := bp.Product == t.ProductFutures

	for _, o := range db.GetFilledLimitLongOrders(qo) {
		if db.GetSLOrder(o.ID) != nil {
			continue
		}

		slPrice := 0.0
		if bp.QuoteSL > 0 {
			// SL by a value of the quote currency
			slPrice = o.OpenPrice - bp.QuoteSL/o.Qty
		} else if bp.AtrSL > 0 {
			// SL by a volatility
			slPrice = o.OpenPrice - bp.AtrSL*atr
		}

		if slPrice <= 0 {
			continue
		}

		slPrice = h.NormalizeDouble(slPrice, bp.PriceDigits)
		stopPrice := h.CalcSLStop(o.Side, slPrice, slStop, bp.PriceDigits)
		if ticker.Price-((stopPrice-slPrice)*2) < stopPrice {
			slo := t.Order{
				ID:          h.GenID(),
				BotID:       bp.BotID,
				Exchange:    bp.Exchange,
				Symbol:      bp.Symbol,
				Side:        t.OrderSideSell,
				Type:        t.OrderTypeSL,
				Status:      t.OrderStatusNew,
				Qty:         h.NormalizeDouble(o.Qty, bp.QtyDigits),
				StopPrice:   stopPrice,
				OpenPrice:   slPrice,
				OpenOrderID: o.ID,
			}
			if isFutures {
				slo.Type = t.OrderTypeFSL
				slo.PosSide = o.PosSide
			}
			closeOrders = append(closeOrders, slo)
		}
	}

	return closeOrders
}

// SLShort creates SL orders of active SHORT orders
func SLShort(db *rdb.DB, bp *t.BotParams, qo t.QueryOrder, ticker t.Ticker, atr float64) []t.Order {
	var closeOrders []t.Order

	slStop := float64(bp.SLim.SLStop)
	isFutures := bp.Product == t.ProductFutures

	for _, o := range db.GetFilledLimitShortOrders(qo) {
		if db.GetSLOrder(o.ID) != nil {
			continue
		}

		slPrice := 0.0
		if bp.QuoteSL > 0 {
			// SL by a value of the quote currency
			slPrice = o.OpenPrice + bp.QuoteSL/o.Qty
		} else if bp.AtrSL > 0 {
			// SL by a volatility
			slPrice = o.OpenPrice + bp.AtrSL*atr
		}

		if slPrice <= 0 {
			continue
		}

		slPrice = h.NormalizeDouble(slPrice, bp.PriceDigits)
		stopPrice := h.CalcSLStop(o.Side, slPrice, slStop, bp.PriceDigits)
		if ticker.Price+((slPrice-stopPrice)*2) > stopPrice {
			slo := t.Order{
				ID:          h.GenID(),
				BotID:       bp.BotID,
				Exchange:    qo.Exchange,
				Symbol:      qo.Symbol,
				Side:        t.OrderSideBuy,
				Type:        t.OrderTypeSL,
				Status:      t.OrderStatusNew,
				Qty:         h.NormalizeDouble(o.Qty, bp.QtyDigits),
				StopPrice:   stopPrice,
				OpenPrice:   slPrice,
				OpenOrderID: o.ID,
			}
			if isFutures {
				slo.Type = t.OrderTypeFSL
				slo.PosSide = o.PosSide
			}
			closeOrders = append(closeOrders, slo)
		}
	}

	return closeOrders
}

// TPLong creates TP orders of active LONG orders
func TPLong(db *rdb.DB, bp *t.BotParams, qo t.QueryOrder, ticker t.Ticker, atr float64) []t.Order {
	var closeOrders []t.Order

	tpStop := float64(bp.SLim.TPStop)
	isFutures := bp.Product == t.ProductFutures

	for _, o := range db.GetFilledLimitLongOrders(qo) {
		if db.GetTPOrder(o.ID) != nil {
			continue
		}

		tpPrice := 0.0
		if bp.QuoteTP > 0 {
			// TP by a value of the quote currency
			tpPrice = o.OpenPrice + bp.QuoteTP/o.Qty
		} else if bp.AtrTP > 0 {
			// TP by a volatility
			tpPrice = o.OpenPrice + bp.AtrTP*atr
		}

		if tpPrice <= 0 {
			continue
		}

		tpPrice = h.NormalizeDouble(tpPrice, bp.PriceDigits)
		stopPrice := h.CalcTPStop(o.Side, tpPrice, tpStop, bp.PriceDigits)
		if ticker.Price+((tpPrice-stopPrice)*2) > stopPrice {
			tpo := t.Order{
				ID:          h.GenID(),
				BotID:       bp.BotID,
				Exchange:    qo.Exchange,
				Symbol:      qo.Symbol,
				Side:        t.OrderSideSell,
				Type:        t.OrderTypeTP,
				Status:      t.OrderStatusNew,
				Qty:         h.NormalizeDouble(o.Qty, bp.QtyDigits),
				StopPrice:   stopPrice,
				OpenPrice:   tpPrice,
				OpenOrderID: o.ID,
			}
			if isFutures {
				tpo.Type = t.OrderTypeFTP
				tpo.PosSide = o.PosSide
			}
			closeOrders = append(closeOrders, tpo)
		}
	}

	return closeOrders
}

// TPShort creates TP orders of active SHORT orders
func TPShort(db *rdb.DB, bp *t.BotParams, qo t.QueryOrder, ticker t.Ticker, atr float64) []t.Order {
	var closeOrders []t.Order

	tpStop := float64(bp.SLim.TPStop)
	isFutures := bp.Product == t.ProductFutures

	for _, o := range db.GetFilledLimitShortOrders(qo) {
		if db.GetTPOrder(o.ID) != nil {
			continue
		}

		tpPrice := 0.0
		if bp.QuoteTP > 0 {
			// TP by a value of the quote currency
			tpPrice = o.OpenPrice - bp.QuoteTP/o.Qty
		} else if bp.AtrTP > 0 {
			// TP by a volatility
			tpPrice = o.OpenPrice - bp.AtrTP*atr
		}

		if tpPrice <= 0 {
			continue
		}

		tpPrice = h.NormalizeDouble(tpPrice, bp.PriceDigits)
		stopPrice := h.CalcTPStop(o.Side, tpPrice, tpStop, bp.PriceDigits)
		if ticker.Price-((stopPrice-tpPrice)*2) < stopPrice {
			tpo := t.Order{
				ID:          h.GenID(),
				BotID:       bp.BotID,
				Exchange:    qo.Exchange,
				Symbol:      qo.Symbol,
				Side:        t.OrderSideBuy,
				Type:        t.OrderTypeTP,
				Status:      t.OrderStatusNew,
				Qty:         h.NormalizeDouble(o.Qty, bp.QtyDigits),
				StopPrice:   stopPrice,
				OpenPrice:   tpPrice,
				OpenOrderID: o.ID,
			}
			if isFutures {
				tpo.Type = t.OrderTypeFTP
				tpo.PosSide = o.PosSide
			}
			closeOrders = append(closeOrders, tpo)
		}
	}

	return closeOrders
}

// CloseLong closes all LONG orders at the ticker price
func CloseLong(db *rdb.DB, bp *t.BotParams, qo t.QueryOrder, ticker t.Ticker) []t.Order {
	var orders []t.Order
	for _, o := range db.GetFilledLimitLongOrders(qo) {
		if ticker.Price > o.OpenPrice {
			order := TPLongNow(db, bp, ticker, o)
			if order != nil {
				orders = append(orders, *order)
			}
		} else if ticker.Price < o.OpenPrice {
			order := SLLongNow(db, bp, ticker, o)
			if order != nil {
				orders = append(orders, *order)
			}
		}
	}
	return orders
}

// CloseShort closes all SHORT orders at the ticker price
func CloseShort(db *rdb.DB, bp *t.BotParams, qo t.QueryOrder, ticker t.Ticker) []t.Order {
	var orders []t.Order
	for _, o := range db.GetFilledLimitShortOrders(qo) {
		if ticker.Price < o.OpenPrice {
			order := TPShortNow(db, bp, ticker, o)
			if order != nil {
				orders = append(orders, *order)
			}
		} else if ticker.Price > o.OpenPrice {
			order := SLShortNow(db, bp, ticker, o)
			if order != nil {
				orders = append(orders, *order)
			}
		}
	}
	return orders
}

// SLLongNow creates a SL order of the LONG order from the ticker price
func SLLongNow(db *rdb.DB, bp *t.BotParams, ticker t.Ticker, o t.Order) *t.Order {
	if db.GetSLOrder(o.ID) != nil {
		return nil
	}

	slPrice := h.CalcStopLowerTicker(ticker.Price, 100, bp.PriceDigits)
	stopPrice := h.CalcSLStop(o.Side, slPrice, 50, bp.PriceDigits)
	slo := t.Order{
		ID:          h.GenID(),
		BotID:       bp.BotID,
		Exchange:    bp.Exchange,
		Symbol:      bp.Symbol,
		Side:        t.OrderSideSell,
		Type:        t.OrderTypeSL,
		Status:      t.OrderStatusNew,
		Qty:         h.NormalizeDouble(o.Qty, bp.QtyDigits),
		StopPrice:   stopPrice,
		OpenPrice:   slPrice,
		OpenOrderID: o.ID,
	}
	if bp.Product == t.ProductFutures {
		slo.Type = t.OrderTypeFSL
		slo.PosSide = o.PosSide
	}
	return &slo
}

// SLShortNow creates a SL order of the SHORT order from the ticker price
func SLShortNow(db *rdb.DB, bp *t.BotParams, ticker t.Ticker, o t.Order) *t.Order {
	if db.GetSLOrder(o.ID) != nil {
		return nil
	}

	slPrice := h.CalcStopUpperTicker(ticker.Price, 100, bp.PriceDigits)
	stopPrice := h.CalcSLStop(o.Side, slPrice, 50, bp.PriceDigits)
	slo := t.Order{
		ID:          h.GenID(),
		BotID:       bp.BotID,
		Exchange:    bp.Exchange,
		Symbol:      bp.Symbol,
		Side:        t.OrderSideBuy,
		Type:        t.OrderTypeSL,
		Status:      t.OrderStatusNew,
		Qty:         h.NormalizeDouble(o.Qty, bp.QtyDigits),
		StopPrice:   stopPrice,
		OpenPrice:   slPrice,
		OpenOrderID: o.ID,
	}
	if bp.Product == t.ProductFutures {
		slo.Type = t.OrderTypeFSL
		slo.PosSide = o.PosSide
	}
	return &slo
}

// TPLongNow creates a TP order of the LONG order from the ticker price
func TPLongNow(db *rdb.DB, bp *t.BotParams, ticker t.Ticker, o t.Order) *t.Order {
	if db.GetTPOrder(o.ID) != nil {
		return nil
	}

	tpPrice := h.CalcStopUpperTicker(ticker.Price, 100, bp.PriceDigits)
	stopPrice := h.CalcTPStop(o.Side, tpPrice, 50, bp.PriceDigits)
	tpo := t.Order{
		ID:          h.GenID(),
		BotID:       bp.BotID,
		Exchange:    bp.Exchange,
		Symbol:      bp.Symbol,
		Side:        t.OrderSideSell,
		Type:        t.OrderTypeTP,
		Status:      t.OrderStatusNew,
		Qty:         h.NormalizeDouble(o.Qty, bp.QtyDigits),
		StopPrice:   stopPrice,
		OpenPrice:   tpPrice,
		OpenOrderID: o.ID,
	}
	if bp.Product == t.ProductFutures {
		tpo.Type = t.OrderTypeFTP
		tpo.PosSide = o.PosSide
	}
	return &tpo
}

// TPShortNow creates a TP order of the SHORT order from the ticker price
func TPShortNow(db *rdb.DB, bp *t.BotParams, ticker t.Ticker, o t.Order) *t.Order {
	if db.GetTPOrder(o.ID) != nil {
		return nil
	}

	tpPrice := h.CalcStopLowerTicker(ticker.Price, 100, bp.PriceDigits)
	stopPrice := h.CalcTPStop(o.Side, tpPrice, 50, bp.PriceDigits)
	tpo := t.Order{
		ID:          h.GenID(),
		BotID:       bp.BotID,
		Exchange:    bp.Exchange,
		Symbol:      bp.Symbol,
		Side:        t.OrderSideBuy,
		Type:        t.OrderTypeTP,
		Status:      t.OrderStatusNew,
		Qty:         h.NormalizeDouble(o.Qty, bp.QtyDigits),
		StopPrice:   stopPrice,
		OpenPrice:   tpPrice,
		OpenOrderID: o.ID,
	}
	if bp.Product == t.ProductFutures {
		tpo.Type = t.OrderTypeFTP
		tpo.PosSide = o.PosSide
	}
	return &tpo
}
