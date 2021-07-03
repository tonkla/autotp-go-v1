package bitkub

import (
	"testing"
)

const symbol = "THB_BNB"

var c = NewClient()

func TestGetTicker(t *testing.T) {
	ticker := c.GetTicker(symbol)
	if ticker.Price == 0 {
		t.Fail()
	}
}

func TestGetHistoricalPrices(t *testing.T) {
	prices := c.GetHistoricalPrices(symbol, 86400)
	if len(prices) != 1 || prices[0].Open == 0 {
		t.Fail()
	}
}

func TestGetOrderBook(t *testing.T) {
	book := c.GetOrderBook(symbol, 5)
	if len(book.Bids) != 5 || len(book.Asks) != 5 || book.Bids[0].Price == 0.0 || book.Asks[0].Price == 0 {
		t.Error()
	}
}
