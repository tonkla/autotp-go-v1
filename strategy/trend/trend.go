package trend

import (
	"math"

	"github.com/tonkla/autotp/db"
	h "github.com/tonkla/autotp/helper"
	"github.com/tonkla/autotp/strategy"
	"github.com/tonkla/autotp/talib"
	t "github.com/tonkla/autotp/types"
)

type OnTickParams struct {
	DB        db.DB
	Ticker    t.Ticker
	BotParams t.BotParams
	HPrices   []t.HistoricalPrice
	IsFutures bool
}

func OnTick(params OnTickParams) *t.TradeOrders {
	var openOrders, closeOrders []t.Order

	db := params.DB
	ticker := params.Ticker
	p := params.BotParams
	prices := params.HPrices
	isFutures := params.IsFutures

	slStop := float64(p.StopLimit.SLStop)
	tpStop := float64(p.StopLimit.TPStop)
	openLimit := float64(p.StopLimit.OpenLimit)

	p_0 := prices[len(prices)-1]
	if p_0.Open == 0 || p_0.High == 0 || p_0.Low == 0 || p_0.Close == 0 {
		return nil
	}

	var closes []float64
	for _, _p := range prices {
		closes = append(closes, _p.Close)
	}
	cma := talib.WMA(closes, int(p.MAPeriod))
	cma_0 := cma[len(cma)-1]
	cma_1 := cma[len(cma)-2]

	atr := strategy.GetATR(prices, int(p.MAPeriod))

	// Query Order
	qo := t.Order{
		BotID:    p.BotID,
		Exchange: ticker.Exchange,
		Symbol:   ticker.Symbol,
		Qty:      h.NormalizeDouble(p.BaseQty, p.QtyDigits),
	}
	_qty := h.NormalizeDouble(p.QuoteQty/ticker.Price, p.QtyDigits)
	if _qty > p.BaseQty {
		qo.Qty = _qty
	}

	// Stop Loss -----------------------------------------------------------------
	if p.AutoSL {
		// SL Long ----------------------------
		for _, o := range db.GetFilledLimitLongOrders(qo) {
			if db.GetSLOrder(o.ID) != nil {
				continue
			}

			slPrice := 0.0
			if p.QuoteSL > 0 {
				// SL by a value of the quote currency
				slPrice = o.OpenPrice - p.QuoteSL/o.Qty
			} else if p.AtrSL > 0 {
				// SL by a volatility
				slPrice = o.OpenPrice - p.AtrSL*atr
			}

			if slPrice <= 0 {
				continue
			}

			stopPrice := h.CalcSLStop(t.OrderSideBuy, slPrice, slStop, p.PriceDigits)
			if ticker.Price-(stopPrice-slPrice) < stopPrice {
				slo := t.Order{
					ID:          h.GenID(),
					BotID:       p.BotID,
					Exchange:    qo.Exchange,
					Symbol:      qo.Symbol,
					Side:        t.OrderSideSell,
					Type:        t.OrderTypeSL,
					Status:      t.OrderStatusNew,
					Qty:         h.NormalizeDouble(o.Qty, p.QtyDigits),
					StopPrice:   stopPrice,
					OpenPrice:   h.NormalizeDouble(slPrice, p.PriceDigits),
					OpenOrderID: o.ID,
				}
				if isFutures {
					slo.Type = t.OrderTypeFSL
					slo.PosSide = o.PosSide
				}
				closeOrders = append(closeOrders, slo)
			}
		}

		// SL Short ---------------------------
		for _, o := range db.GetFilledLimitShortOrders(qo) {
			if db.GetSLOrder(o.ID) != nil {
				continue
			}

			slPrice := 0.0
			if p.QuoteSL > 0 {
				// SL by a value of the quote currency
				slPrice = o.OpenPrice + p.QuoteSL/o.Qty
			} else if p.AtrSL > 0 {
				// SL by a volatility
				slPrice = o.OpenPrice + p.AtrSL*atr
			}

			if slPrice <= 0 {
				continue
			}

			stopPrice := h.CalcSLStop(t.OrderSideSell, slPrice, slStop, p.PriceDigits)
			if ticker.Price+(slPrice-stopPrice) > stopPrice {
				slo := t.Order{
					ID:          h.GenID(),
					BotID:       p.BotID,
					Exchange:    qo.Exchange,
					Symbol:      qo.Symbol,
					Side:        t.OrderSideBuy,
					Type:        t.OrderTypeSL,
					Status:      t.OrderStatusNew,
					Qty:         h.NormalizeDouble(o.Qty, p.QtyDigits),
					StopPrice:   stopPrice,
					OpenPrice:   h.NormalizeDouble(slPrice, p.PriceDigits),
					OpenOrderID: o.ID,
				}
				if isFutures {
					slo.Side = t.OrderSideSell
					slo.Type = t.OrderTypeFSL
					slo.PosSide = o.PosSide
				}
				closeOrders = append(closeOrders, slo)
			}
		}
	}

	// Take Profit ---------------------------------------------------------------
	if p.AutoTP {
		// TP Long ----------------------------
		for _, o := range db.GetFilledLimitLongOrders(qo) {
			if db.GetTPOrder(o.ID) != nil {
				continue
			}

			tpPrice := 0.0
			if p.QuoteTP > 0 {
				// TP by a value of the quote currency
				tpPrice = o.OpenPrice + p.QuoteTP/o.Qty
			} else if p.AtrTP > 0 {
				// TP by a volatility
				tpPrice = o.OpenPrice + p.AtrTP*atr
			}

			if tpPrice <= 0 {
				continue
			}

			stopPrice := h.CalcTPStop(t.OrderSideBuy, tpPrice, tpStop, p.PriceDigits)
			if ticker.Price+(tpPrice-stopPrice) > stopPrice {
				tpo := t.Order{
					ID:          h.GenID(),
					BotID:       p.BotID,
					Exchange:    qo.Exchange,
					Symbol:      qo.Symbol,
					Side:        t.OrderSideSell,
					Type:        t.OrderTypeTP,
					Status:      t.OrderStatusNew,
					Qty:         h.NormalizeDouble(o.Qty, p.QtyDigits),
					StopPrice:   stopPrice,
					OpenPrice:   h.NormalizeDouble(tpPrice, p.PriceDigits),
					OpenOrderID: o.ID,
				}
				if isFutures {
					tpo.Type = t.OrderTypeFTP
					tpo.PosSide = o.PosSide
				}
				closeOrders = append(closeOrders, tpo)
			}
		}

		// TP Short ---------------------------
		for _, o := range db.GetFilledLimitShortOrders(qo) {
			if db.GetTPOrder(o.ID) != nil {
				continue
			}

			tpPrice := 0.0
			if p.QuoteTP > 0 {
				// TP by a value of the quote currency
				tpPrice = o.OpenPrice - p.QuoteTP/o.Qty
			} else if p.AtrTP > 0 {
				// TP by a volatility
				tpPrice = o.OpenPrice - p.AtrTP*atr
			}

			if tpPrice <= 0 {
				continue
			}

			stopPrice := h.CalcTPStop(t.OrderSideSell, tpPrice, tpStop, p.PriceDigits)
			if ticker.Price-(stopPrice-tpPrice) < stopPrice {
				tpo := t.Order{
					ID:          h.GenID(),
					BotID:       p.BotID,
					Exchange:    qo.Exchange,
					Symbol:      qo.Symbol,
					Side:        t.OrderSideBuy,
					Type:        t.OrderTypeTP,
					Status:      t.OrderStatusNew,
					Qty:         h.NormalizeDouble(o.Qty, p.QtyDigits),
					StopPrice:   stopPrice,
					OpenPrice:   h.NormalizeDouble(tpPrice, p.PriceDigits),
					OpenOrderID: o.ID,
				}
				if isFutures {
					tpo.Side = t.OrderSideSell
					tpo.Type = t.OrderTypeFTP
					tpo.PosSide = o.PosSide
				}
				closeOrders = append(closeOrders, tpo)
			}
		}
	}

	// Uptrend: Open Long --------------------------------------------------------
	if cma_1 < cma_0 {
		if p.View == t.ViewNeutral || p.View == t.ViewLong {
			_openPrice := h.CalcStopBehindTicker(t.OrderSideBuy, ticker.Price, openLimit, p.PriceDigits)
			qo.Side = t.OrderSideBuy
			qo.OpenPrice = _openPrice
			norder := db.GetNearestOrder(qo)
			// Open a new limit order with safe minimum price gap
			if _openPrice < cma_0 && (norder == nil || math.Abs(norder.OpenPrice-_openPrice) >= p.MinGap) {
				o := t.Order{
					ID:        h.GenID(),
					BotID:     p.BotID,
					Exchange:  qo.Exchange,
					Symbol:    qo.Symbol,
					Side:      t.OrderSideBuy,
					Type:      t.OrderTypeLimit,
					Status:    t.OrderStatusNew,
					Qty:       qo.Qty,
					OpenPrice: qo.OpenPrice,
				}
				if isFutures {
					o.PosSide = t.OrderPosSideLong
				}
				openOrders = append(openOrders, o)
			}
		}
	}

	// Downtrend: Open Short -----------------------------------------------------
	if cma_1 > cma_0 {
		if p.View == t.ViewNeutral || p.View == t.ViewShort {
			_openPrice := h.CalcStopBehindTicker(t.OrderSideSell, ticker.Price, openLimit, p.PriceDigits)
			qo.Side = t.OrderSideSell
			qo.OpenPrice = _openPrice
			norder := db.GetNearestOrder(qo)
			// Open a new limit order with safe minimum price gap
			if _openPrice > cma_0 && (norder == nil || math.Abs(norder.OpenPrice-_openPrice) >= p.MinGap) {
				o := t.Order{
					ID:        h.GenID(),
					BotID:     p.BotID,
					Exchange:  qo.Exchange,
					Symbol:    qo.Symbol,
					Side:      t.OrderSideSell,
					Type:      t.OrderTypeLimit,
					Status:    t.OrderStatusNew,
					Qty:       qo.Qty,
					OpenPrice: qo.OpenPrice,
				}
				if isFutures {
					o.Side = t.OrderSideBuy
					o.PosSide = t.OrderPosSideShort
				}
				openOrders = append(openOrders, o)
			}
		}
	}

	return &t.TradeOrders{
		OpenOrders:  openOrders,
		CloseOrders: closeOrders,
	}
}
