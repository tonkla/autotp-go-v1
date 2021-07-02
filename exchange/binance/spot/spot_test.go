package spot

import (
	"testing"
)

const symbol = "BNBUSDT"

func TestGetTicker(t *testing.T) {
	ticker := GetTicker(symbol)
	if ticker == nil || ticker.Price <= 0 {
		t.Fail()
	}
}

func TestGetOrderBook(t *testing.T) {
	book := GetOrderBook(symbol, 5)
	if book == nil || len(book.Asks) != 5 || len(book.Bids) != 5 ||
		book.Asks[0].Price <= 0 || book.Bids[0].Price <= 0 {
		t.Fail()
	}
}

func TestGetHistoricalPrices(t *testing.T) {
	prices := GetHistoricalPrices(symbol, "1d", 10)
	if len(prices) == 0 || len(prices) != 10 || prices[0].Open == 0 {
		t.Fail()
	}
}
