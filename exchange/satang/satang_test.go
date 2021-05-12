package satang

import (
	"testing"
)

const (
	name   = "SATANG"
	symbol = "bnb_thb"
)

func TestGetName(t *testing.T) {
	ex := New()
	if ex.GetName() != name {
		t.Fail()
	}
}

func TestGetTicker(t *testing.T) {
	ex := New()
	ticker := ex.GetTicker(symbol)
	if ticker.Price == 0 || ticker.Qty == 0 {
		t.Fail()
	}
}

func TestGetHistoricalPrices(t *testing.T) {
	ex := New()
	prices := ex.GetHistoricalPrices(symbol, "1d", 1)
	if len(prices) != 1 || prices[0].Open == 0 {
		t.Fail()
	}
}

func TestGetOrderBook(t *testing.T) {
	ex := New()
	book := ex.GetOrderBook(symbol, 5)
	if len(book.Bids) != 5 || len(book.Asks) != 5 || book.Bids[0].Price == 0 || book.Asks[0].Price == 0 {
		t.Fail()
	}
}
