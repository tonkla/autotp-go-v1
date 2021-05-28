package fusd

import "testing"

var symbol = "BNBUSDT"

func TestGetTicker(t *testing.T) {
	ticker := GetTicker(symbol)
	if ticker.Price == 0 {
		t.Fail()
	}
}

func TestGetHistoricalPrices(t *testing.T) {
	prices := GetHistoricalPrices(symbol, "1d", 10)
	if len(prices) == 0 || len(prices) != 10 || prices[0].Open == 0 {
		t.Fail()
	}
}
