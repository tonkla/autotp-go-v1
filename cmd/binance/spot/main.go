package main

import (
	binance "github.com/tonkla/autotp/exchange/binance/spot"
	s "github.com/tonkla/autotp/strategy/mama"
	"github.com/tonkla/autotp/types"
)

func main() {
	tick := types.Ticker{Exchange: types.Exchange{Name: types.EXC_BINANCE}, Symbol: "BNBBUSD", Price: 0, Qty: 0}
	advice := s.OnTick(tick)
	b := binance.New()
	if advice.Side == "buy" {
		b.OpenOrder(types.Order{
			Side:  "B",
			Price: advice.Price,
			Qty:   advice.Qty,
		})
	} else if advice.Side == "sell" {
		b.CloseOrder(types.Order{
			Side:  "S",
			Price: 0,
			Qty:   0,
		})
	}
}
