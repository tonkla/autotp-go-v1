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
