package trend

import (
	"strings"

	"github.com/tonkla/autotp/db"
	"github.com/tonkla/autotp/strategy"
	t "github.com/tonkla/autotp/types"
)

type OnTickParams struct {
	Ticker    t.Ticker
	OrderBook t.OrderBook
	BotParams t.BotParams
	HPrices   []t.HistoricalPrice
	DB        db.DB
}

func OnTick(params OnTickParams) *t.TradeOrders {
	ticker := params.Ticker
	odbook := params.OrderBook
	p := params.BotParams
	prices := params.HPrices
	db := params.DB

	p_0 := prices[len(prices)-1]
	if p_0.Open == 0 || p_0.High == 0 || p_0.Low == 0 || p_0.Close == 0 {
		return nil
	}
	p_1 := prices[len(prices)-2]
	o_1 := p_1.Open
	c_0 := p_0.Close
	c_1 := p_1.Close
	h_1 := p_1.High
	l_1 := p_1.Low

	var openOrders, closeOrders []t.Order

	trend := strategy.GetTrend(prices, int(p.MAPeriod))
	upperPrice := odbook.Asks[1].Price
	lowerPrice := odbook.Bids[1].Price

	// Query Order
	qo := t.Order{
		BotID:    p.BotID,
		Exchange: ticker.Exchange,
		Symbol:   ticker.Symbol,
		Qty:      p.Qty,
	}

	// Uptrend
	if trend >= t.TrendUp1 {
		// Stop Loss, for SELL orders
		qo.Side = t.OrderSideSell
		for _, o := range db.GetActiveOrders(qo) {
			o.ClosePrice = lowerPrice
			closeOrders = append(closeOrders, o)
		}

		// Take Profit, when lower low or previous bar was red
		if c_0 < l_1 || c_1 < o_1 {
			qo.Side = t.OrderSideBuy
			qo.ClosePrice = upperPrice
			for _, o := range db.GetProfitOrders(qo) {
				o.ClosePrice = upperPrice
				closeOrders = append(closeOrders, o)
			}
		}

		// Open a new limit order, when no active BUY order
		if trend < t.TrendUp4 {
			qo.Side = t.OrderSideBuy
			qo.OpenPrice = lowerPrice
			qo.ClosePrice = 0
			if len(db.GetActiveOrders(qo)) == 0 {
				openOrders = append(openOrders, qo)
			}
		}
	}

	// Downtrend
	v := strings.ToUpper(p.View)
	if trend <= t.TrendDown1 && (v == t.ViewNeutral || v == "N" || v == t.ViewShort || v == "S") {
		// Stop Loss, for BUY orders
		qo.Side = t.OrderSideBuy
		closeOrders = append(closeOrders, db.GetActiveOrders(qo)...)

		// Take Profit, when higher high or previous bar was green
		if c_0 > h_1 || c_1 > o_1 {
			qo.Side = t.OrderSideSell
			qo.ClosePrice = lowerPrice
			for _, o := range db.GetProfitOrders(qo) {
				o.ClosePrice = lowerPrice
				closeOrders = append(closeOrders, o)
			}
		}

		// Open a new limit order, when no active SELL order
		if trend > t.TrendDown4 {
			qo.Side = t.OrderSideSell
			qo.OpenPrice = upperPrice
			qo.ClosePrice = 0
			if len(db.GetActiveOrders(qo)) == 0 {
				openOrders = append(openOrders, qo)
			}
		}
	}

	return &t.TradeOrders{
		OpenOrders:  openOrders,
		CloseOrders: closeOrders,
	}
}
