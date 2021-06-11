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

	bar := prices[len(prices)-1]
	if bar.Open == 0 || bar.High == 0 || bar.Low == 0 || bar.Close == 0 {
		return nil
	}

	var openOrders, closeOrders []types.Order

	trend := strategy.GetTrend(prices, int(params.MAPeriod))

	order := types.Order{
		BotID:    params.BotID,
		Exchange: ticker.Exchange,
		Symbol:   ticker.Symbol,
	}

	if trend > 0 {
		// Stop Loss
		order.Side = types.SIDE_SELL
		closeOrders = append(closeOrders, db.GetActiveOrders(order)...)

		// Take Profit

		// Open a new limit order
	} else if trend < 0 {
		// Stop Loss
		order.Side = types.SIDE_BUY
		closeOrders = append(closeOrders, db.GetActiveOrders(order)...)

		// Take Profit

		// Open a new limit order
	}

	return &types.TradeOrders{
		OpenOrders:  openOrders,
		CloseOrders: closeOrders,
	}
}
