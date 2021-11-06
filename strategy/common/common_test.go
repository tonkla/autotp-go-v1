package common

import (
	"testing"

	binance "github.com/tonkla/autotp/exchange/binance/spot"
	"github.com/tonkla/autotp/talib"
	"github.com/tonkla/autotp/types"
)

const (
	symbol    = "BNBBUSD"
	timeframe = "1d"
	period    = 8
)

func TestGetTrend(t *testing.T) {
	client := binance.NewSpotClient("", "")
	bars := client.GetHistoricalPrices(symbol, timeframe, 20)
	trend := GetTrend(bars, period)
	t.Error(trend)
}

func TestGetATR(t *testing.T) {
	client := binance.NewSpotClient("", "")
	bars := client.GetHistoricalPrices(symbol, timeframe, 100)

	var h, l, c []float64
	for _, b := range bars {
		h = append(h, b.High)
		l = append(l, b.Low)
		c = append(c, b.Close)
	}

	r := talib.ATR(h, l, c, period)
	atr1 := r[len(r)-1]
	t.Error("TALib ATR:\t", atr1)

	atr2 := GetATR(bars, period)
	t.Error("Custom ATR:\t", atr2)

	t.Error("Skip")
}

func TestGetGridRange(t *testing.T) {
	// 60%5=0, width=300/60=5
	lower, upper, grid := GetGridRange(554, 500, 800, 60)
	if lower != 550 || upper != 555 || grid != 5 {
		t.Error("550-554-555")
	}
	lower, upper, _ = GetGridRange(555, 500, 800, 60)
	if lower != 550 || upper != 560 {
		t.Error("550-555-560")
	}
	lower, upper, _ = GetGridRange(556, 500, 800, 60)
	if lower != 555 || upper != 560 {
		t.Error("555-556-560")
	}

	// 10%5=0, width=100/10=10
	lower, upper, _ = GetGridRange(22, 10, 110, 10)
	if lower != 20 || upper != 30 {
		t.Error("20-22-30")
	}

	// 24%4=0, width=192/24=8
	lower, upper, _ = GetGridRange(164, 10, 202, 24)
	if lower != 162 || upper != 170 {
		t.Error("162-164-170")
	}

	// 18%3=0, width=126/18=7
	lower, upper, _ = GetGridRange(71, 10, 136, 18)
	if lower != 66 || upper != 73 {
		t.Error("66-71-73")
	}

	// 14%2=0, width=84/14=6
	lower, upper, _ = GetGridRange(90, 10, 94, 14)
	if lower != 88 || upper != 94 {
		t.Error("88-90-94")
	}
}

func TestGetGridZones(t *testing.T) {
	zones, grid := GetGridZones(554, 500, 800, 60)
	if len(zones) != (800-550)/int(grid) || grid != 5 {
		t.Error("550-554-555")
	}
	zones, _ = GetGridZones(555, 500, 800, 60)
	if len(zones) != (800-550)/int(grid) {
		t.Error("550-555-560")
	}
	zones, _ = GetGridZones(556, 500, 800, 60)
	if len(zones) != (800-555)/int(grid) {
		t.Error("555-556-560")
	}
}

func TestGetPercentHL(t *testing.T) {
	type expect struct {
		i types.Ticker
		o float64
	}

	prices := []types.HistoricalPrice{
		{Open: 50, High: 110, Low: 10},
		{Open: 50, High: 110, Low: 10},
	}

	data := []expect{
		{i: types.Ticker{Price: 10}, o: 0},
		{i: types.Ticker{Price: 50}, o: 0.4},
		{i: types.Ticker{Price: 60}, o: 0.5},
		{i: types.Ticker{Price: 108.5}, o: 0.985},
		{i: types.Ticker{Price: 109}, o: 0.99},
		{i: types.Ticker{Price: 110}, o: 1},
	}

	for _, d := range data {
		r := GetPercentHL(prices, d.i)
		if r == nil {
			t.Error()
		} else {
			if d.o != *r {
				t.Errorf("Expect: %f, Got: %f", d.o, *r)
			}
		}
	}
}

func TestGetPercentHLTicker(t *testing.T) {
	client := binance.NewSpotClient("", "")
	prices4h := client.GetHistoricalPrices(symbol, "4h", 8)
	prices1h := client.GetHistoricalPrices(symbol, "1h", 8)
	prices15m := client.GetHistoricalPrices(symbol, "15m", 8)
	ticker := client.GetTicker(symbol)
	if ticker == nil {
		return
	}
	p4h := GetPercentHL(prices4h, *ticker)
	if p4h != nil {
		t.Errorf(" 4H: %f", *p4h)
	}
	p1h := GetPercentHL(prices1h, *ticker)
	if p1h != nil {
		t.Errorf(" 1H: %f", *p1h)
	}
	p15m := GetPercentHL(prices15m, *ticker)
	if p15m != nil {
		t.Errorf("15M: %f", *p15m)
	}
}

func TestGetHighsLows(t *testing.T) {
	client := binance.NewSpotClient("", "")
	prices := client.GetHistoricalPrices(symbol, timeframe, 20)
	h, l := GetHighsLows(prices)
	t.Errorf("H0=%f, L0=%f", h[len(h)-1], l[len(l)-1])
}
