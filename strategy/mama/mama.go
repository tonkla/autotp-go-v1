package macd

import "github.com/tonkla/autotp/types"

func OnTick(ticker types.Ticker) *types.Advice {
	advice := &types.Advice{
		Symbol: ticker.Symbol,
		Side:   "",
		Price:  0,
		Qty:    0,
	}
	return advice
}
