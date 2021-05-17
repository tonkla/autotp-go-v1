package macd

import "github.com/tonkla/autotp/types"

func OnTick(ticker types.Ticker) *types.Order {
	order := &types.Order{
		Symbol: ticker.Symbol,
		Side:   "",
		Price:  0,
		Qty:    0,
	}
	return order
}
