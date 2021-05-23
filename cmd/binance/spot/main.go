package main

import (
	binance "github.com/tonkla/autotp/exchange/binance/spot"
	strategy "github.com/tonkla/autotp/strategy/mama"
	"github.com/tonkla/autotp/types"
)

func main() {
	tick := types.Ticker{Symbol: "BNBBUSD", Price: 0, Qty: 0}
	order := strategy.OnTick(tick)
	b := binance.New()
	if order.Side == types.SIDE_BUY {
		b.OpenOrder(types.Order{
			Side:  order.Side,
			Price: order.Price,
			Qty:   order.Qty,
		})
	} else if order.Side == types.SIDE_SELL {
		b.CloseOrder(types.Order{
			Side:  order.Side,
			Price: 0,
			Qty:   0,
		})
	}
}
