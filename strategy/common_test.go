package strategy

import (
	"fmt"
	"testing"

	binance "github.com/tonkla/autotp/exchange/binance/fusd"
	"github.com/tonkla/autotp/talib"
)

const (
	symbol    = "BNBUSDT"
	timeframe = "1d"
	period    = 8
)

func TestGetTrend(t *testing.T) {
	bars := binance.GetHistoricalPrices(symbol, timeframe, 20)
	trend := GetTrend(bars, period)
	fmt.Println(trend)
	t.Error("Skip")
}

func TestGetATR(t *testing.T) {
	bars := binance.GetHistoricalPrices(symbol, timeframe, 100)

	var h, l, c []float64
	for _, b := range bars {
		h = append(h, b.High)
		l = append(l, b.Low)
		c = append(c, b.Close)
	}

	r := talib.ATR(h, l, c, period)
	atr1 := r[len(r)-1]
	fmt.Println("TALib ATR:\t", atr1)

	atr2 := GetATR(bars, period)
	fmt.Println("Custom ATR:\t", atr2)

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
