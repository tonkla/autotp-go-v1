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
	slLimit := float64(p.StopLimit.SLLimit)
	tpStop := float64(p.StopLimit.TPStop)
	tpLimit := float64(p.StopLimit.TPLimit)
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

	// Uptrend -------------------------------------------------------------------
	if cma_1 < cma_0 {
		// Stop Loss, for SELL orders
		if p.AutoSL {
			qo.Side = t.OrderSideSell
			for _, o := range db.GetFilledLimitOrdersBySide(qo) {
				if db.GetSLOrder(o.ID) != nil {
					continue
				}
				slo := t.Order{
					ID:          h.GenID(),
					BotID:       p.BotID,
					Exchange:    qo.Exchange,
					Symbol:      qo.Symbol,
					Side:        t.OrderSideBuy,
					Type:        t.OrderTypeSL,
					Status:      t.OrderStatusNew,
					Qty:         h.NormalizeDouble(o.Qty, p.QtyDigits),
					StopPrice:   h.CalcSLStop(t.OrderSideBuy, ticker.Price, slStop, p.PriceDigits),
					OpenPrice:   h.CalcSLStop(t.OrderSideBuy, ticker.Price, slLimit, p.PriceDigits),
					OpenOrderID: o.ID,
				}
				if isFutures {
					slo.PosSide = o.PosSide
				}
				closeOrders = append(closeOrders, slo)
			}
		}

		// Take Profit, by the configured Volatility Stop (TP)
		if p.AutoTP {
			qo.Side = t.OrderSideBuy
			for _, o := range db.GetFilledLimitOrdersBySide(qo) {
				if ticker.Price > o.OpenPrice+atr*p.AtrTP && db.GetTPOrder(o.ID) == nil {
					tpo := t.Order{
						ID:          h.GenID(),
						BotID:       p.BotID,
						Exchange:    qo.Exchange,
						Symbol:      qo.Symbol,
						Side:        t.OrderSideSell,
						Type:        t.OrderTypeTP,
						Status:      t.OrderStatusNew,
						Qty:         h.NormalizeDouble(o.Qty, p.QtyDigits),
						StopPrice:   h.CalcTPStop(t.OrderSideSell, ticker.Price, tpStop, p.PriceDigits),
						OpenPrice:   h.CalcTPStop(t.OrderSideSell, ticker.Price, tpLimit, p.PriceDigits),
						OpenOrderID: o.ID,
					}
					if isFutures {
						tpo.PosSide = o.PosSide
					}
					closeOrders = append(closeOrders, tpo)
				}
			}
		}

		// Open a new limit order with safe minimum price gap
		if p.View == t.ViewLong || p.View == t.ViewNeutral {
			_openPrice := h.CalcAfterLimitStop(t.OrderSideBuy, ticker.Price, openLimit, p.PriceDigits)
			qo.Side = t.OrderSideBuy
			qo.OpenPrice = _openPrice
			norder := db.GetNearestOrder(qo)
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

	// Downtrend -----------------------------------------------------------------
	if cma_1 > cma_0 {
		// Stop Loss, for BUY orders
		if p.AutoSL {
			qo.Side = t.OrderSideBuy
			for _, o := range db.GetFilledLimitOrdersBySide(qo) {
				if db.GetSLOrder(o.ID) != nil {
					continue
				}
				slo := t.Order{
					ID:          h.GenID(),
					BotID:       p.BotID,
					Exchange:    qo.Exchange,
					Symbol:      qo.Symbol,
					Side:        t.OrderSideSell,
					Type:        t.OrderTypeSL,
					Status:      t.OrderStatusNew,
					Qty:         h.NormalizeDouble(o.Qty, p.QtyDigits),
					StopPrice:   h.CalcSLStop(t.OrderSideSell, ticker.Price, slStop, p.PriceDigits),
					OpenPrice:   h.CalcSLStop(t.OrderSideSell, ticker.Price, slLimit, p.PriceDigits),
					OpenOrderID: o.ID,
				}
				if isFutures {
					slo.PosSide = o.PosSide
				}
				closeOrders = append(closeOrders, slo)
			}
		}

		// Take Profit, by the configured Volatility Stop (TP)
		if p.AutoTP {
			qo.Side = t.OrderSideSell
			for _, o := range db.GetFilledLimitOrdersBySide(qo) {
				if ticker.Price < o.OpenPrice-atr*p.AtrTP && db.GetTPOrder(o.ID) == nil {
					tpo := t.Order{
						ID:          h.GenID(),
						BotID:       p.BotID,
						Exchange:    qo.Exchange,
						Symbol:      qo.Symbol,
						Side:        t.OrderSideBuy,
						Type:        t.OrderTypeTP,
						Status:      t.OrderStatusNew,
						Qty:         h.NormalizeDouble(o.Qty, p.QtyDigits),
						StopPrice:   h.CalcTPStop(t.OrderSideBuy, ticker.Price, tpStop, p.PriceDigits),
						OpenPrice:   h.CalcTPStop(t.OrderSideBuy, ticker.Price, tpLimit, p.PriceDigits),
						OpenOrderID: o.ID,
					}
					if isFutures {
						tpo.PosSide = o.PosSide
					}
					closeOrders = append(closeOrders, tpo)
				}
			}
		}

		// Open a new limit order with safe minimum price gap
		if p.View == t.ViewShort || p.View == t.ViewNeutral {
			_openPrice := h.CalcAfterLimitStop(t.OrderSideSell, ticker.Price, openLimit, p.PriceDigits)
			qo.Side = t.OrderSideSell
			qo.OpenPrice = _openPrice
			norder := db.GetNearestOrder(qo)
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
