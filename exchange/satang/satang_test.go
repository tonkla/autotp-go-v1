package satang

import (
	"testing"
)

const symbol = "bnb_thb"

var c = NewClient()

func TestGetTicker(t *testing.T) {
	ticker := c.GetTicker(symbol)
	if ticker.Price == 0 || ticker.Qty == 0 {
		t.Fail()
	}
}

func TestGetHistoricalPrices(t *testing.T) {
	prices := c.GetHistoricalPrices(symbol, "1d", 1)
	if len(prices) != 1 || prices[0].Open == 0 {
		t.Fail()
	}
}

func TestGetOrderBook(t *testing.T) {
	book := c.GetOrderBook(symbol, 5)
	if len(book.Bids) != 5 || len(book.Asks) != 5 || book.Bids[0].Price == 0 || book.Asks[0].Price == 0 {
		t.Fail()
	}
}
