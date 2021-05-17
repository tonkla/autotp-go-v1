package main

import (
	"log"

	"github.com/tonkla/autotp/db"
	binance "github.com/tonkla/autotp/exchange/binance/fusd"
	strategy "github.com/tonkla/autotp/strategy/grid"
)

func main() {
	symbol := ""
	ticker := binance.GetTicker(symbol)
	orders := strategy.OnTick(ticker)
	if len(orders) == 0 {
		return
	}

	for _, order := range orders {
		result := binance.Trade(order)
		if result == nil {
			continue
		}

		err := db.CreateRecord(*result)
		if err != nil {
			log.Println(err)
		}
	}
}
