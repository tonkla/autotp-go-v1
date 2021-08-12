package main

import (
	"fmt"
	"sync"

	"github.com/tonkla/autotp/exchange/bitkub"
	"github.com/tonkla/autotp/exchange/satang"
	"github.com/tonkla/autotp/types"
)

type tmp struct {
	exchange string
	book     types.OrderBook
}

func main() {
	var wg sync.WaitGroup
	ch := make(chan tmp, 2)

	wg.Add(2)

	go func() {
		wg.Wait()
		close(ch)
	}()

	go func() {
		defer wg.Done()
		c := bitkub.NewClient()
		ch <- tmp{exchange: types.ExcBitkub, book: c.GetOrderBook("thb_bnb", 5)}
	}()

	go func() {
		defer wg.Done()
		c := satang.NewClient()
		ch <- tmp{exchange: types.ExcSatang, book: c.GetOrderBook("bnb_thb", 5)}
	}()

	var bitkub types.OrderBook
	var satang types.OrderBook

	for book := range ch {
		if book.exchange == types.ExcBitkub {
			bitkub = book.book
		} else if book.exchange == types.ExcSatang {
			satang = book.book
		}
	}

	bestBid := tmp{exchange: types.ExcBitkub, book: bitkub}
	if satang.Bids[0].Price > bitkub.Bids[0].Price {
		bestBid = tmp{exchange: types.ExcSatang, book: satang}
	}

	bestAsk := tmp{exchange: types.ExcBitkub, book: bitkub}
	if satang.Asks[0].Price < bitkub.Asks[0].Price {
		bestAsk = tmp{exchange: types.ExcSatang, book: satang}
	}

	bidPrice := bestBid.book.Bids[0].Price
	bidQty := bestBid.book.Bids[0].Qty
	askPrice := bestAsk.book.Asks[0].Price
	askQty := bestAsk.book.Asks[0].Qty

	buyFee := bestAsk.book.Asks[0].Price * 0.0025
	sellFee := bestBid.book.Bids[0].Price * 0.0025
	tfFee := 10.0

	if bestBid.book.Bids[0].Price > 0 && bestAsk.book.Asks[0].Price > 0 {
		fmt.Printf("Best Buy\t@ %s\t%.2f THB\t%.4f BNB\n", bestAsk.exchange, askPrice, askQty)
		fmt.Printf("Best Sell\t@ %s\t%.2f THB\t%.4f BNB\n", bestBid.exchange, bidPrice, bidQty)
		fmt.Printf("Profit\t\t= %.2f THB\t(Spread = %.2f , Fee = %.2f + %.2f + %.2f)\n",
			(bestBid.book.Bids[0].Price - bestAsk.book.Asks[0].Price - buyFee - sellFee - tfFee),
			(bestBid.book.Bids[0].Price - bestAsk.book.Asks[0].Price), buyFee, sellFee, tfFee)
	}
}
