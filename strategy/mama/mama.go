package macd

import (
	"github.com/tonkla/autotp/common"
)

func OnTick(ticker common.Ticker) *common.Advice {
	advice := &common.Advice{
		Symbol: ticker.Symbol,
		Side:   "",
		Price:  0,
		Qty:    0,
	}
	return advice
}
