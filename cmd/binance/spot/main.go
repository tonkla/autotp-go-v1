package main

import (
	"github.com/tonkla/autotp/common"
	"github.com/tonkla/autotp/exchange/binance"
	s "github.com/tonkla/autotp/strategy/mama"
)

func main() {
	tick := common.Ticker{Exchange: common.Exchange{Name: common.EXC_BINANCE}, Symbol: "BNBBUSD", Price: 0, Qty: 0}
	advice := s.OnTick(tick)
	b := binance.New()
	if advice.Side == "buy" {
		b.OpenOrder(common.Order{
			Side:  "B",
			Price: advice.Price,
			Qty:   advice.Qty,
		})
	} else if advice.Side == "sell" {
		b.CloseOrder(common.Order{
			Side:  "S",
			Price: 0,
			Qty:   0,
		})
	}
}
