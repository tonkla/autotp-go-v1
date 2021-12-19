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

// GetSimpleATR returns a simple ATR, not the J. Welles Wilder Jr.'s ATR
func GetSimpleATR(prices []t.HistoricalPrice, period int) float64 {
	var h, l = GetHighsLows(prices)
	hwma := talib.WMA(h, period)
	hma_0 := hwma[len(hwma)-1]
	lwma := talib.WMA(l, period)
	lma_0 := lwma[len(lwma)-1]
	return hma_0 - lma_0
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

// GetLows returns LOW prices of the historical prices
func GetLows(prices []t.HistoricalPrice) []float64 {
	var l []float64
	for _, p := range prices {
		l = append(l, p.Low)
	}
	return l
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
		c = append(c, p.Close)
	}
	return h, l, c
}

// GetHLRatio returns a ratio of the ticker price compared to the previous Highest<->Lowest
func GetHLRatio(prices []t.HistoricalPrice, ticker t.Ticker) float64 {
	var h, l float64
	for _, p := range prices {
		if p.High > h {
			h = p.High
		}
		if l == 0 || p.Low < l {
			l = p.Low
		}
	}
	return (ticker.Price - l) / (h - l)
}

// CloseProfitSpot creates STOP orders for profitable SPOT orders at the ticker price
func CloseProfitSpot(db *rdb.DB, bp *t.BotParams, qo t.QueryOrder, ticker t.Ticker) []t.Order {
	var orders []t.Order
	for _, o := range db.GetFilledLimitBuyOrders(qo) {
		if ticker.Price > o.OpenPrice && (h.Now13()-o.UpdateTime)/1000.0 > bp.TimeSecTP {
			order := TPLongNow(db, bp, ticker, o)
			if order != nil {
				orders = append(orders, *order)
			}
		}
	}
	return orders
}

// CloseProfitLong creates STOP orders for profitable LONG orders at the ticker price
func CloseProfitLong(db *rdb.DB, bp *t.BotParams, qo t.QueryOrder, ticker t.Ticker) []t.Order {
	var orders []t.Order
	for _, o := range db.GetFilledLimitLongOrders(qo) {
		if ticker.Price > o.OpenPrice && (h.Now13()-o.UpdateTime)/1000.0 > 600 {
			order := TPLongNow(db, bp, ticker, o)
			if order != nil {
				orders = append(orders, *order)
			}
		}
	}
	return orders
}

// CloseProfitShort creates STOP orders for profitable SHORT orders at the ticker price
func CloseProfitShort(db *rdb.DB, bp *t.BotParams, qo t.QueryOrder, ticker t.Ticker) []t.Order {
	var orders []t.Order
	for _, o := range db.GetFilledLimitShortOrders(qo) {
		if ticker.Price < o.OpenPrice && (h.Now13()-o.UpdateTime)/1000.0 > 600 {
			order := TPShortNow(db, bp, ticker, o)
			if order != nil {
				orders = append(orders, *order)
			}
		}
	}
	return orders
}

// CloseLong creates STOP orders for active LONG orders at the ticker price
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

// CloseShort creates STOP orders for active SHORT orders at the ticker price
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

// CloseOpposite creates STOP orders for the older opposite order
// when LONG/SHORT orders have been openned at the same time
func CloseOpposite(db *rdb.DB, bp *t.BotParams, qo t.QueryOrder, ticker t.Ticker) []t.Order {
	var orders []t.Order
	lorders := db.GetFilledLimitLongOrders(qo)
	sorders := db.GetFilledLimitShortOrders(qo)
	if len(lorders) > 0 && len(sorders) > 0 {
		if lorders[0].UpdateTime < sorders[0].UpdateTime {
			orders = CloseLong(db, bp, qo, ticker)
		} else {
			orders = CloseShort(db, bp, qo, ticker)
		}
	}
	return orders
}

// TimeSL creates SL orders based on the order open time
func TimeSL(db *rdb.DB, bp *t.BotParams, qo t.QueryOrder, ticker t.Ticker) []t.Order {
	if bp.TimeSecSL <= 0 {
		return nil
	}

	var closeOrders []t.Order

	for _, o := range db.GetFilledLimitLongOrders(qo) {
		if db.GetSLOrder(o.ID) != nil {
			continue
		}

		if ticker.Price < o.OpenPrice && (h.Now13()-o.UpdateTime)/1000.0 > bp.TimeSecSL {
			_o := SLLongNow(db, bp, ticker, o)
			if _o != nil {
				closeOrders = append(closeOrders, *_o)
			}
		}
	}

	for _, o := range db.GetFilledLimitShortOrders(qo) {
		if db.GetSLOrder(o.ID) != nil {
			continue
		}

		if ticker.Price > o.OpenPrice && (h.Now13()-o.UpdateTime)/1000.0 > bp.TimeSecSL {
			_o := SLShortNow(db, bp, ticker, o)
			if _o != nil {
				closeOrders = append(closeOrders, *_o)
			}
		}
	}

	return closeOrders
}

