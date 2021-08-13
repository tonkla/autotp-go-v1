package binance

import (
	"testing"
)

const ssymbol = "BNBBUSD"

var sc = NewSpotClient("", "")

func TestSpotGetExchangeInfo(t *testing.T) {
	err := sc.GetExchangeInfo(ssymbol)
	if err != nil {
		t.Error(err)
	}
}

func TestSpotGetTicker(t *testing.T) {
	ticker := sc.GetTicker(ssymbol)
	if ticker == nil || ticker.Price <= 0 {
		t.Fail()
	}
}

func TestSpotGetOrderBook(t *testing.T) {
	book := sc.GetOrderBook(ssymbol, 5)
	if book == nil || len(book.Asks) != 5 || len(book.Bids) != 5 ||
		book.Asks[0].Price <= 0 || book.Bids[0].Price <= 0 {
		t.Fail()
	}
}

func TestSpotGetHistoricalPrices(t *testing.T) {
	prices := sc.GetHistoricalPrices(ssymbol, "1d", 10)
	if len(prices) == 0 || len(prices) != 10 || prices[0].Open == 0 {
		t.Fail()
	}
}
