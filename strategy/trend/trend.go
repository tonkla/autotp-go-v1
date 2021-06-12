package trend

import (
	"github.com/tonkla/autotp/db"
	"github.com/tonkla/autotp/strategy"
	"github.com/tonkla/autotp/types"
)

type OnTickParams struct {
	Ticker    types.Ticker
	BotParams types.BotParams
	HPrices   []types.HistoricalPrice
	DB        db.DB
}

func OnTick(p OnTickParams) *types.TradeOrders {
	ticker := p.Ticker
	params := p.BotParams
	prices := p.HPrices
	db := p.DB

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

	var openOrders, closeOrders []types.Order

	trend := strategy.GetTrend(prices, int(params.MAPeriod))
	upperPrice := strategy.GetUpperPrice(ticker.Price)
	lowerPrice := strategy.GetLowerPrice(ticker.Price)

	// Query Order
	qo := types.Order{
		BotID:    params.BotID,
		Exchange: ticker.Exchange,
		Symbol:   ticker.Symbol,
		Qty:      params.Qty,
	}

	// Uptrend
	if trend >= types.TREND_UP_1 {
		// Stop Loss, for SELL orders
		qo.Side = types.SIDE_SELL
		closeOrders = append(closeOrders, db.GetActiveOrders(qo)...)

		// Take Profit, when lower low or previous bar was red
		if c_0 < l_1 || c_1 < o_1 {
			qo.Side = types.SIDE_BUY
			qo.ClosePrice = upperPrice
			for _, o := range db.GetProfitOrders(qo) {
				o.ClosePrice = upperPrice
				closeOrders = append(closeOrders, o)
			}
		}

		// Open a new limit order, when no active BUY order at the price
		if trend < types.TREND_UP_3 {
			qo.Side = types.SIDE_BUY
			qo.OpenPrice = lowerPrice
			qo.ClosePrice = 0
			if db.GetActiveOrder(qo, params.Slippage) == nil {
				openOrders = append(openOrders, qo)
			}
		}
	}

	// Downtrend
	if trend <= types.TREND_DOWN_1 {
		// Stop Loss, for BUY orders
		qo.Side = types.SIDE_BUY
		closeOrders = append(closeOrders, db.GetActiveOrders(qo)...)

		// Take Profit, when higher high or previous bar was green
		if c_0 > h_1 || c_1 > o_1 {
			qo.Side = types.SIDE_SELL
			qo.ClosePrice = lowerPrice
			for _, o := range db.GetProfitOrders(qo) {
				o.ClosePrice = lowerPrice
				closeOrders = append(closeOrders, o)
			}
		}

		// Open a new limit order, when no active SELL order at the price
		if trend > types.TREND_DOWN_3 {
			qo.Side = types.SIDE_SELL
			qo.OpenPrice = upperPrice
			qo.ClosePrice = 0
			if db.GetActiveOrder(qo, params.Slippage) == nil {
				openOrders = append(openOrders, qo)
			}
		}
	}

	return &types.TradeOrders{
		OpenOrders:  openOrders,
		CloseOrders: closeOrders,
	}
}