// TimeTP creates TP orders based on the order open time
func TimeTP(db *rdb.DB, bp *t.BotParams, qo t.QueryOrder, ticker t.Ticker) []t.Order {
	if bp.TimeSecTP <= 0 {
		return nil
	}

	var closeOrders []t.Order

	for _, o := range db.GetFilledLimitLongOrders(qo) {
		if db.GetTPOrder(o.ID) != nil {
			continue
		}

		if ticker.Price > o.OpenPrice && (h.Now13()-o.UpdateTime)/1000.0 > bp.TimeSecTP {
			_o := TPLongNow(db, bp, ticker, o)
			if _o != nil {
				closeOrders = append(closeOrders, *_o)
			}
		}
	}

	for _, o := range db.GetFilledLimitShortOrders(qo) {
		if db.GetTPOrder(o.ID) != nil {
			continue
		}

		if ticker.Price < o.OpenPrice && (h.Now13()-o.UpdateTime)/1000.0 > bp.TimeSecTP {
			_o := TPShortNow(db, bp, ticker, o)
			if _o != nil {
				closeOrders = append(closeOrders, *_o)
			}
		}
	}

	return closeOrders
}

// SLLong creates SL orders of active LONG orders
func SLLong(db *rdb.DB, bp *t.BotParams, qo t.QueryOrder, ticker t.Ticker, atr float64) []t.Order {
	if bp.QuoteSL <= 0 && bp.AtrSL <= 0 {
		return nil
	}

	var closeOrders []t.Order

	for _, o := range db.GetFilledLimitLongOrders(qo) {
		if db.GetSLOrder(o.ID) != nil {
			continue
		}

		slPrice := 0.0
		if bp.QuoteSL > 0 {
			// SL by a value of the quote currency
			slPrice = o.OpenPrice - bp.QuoteSL/o.Qty
		} else if bp.AtrSL > 0 && atr > 0 {
			// SL by a volatility
			slPrice = o.OpenPrice - bp.AtrSL*atr
		}

		if slPrice <= 0 {
			continue
		}

		stopPrice := h.CalcSLStop(o.Side, slPrice, float64(bp.Gap.SLStop), bp.PriceDigits)
		if ticker.Price-(stopPrice-slPrice) < stopPrice {
			slPrice = h.CalcStopLowerTicker(ticker.Price, float64(bp.Gap.SLLimit), bp.PriceDigits)
			stopPrice = h.CalcStopLowerTicker(ticker.Price, float64(bp.Gap.SLStop), bp.PriceDigits)
			slo := t.Order{
				ID:          h.GenID(),
				BotID:       bp.BotID,
				Exchange:    bp.Exchange,
				Symbol:      bp.Symbol,
				Side:        t.OrderSideSell,
				PosSide:     t.OrderPosSideLong,
				Type:        t.OrderTypeFSL,
				Status:      t.OrderStatusNew,
				Qty:         h.NormalizeDouble(o.Qty, bp.QtyDigits),
				StopPrice:   stopPrice,
				OpenPrice:   slPrice,
				OpenOrderID: o.ID,
			}
			closeOrders = append(closeOrders, slo)
		}
	}

	return closeOrders
}

// SLShort creates SL orders of active SHORT orders
func SLShort(db *rdb.DB, bp *t.BotParams, qo t.QueryOrder, ticker t.Ticker, atr float64) []t.Order {
	if bp.QuoteSL <= 0 && bp.AtrSL <= 0 {
		return nil
	}

	var closeOrders []t.Order

	for _, o := range db.GetFilledLimitShortOrders(qo) {
		if db.GetSLOrder(o.ID) != nil {
			continue
		}

		slPrice := 0.0
		if bp.QuoteSL > 0 {
			// SL by a value of the quote currency
			slPrice = o.OpenPrice + bp.QuoteSL/o.Qty
		} else if bp.AtrSL > 0 && atr > 0 {
			// SL by a volatility
			slPrice = o.OpenPrice + bp.AtrSL*atr
		}

		if slPrice <= 0 {
			continue
		}

		stopPrice := h.CalcSLStop(o.Side, slPrice, float64(bp.Gap.SLStop), bp.PriceDigits)
		if ticker.Price+(slPrice-stopPrice) > stopPrice {
			slPrice = h.CalcStopUpperTicker(ticker.Price, float64(bp.Gap.SLLimit), bp.PriceDigits)
			stopPrice = h.CalcStopUpperTicker(ticker.Price, float64(bp.Gap.SLStop), bp.PriceDigits)
			slo := t.Order{
				ID:          h.GenID(),
				BotID:       bp.BotID,
				Exchange:    qo.Exchange,
				Symbol:      qo.Symbol,
				Side:        t.OrderSideBuy,
				PosSide:     t.OrderPosSideShort,
				Type:        t.OrderTypeFSL,
				Status:      t.OrderStatusNew,
				Qty:         h.NormalizeDouble(o.Qty, bp.QtyDigits),
				StopPrice:   stopPrice,
				OpenPrice:   slPrice,
				OpenOrderID: o.ID,
			}
			closeOrders = append(closeOrders, slo)
		}
	}

	return closeOrders
}

