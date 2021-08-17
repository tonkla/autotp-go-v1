package daily

import (
	"github.com/tonkla/autotp/db"
	h "github.com/tonkla/autotp/helper"
	"github.com/tonkla/autotp/strategy"
	t "github.com/tonkla/autotp/types"
)

type OnTickParams struct {
	DB        db.DB
	Ticker    t.Ticker
	BotParams t.BotParams
	HPrices   []t.HistoricalPrice
}

func OnTick(params OnTickParams) *t.TradeOrders {
	var openOrders, closeOrders []t.Order

	db := params.DB
	ticker := params.Ticker
	p := params.BotParams
	prices := params.HPrices

	p_0 := prices[len(prices)-1]
	if p_0.Open == 0 || p_0.High == 0 || p_0.Low == 0 || p_0.Close == 0 {
		return nil
	}
	p_1 := prices[len(prices)-2]
	c_1 := p_1.Close
	h_1 := p_1.High
	l_1 := p_1.Low

	trend := strategy.GetTrend(prices, int(p.MAPeriod))
	atr := strategy.GetATR(prices, int(p.MAPeriod))
	mos := (h_1 - l_1) / 3 // The Margin of Safety

	// Query Order
	qo := t.Order{
		BotID:    p.BotID,
		Exchange: ticker.Exchange,
		Symbol:   ticker.Symbol,
		Qty:      p.BaseQty,
	}
	_qty := h.NormalizeDouble(p.QuoteQty/ticker.Price, p.QtyDigits)
	if _qty > p.BaseQty {
		qo.Qty = _qty
	}

	// Uptrend -------------------------------------------------------------------
	if trend >= t.TrendUp1 {
		// Stop Loss, for SELL orders
		if p.AutoSL {
			qo.Side = t.OrderSideSell
			for _, o := range db.GetFilledOrdersBySide(qo) {
				o.Side = t.OrderSideBuy
				o.Type = t.OrderTypeSL
				o.StopPrice = h.CalcSLStop(t.OrderSideBuy, ticker.Price, 300, p.PriceDigits)
				o.OpenPrice = h.CalcSLStop(t.OrderSideBuy, ticker.Price, 500, p.PriceDigits)
				closeOrders = append(closeOrders, o)
			}
		}

		// Take Profit, by the configured Volatility Stop (TP)
		if p.AutoTP {
			qo.Side = t.OrderSideBuy
			for _, o := range db.GetFilledOrdersBySide(qo) {
				if ticker.Price > o.OpenPrice+atr*p.AtrTP {
					o.Side = t.OrderSideSell
					o.Type = t.OrderTypeTP
					o.StopPrice = h.CalcTPStop(t.OrderSideBuy, ticker.Price, 300, p.PriceDigits)
					o.OpenPrice = h.CalcTPStop(t.OrderSideBuy, ticker.Price, 500, p.PriceDigits)
					closeOrders = append(closeOrders, o)
				}
			}
		}

		// Open a new limit order, when no active BUY order
		if (p.View == t.ViewLong || p.View == t.ViewNeutral) && ticker.Price < h_1-mos && ticker.Price < c_1 {
			qo.Side = t.OrderSideBuy
			qo.Type = t.OrderTypeLimit
			qo.OpenPrice = h.CalcTPStop(t.OrderSideBuy, ticker.Price, 300, p.PriceDigits)
			if len(db.GetLimitOrdersBySide(qo)) == 0 {
				openOrders = append(openOrders, qo)
			}
		}
	}

	// Downtrend -----------------------------------------------------------------
	if trend <= t.TrendDown1 {
		// Stop Loss, for BUY orders
		if p.AutoSL {
			qo.Side = t.OrderSideBuy
			for _, o := range db.GetFilledOrdersBySide(qo) {
				o.Side = t.OrderSideSell
				o.Type = t.OrderTypeSL
				o.StopPrice = h.CalcSLStop(t.OrderSideSell, ticker.Price, 300, p.PriceDigits)
				o.OpenPrice = h.CalcSLStop(t.OrderSideSell, ticker.Price, 500, p.PriceDigits)
				closeOrders = append(closeOrders, o)
			}
		}

		// Take Profit, by the configured Volatility Stop (TP)
		if p.AutoTP {
			qo.Side = t.OrderSideSell
			for _, o := range db.GetFilledOrdersBySide(qo) {
				if ticker.Price < o.OpenPrice-atr*p.AtrTP {
					o.Side = t.OrderSideBuy
					o.Type = t.OrderTypeTP
					o.StopPrice = h.CalcTPStop(t.OrderSideSell, ticker.Price, 300, p.PriceDigits)
					o.OpenPrice = h.CalcTPStop(t.OrderSideSell, ticker.Price, 500, p.PriceDigits)
					closeOrders = append(closeOrders, o)
				}
			}
		}

		// Open a new limit order, when no active SELL order
		if (p.View == t.ViewShort || p.View == t.ViewNeutral) && ticker.Price > l_1+mos && ticker.Price > c_1 {
			qo.Side = t.OrderSideSell
			qo.Type = t.OrderTypeLimit
			qo.OpenPrice = h.CalcTPStop(t.OrderSideSell, ticker.Price, 300, p.PriceDigits)
			if len(db.GetLimitOrdersBySide(qo)) == 0 {
				openOrders = append(openOrders, qo)
			}
		}
	}

	return &t.TradeOrders{
		OpenOrders:  openOrders,
		CloseOrders: closeOrders,
	}
}
