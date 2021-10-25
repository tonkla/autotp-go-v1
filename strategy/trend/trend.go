package trend

import (
	"math"

	h "github.com/tonkla/autotp/helper"
	"github.com/tonkla/autotp/rdb"
	"github.com/tonkla/autotp/strategy/common"
	"github.com/tonkla/autotp/talib"
	t "github.com/tonkla/autotp/types"
)

type Strategy struct {
	DB *rdb.DB
	BP *t.BotParams
}

func New(db *rdb.DB, bp *t.BotParams) Strategy {
	return Strategy{
		DB: db,
		BP: bp,
	}
}

func (s Strategy) OnTick(ticker t.Ticker) *t.TradeOrders {
	var openOrders, closeOrders []t.Order

	prices := s.BP.HPrices

	p_0 := prices[len(prices)-1]
	if p_0.Open == 0 || p_0.High == 0 || p_0.Low == 0 || p_0.Close == 0 {
		return nil
	}

	var closes []float64
	for _, _p := range prices {
		closes = append(closes, _p.Close)
	}
	cma := talib.WMA(closes, int(s.BP.MAPeriod))
	cma_0 := cma[len(cma)-1]
	cma_1 := cma[len(cma)-2]

	atr := common.GetATR(prices, int(s.BP.MAPeriod))

	qo := t.QueryOrder{
		BotID:    s.BP.BotID,
		Exchange: ticker.Exchange,
		Symbol:   ticker.Symbol,
		Qty:      h.NormalizeDouble(s.BP.BaseQty, s.BP.QtyDigits),
	}

	_qty := h.NormalizeDouble(s.BP.QuoteQty/ticker.Price, s.BP.QtyDigits)
	if _qty > s.BP.BaseQty {
		qo.Qty = _qty
	}

	slStop := float64(s.BP.SLim.SLStop)
	tpStop := float64(s.BP.SLim.TPStop)
	openLimit := float64(s.BP.SLim.OpenLimit)
	isFutures := s.BP.Product == t.ProductFutures

	// Stop Loss -----------------------------------------------------------------
	if s.BP.AutoSL {
		// SL Long ----------------------------
		for _, o := range s.DB.GetFilledLimitLongOrders(qo) {
			if s.DB.GetSLOrder(o.ID) != nil {
				continue
			}

			slPrice := 0.0
			if s.BP.QuoteSL > 0 {
				// SL by a value of the quote currency
				slPrice = o.OpenPrice - s.BP.QuoteSL/o.Qty
			} else if s.BP.AtrSL > 0 {
				// SL by a volatility
				slPrice = o.OpenPrice - s.BP.AtrSL*atr
			}

			if slPrice <= 0 {
				continue
			}

			stopPrice := h.CalcSLStop(t.OrderSideBuy, slPrice, slStop, s.BP.PriceDigits)
			if ticker.Price-(stopPrice-slPrice) < stopPrice {
				slo := t.Order{
					ID:          h.GenID(),
					BotID:       s.BP.BotID,
					Exchange:    qo.Exchange,
					Symbol:      qo.Symbol,
					Side:        t.OrderSideSell,
					Type:        t.OrderTypeSL,
					Status:      t.OrderStatusNew,
					Qty:         h.NormalizeDouble(o.Qty, s.BP.QtyDigits),
					StopPrice:   stopPrice,
					OpenPrice:   h.NormalizeDouble(slPrice, s.BP.PriceDigits),
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
		for _, o := range s.DB.GetFilledLimitShortOrders(qo) {
			if s.DB.GetSLOrder(o.ID) != nil {
				continue
			}

			slPrice := 0.0
			if s.BP.QuoteSL > 0 {
				// SL by a value of the quote currency
				slPrice = o.OpenPrice + s.BP.QuoteSL/o.Qty
			} else if s.BP.AtrSL > 0 {
				// SL by a volatility
				slPrice = o.OpenPrice + s.BP.AtrSL*atr
			}

			if slPrice <= 0 {
				continue
			}

			stopPrice := h.CalcSLStop(t.OrderSideSell, slPrice, slStop, s.BP.PriceDigits)
			if ticker.Price+(slPrice-stopPrice) > stopPrice {
				slo := t.Order{
					ID:          h.GenID(),
					BotID:       s.BP.BotID,
					Exchange:    qo.Exchange,
					Symbol:      qo.Symbol,
					Side:        t.OrderSideBuy,
					Type:        t.OrderTypeSL,
					Status:      t.OrderStatusNew,
					Qty:         h.NormalizeDouble(o.Qty, s.BP.QtyDigits),
					StopPrice:   stopPrice,
					OpenPrice:   h.NormalizeDouble(slPrice, s.BP.PriceDigits),
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
	if s.BP.AutoTP {
		// TP Long ----------------------------
		for _, o := range s.DB.GetFilledLimitLongOrders(qo) {
			if s.DB.GetTPOrder(o.ID) != nil {
				continue
			}

			tpPrice := 0.0
			if s.BP.QuoteTP > 0 {
				// TP by a value of the quote currency
				tpPrice = o.OpenPrice + s.BP.QuoteTP/o.Qty
			} else if s.BP.AtrTP > 0 {
				// TP by a volatility
				tpPrice = o.OpenPrice + s.BP.AtrTP*atr
			}

			if tpPrice <= 0 {
				continue
			}

			stopPrice := h.CalcTPStop(t.OrderSideBuy, tpPrice, tpStop, s.BP.PriceDigits)
			if ticker.Price+(tpPrice-stopPrice) > stopPrice {
				tpo := t.Order{
					ID:          h.GenID(),
					BotID:       s.BP.BotID,
					Exchange:    qo.Exchange,
					Symbol:      qo.Symbol,
					Side:        t.OrderSideSell,
					Type:        t.OrderTypeTP,
					Status:      t.OrderStatusNew,
					Qty:         h.NormalizeDouble(o.Qty, s.BP.QtyDigits),
					StopPrice:   stopPrice,
					OpenPrice:   h.NormalizeDouble(tpPrice, s.BP.PriceDigits),
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
		for _, o := range s.DB.GetFilledLimitShortOrders(qo) {
			if s.DB.GetTPOrder(o.ID) != nil {
				continue
			}

			tpPrice := 0.0
			if s.BP.QuoteTP > 0 {
				// TP by a value of the quote currency
				tpPrice = o.OpenPrice - s.BP.QuoteTP/o.Qty
			} else if s.BP.AtrTP > 0 {
				// TP by a volatility
				tpPrice = o.OpenPrice - s.BP.AtrTP*atr
			}

			if tpPrice <= 0 {
				continue
			}

			stopPrice := h.CalcTPStop(t.OrderSideSell, tpPrice, tpStop, s.BP.PriceDigits)
			if ticker.Price-(stopPrice-tpPrice) < stopPrice {
				tpo := t.Order{
					ID:          h.GenID(),
					BotID:       s.BP.BotID,
					Exchange:    qo.Exchange,
					Symbol:      qo.Symbol,
					Side:        t.OrderSideBuy,
					Type:        t.OrderTypeTP,
					Status:      t.OrderStatusNew,
					Qty:         h.NormalizeDouble(o.Qty, s.BP.QtyDigits),
					StopPrice:   stopPrice,
					OpenPrice:   h.NormalizeDouble(tpPrice, s.BP.PriceDigits),
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
		if s.BP.View == t.ViewNeutral || s.BP.View == t.ViewLong {
			_openPrice := h.CalcStopBehindTicker(t.OrderSideBuy, ticker.Price, openLimit, s.BP.PriceDigits)
			qo.Side = t.OrderSideBuy
			qo.OpenPrice = _openPrice
			norder := s.DB.GetNearestOrder(qo)
			// Open a new limit order with safe minimum price gap
			if _openPrice < cma_0 && (norder == nil || math.Abs(norder.OpenPrice-_openPrice) >= s.BP.OrderGap) {
				o := t.Order{
					ID:        h.GenID(),
					BotID:     s.BP.BotID,
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
		if s.BP.View == t.ViewNeutral || s.BP.View == t.ViewShort {
			_openPrice := h.CalcStopBehindTicker(t.OrderSideSell, ticker.Price, openLimit, s.BP.PriceDigits)
			qo.Side = t.OrderSideSell
			qo.OpenPrice = _openPrice
			norder := s.DB.GetNearestOrder(qo)
			// Open a new limit order with safe minimum price gap
			if _openPrice > cma_0 && (norder == nil || math.Abs(norder.OpenPrice-_openPrice) >= s.BP.OrderGap) {
				o := t.Order{
					ID:        h.GenID(),
					BotID:     s.BP.BotID,
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