// TPSpot creates TP orders of active SPOT orders
func TPSpot(db *rdb.DB, bp *t.BotParams, qo t.QueryOrder, ticker t.Ticker, atr float64) []t.Order {
	if bp.QuoteTP <= 0 && bp.AtrTP <= 0 {
		return nil
	}

	var closeOrders []t.Order

	for _, o := range db.GetFilledLimitBuyOrders(qo) {
		if db.GetTPOrder(o.ID) != nil {
			continue
		}

		tpPrice := 0.0
		if bp.QuoteTP > 0 {
			// TP by a value of the quote currency
			tpPrice = o.OpenPrice + bp.QuoteTP/o.Qty
		} else if bp.AtrTP > 0 && atr > 0 {
			// TP by a volatility
			tpPrice = o.OpenPrice + bp.AtrTP*atr
		}

		if tpPrice <= 0 {
			continue
		}

		if ticker.Price > tpPrice {
			tpPrice = h.CalcStopUpperTicker(ticker.Price, float64(bp.Gap.TPLimit), bp.PriceDigits)
			stopPrice := h.CalcStopUpperTicker(ticker.Price, float64(bp.Gap.TPStop), bp.PriceDigits)
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
			closeOrders = append(closeOrders, tpo)
		}
	}

	return closeOrders
}

// TPLong creates TP orders of active LONG orders
func TPLong(db *rdb.DB, bp *t.BotParams, qo t.QueryOrder, ticker t.Ticker, atr float64) []t.Order {
	if bp.QuoteTP <= 0 && bp.AtrTP <= 0 {
		return nil
	}

	var closeOrders []t.Order

	for _, o := range db.GetFilledLimitLongOrders(qo) {
		if db.GetTPOrder(o.ID) != nil {
			continue
		}

		tpPrice := 0.0
		if bp.QuoteTP > 0 {
			// TP by a value of the quote currency
			tpPrice = o.OpenPrice + bp.QuoteTP/o.Qty
		} else if bp.AtrTP > 0 && atr > 0 {
			// TP by a volatility
			tpPrice = o.OpenPrice + bp.AtrTP*atr
		}

		if tpPrice <= 0 {
			continue
		}

		if ticker.Price > tpPrice {
			tpPrice = h.CalcStopUpperTicker(ticker.Price, float64(bp.Gap.TPLimit), bp.PriceDigits)
			stopPrice := h.CalcStopUpperTicker(ticker.Price, float64(bp.Gap.TPStop), bp.PriceDigits)
			tpo := t.Order{
				ID:          h.GenID(),
				BotID:       bp.BotID,
				Exchange:    qo.Exchange,
				Symbol:      qo.Symbol,
				Side:        t.OrderSideSell,
				PosSide:     t.OrderPosSideLong,
				Type:        t.OrderTypeFTP,
				Status:      t.OrderStatusNew,
				Qty:         h.NormalizeDouble(o.Qty, bp.QtyDigits),
				StopPrice:   stopPrice,
				OpenPrice:   tpPrice,
				OpenOrderID: o.ID,
			}
			closeOrders = append(closeOrders, tpo)
		}
	}

	return closeOrders
}

