package trend

import (
	"math"

	"github.com/tonkla/autotp/db"
	h "github.com/tonkla/autotp/helper"
	s "github.com/tonkla/autotp/strategy/common"
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
	bp := params.BotParams
	prices := params.HPrices
	isFutures := params.IsFutures

	slStop := float64(bp.SLim.SLStop)
	tpStop := float64(bp.SLim.TPStop)
	openLimit := float64(bp.SLim.OpenLimit)

	p_0 := prices[len(prices)-1]
	if p_0.Open == 0 || p_0.High == 0 || p_0.Low == 0 || p_0.Close == 0 {
		return nil
	}

	var closes []float64
	for _, _p := range prices {
		closes = append(closes, _p.Close)
	}
	cma := talib.WMA(closes, int(bp.MAPeriod))
	cma_0 := cma[len(cma)-1]
	cma_1 := cma[len(cma)-2]

	atr := s.GetATR(prices, int(bp.MAPeriod))

	qo := t.QueryOrder{
		BotID:    bp.BotID,
		Exchange: ticker.Exchange,
		Symbol:   ticker.Symbol,
		Qty:      h.NormalizeDouble(bp.BaseQty, bp.QtyDigits),
	}

	_qty := h.NormalizeDouble(bp.QuoteQty/ticker.Price, bp.QtyDigits)
	if _qty > bp.BaseQty {
		qo.Qty = _qty
	}

	// Stop Loss -----------------------------------------------------------------
	if bp.AutoSL {
		// SL Long ----------------------------
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

			stopPrice := h.CalcSLStop(t.OrderSideBuy, slPrice, slStop, bp.PriceDigits)
			if ticker.Price-(stopPrice-slPrice) < stopPrice {
				slo := t.Order{
					ID:          h.GenID(),
					BotID:       bp.BotID,
					Exchange:    qo.Exchange,
					Symbol:      qo.Symbol,
					Side:        t.OrderSideSell,
					Type:        t.OrderTypeSL,
					Status:      t.OrderStatusNew,
					Qty:         h.NormalizeDouble(o.Qty, bp.QtyDigits),
					StopPrice:   stopPrice,
					OpenPrice:   h.NormalizeDouble(slPrice, bp.PriceDigits),
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

			stopPrice := h.CalcSLStop(t.OrderSideSell, slPrice, slStop, bp.PriceDigits)
			if ticker.Price+(slPrice-stopPrice) > stopPrice {
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
					OpenPrice:   h.NormalizeDouble(slPrice, bp.PriceDigits),
					OpenOrderID: o.ID,
				}
				if isFutures {
					slo.Type = t.OrderTypeFSL
					slo.PosSide = o.PosSide
				}
				closeOrders = append(closeOrders, slo)
			}
		}
	}

	// Take Profit ---------------------------------------------------------------
	if bp.AutoTP {
		// TP Long ----------------------------
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

			stopPrice := h.CalcTPStop(t.OrderSideBuy, tpPrice, tpStop, bp.PriceDigits)
			if ticker.Price+(tpPrice-stopPrice) > stopPrice {
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
					OpenPrice:   h.NormalizeDouble(tpPrice, bp.PriceDigits),
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

			stopPrice := h.CalcTPStop(t.OrderSideSell, tpPrice, tpStop, bp.PriceDigits)
			if ticker.Price-(stopPrice-tpPrice) < stopPrice {
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
					OpenPrice:   h.NormalizeDouble(tpPrice, bp.PriceDigits),
					OpenOrderID: o.ID,
				}
				if isFutures {
					tpo.Type = t.OrderTypeFTP
					tpo.PosSide = o.PosSide
				}
				closeOrders = append(closeOrders, tpo)
			}
		}
	}

	// Uptrend: Open Long --------------------------------------------------------
	if cma_1 < cma_0 {
		if bp.View == t.ViewNeutral || bp.View == t.ViewLong {
			_openPrice := h.CalcStopBehindTicker(t.OrderSideBuy, ticker.Price, openLimit, bp.PriceDigits)
			qo.Side = t.OrderSideBuy
			qo.OpenPrice = _openPrice
			norder := db.GetNearestOrder(qo)
			// Open a new limit order with safe minimum price gap
			if _openPrice < cma_0 && (norder == nil || math.Abs(norder.OpenPrice-_openPrice) >= bp.OrderGap) {
				o := t.Order{
					ID:        h.GenID(),
					BotID:     bp.BotID,
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
		if bp.View == t.ViewNeutral || bp.View == t.ViewShort {
			_openPrice := h.CalcStopBehindTicker(t.OrderSideSell, ticker.Price, openLimit, bp.PriceDigits)
			qo.Side = t.OrderSideSell
			qo.OpenPrice = _openPrice
			norder := db.GetNearestOrder(qo)
			// Open a new limit order with safe minimum price gap
			if _openPrice > cma_0 && (norder == nil || math.Abs(norder.OpenPrice-_openPrice) >= bp.OrderGap) {
				o := t.Order{
					ID:        h.GenID(),
					BotID:     bp.BotID,
					Exchange:  qo.Exchange,
					Symbol:    qo.Symbol,
					Side:      t.OrderSideSell,
					Type:      t.OrderTypeLimit,
					Status:    t.OrderStatusNew,
					Qty:       qo.Qty,
					OpenPrice: qo.OpenPrice,
				}
				if isFutures {
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
