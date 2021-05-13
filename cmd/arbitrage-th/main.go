package main

import (
	"fmt"
	"sync"

	"github.com/tonkla/autotp/exchange/bitkub"
	"github.com/tonkla/autotp/exchange/satang"
	"github.com/tonkla/autotp/types"
)

func main() {
	var wg sync.WaitGroup
	ch := make(chan types.OrderBook, 2)

	wg.Add(2)

	go func() {
		wg.Wait()
		close(ch)
	}()

	go func() {
		defer wg.Done()
		ex := bitkub.New()
		ch <- ex.GetOrderBook("thb_bnb", 5)
	}()

	go func() {
		defer wg.Done()
		ex := satang.New()
		ch <- ex.GetOrderBook("bnb_thb", 5)
	}()

	var bitkub types.OrderBook
	var satang types.OrderBook

	for book := range ch {
		if book.Exchange.Name == types.EXC_BITKUB {
			bitkub = book
		} else if book.Exchange.Name == types.EXC_SATANG {
			satang = book
		}
	}

	bestBid := bitkub
	if satang.Bids[0].Price > bitkub.Bids[0].Price {
		bestBid = satang
	}

	bestAsk := bitkub
	if satang.Asks[0].Price < bitkub.Asks[0].Price {
		bestAsk = satang
	}

	bidPrice := bestBid.Bids[0].Price
	bidQty := bestBid.Bids[0].Qty
	askPrice := bestAsk.Asks[0].Price
	askQty := bestAsk.Asks[0].Qty

	buyFee := bestAsk.Asks[0].Price * 0.0025
	sellFee := bestBid.Bids[0].Price * 0.0025
	tfFee := 10.0

	if bestBid.Bids[0].Price > 0 && bestAsk.Asks[0].Price > 0 {
		fmt.Printf("Best Buy\t@ %s\t%.2f THB\t%.4f BNB\n", bestAsk.Exchange.Name, askPrice, askQty)
		fmt.Printf("Best Sell\t@ %s\t%.2f THB\t%.4f BNB\n", bestBid.Exchange.Name, bidPrice, bidQty)
		fmt.Printf("Profit\t\t= %.2f THB\t(Spread = %.2f , Fee = %.2f + %.2f + %.2f)\n",
			(bestBid.Bids[0].Price - bestAsk.Asks[0].Price - buyFee - sellFee - tfFee),
			(bestBid.Bids[0].Price - bestAsk.Asks[0].Price), buyFee, sellFee, tfFee)
	}
}