// TPShort creates TP orders of active SHORT orders
func TPShort(db *rdb.DB, bp *t.BotParams, qo t.QueryOrder, ticker t.Ticker, atr float64) []t.Order {
	if bp.QuoteTP <= 0 && bp.AtrTP <= 0 {
		return nil
	}

	var closeOrders []t.Order

	for _, o := range db.GetFilledLimitShortOrders(qo) {
		if db.GetTPOrder(o.ID) != nil {
			continue
		}

		tpPrice := 0.0
		if bp.QuoteTP > 0 {
			// TP by a value of the quote currency
			tpPrice = o.OpenPrice - bp.QuoteTP/o.Qty
		} else if bp.AtrTP > 0 && atr > 0 {
			// TP by a volatility
			tpPrice = o.OpenPrice - bp.AtrTP*atr
		}

		if tpPrice <= 0 {
			continue
		}

		if ticker.Price < tpPrice {
			tpPrice = h.CalcStopLowerTicker(ticker.Price, float64(bp.Gap.TPLimit), bp.PriceDigits)
			stopPrice := h.CalcStopLowerTicker(ticker.Price, float64(bp.Gap.TPStop), bp.PriceDigits)
			tpo := t.Order{
				ID:          h.GenID(),
				BotID:       bp.BotID,
				Exchange:    qo.Exchange,
				Symbol:      qo.Symbol,
				Side:        t.OrderSideBuy,
				PosSide:     t.OrderPosSideShort,
				Type:        t.OrderTypeFTP,
				Status:      t.OrderStatusNew,
				Qty:         h.NormalizeDouble(o.Qty, bp.QtyDigits),
				StopPrice:   stopPrice,
				OpenPrice:   tpPrice,
				OpenOrderID: o.ID,
			}
			closeOrders = append(closeOrders, tpo)
		}
	}

	return closeOrders
}

// SLLongNow creates a SL order of the LONG order from the ticker price
func SLLongNow(db *rdb.DB, bp *t.BotParams, ticker t.Ticker, o t.Order) *t.Order {
	if db.GetSLOrder(o.ID) != nil {
		return nil
	}

	slPrice := h.CalcStopLowerTicker(ticker.Price, float64(bp.Gap.SLLimit), bp.PriceDigits)
	stopPrice := h.CalcStopLowerTicker(ticker.Price, float64(bp.Gap.SLStop), bp.PriceDigits)
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
		slo.PosSide = t.OrderPosSideLong
	}
	return &slo
}

// SLShortNow creates a SL order of the SHORT order from the ticker price
func SLShortNow(db *rdb.DB, bp *t.BotParams, ticker t.Ticker, o t.Order) *t.Order {
	if db.GetSLOrder(o.ID) != nil {
		return nil
	}

	slPrice := h.CalcStopUpperTicker(ticker.Price, float64(bp.Gap.SLLimit), bp.PriceDigits)
	stopPrice := h.CalcStopUpperTicker(ticker.Price, float64(bp.Gap.SLStop), bp.PriceDigits)
	slo := t.Order{
		ID:          h.GenID(),
		BotID:       bp.BotID,
		Exchange:    bp.Exchange,
		Symbol:      bp.Symbol,
		Side:        t.OrderSideBuy,
		PosSide:     t.OrderPosSideShort,
		Type:        t.OrderTypeFSL,
		Status:      t.OrderStatusNew,
		Qty:         h.NormalizeDouble(o.Qty, bp.QtyDigits),
		StopPrice:   stopPrice,
		OpenPrice:   slPrice,
		OpenOrderID: o.ID,
	}
	return &slo
}

// TPLongNow creates a TP order of the LONG order from the ticker price
func TPLongNow(db *rdb.DB, bp *t.BotParams, ticker t.Ticker, o t.Order) *t.Order {
	if db.GetTPOrder(o.ID) != nil {
		return nil
	}

	tpPrice := h.CalcStopUpperTicker(ticker.Price, float64(bp.Gap.TPLimit), bp.PriceDigits)
	stopPrice := h.CalcStopUpperTicker(ticker.Price, float64(bp.Gap.TPStop), bp.PriceDigits)
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
		tpo.PosSide = t.OrderPosSideLong
	}
	return &tpo
}

// TPShortNow creates a TP order of the SHORT order from the ticker price
func TPShortNow(db *rdb.DB, bp *t.BotParams, ticker t.Ticker, o t.Order) *t.Order {
	if db.GetTPOrder(o.ID) != nil {
		return nil
	}

	tpPrice := h.CalcStopLowerTicker(ticker.Price, float64(bp.Gap.TPLimit), bp.PriceDigits)
	stopPrice := h.CalcStopLowerTicker(ticker.Price, float64(bp.Gap.TPStop), bp.PriceDigits)
	tpo := t.Order{
		ID:          h.GenID(),
		BotID:       bp.BotID,
		Exchange:    bp.Exchange,
		Symbol:      bp.Symbol,
		Side:        t.OrderSideBuy,
		PosSide:     t.OrderPosSideShort,
		Type:        t.OrderTypeFTP,
		Status:      t.OrderStatusNew,
		Qty:         h.NormalizeDouble(o.Qty, bp.QtyDigits),
		StopPrice:   stopPrice,
		OpenPrice:   tpPrice,
		OpenOrderID: o.ID,
	}
	return &tpo
}
