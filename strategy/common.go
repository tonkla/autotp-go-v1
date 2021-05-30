package strategy

import (
	"github.com/tonkla/autotp/talib"
	"github.com/tonkla/autotp/types"
)

// GetTrend returns a stupid trend, do not trust him
func GetTrend(bars []types.HistoricalPrice, period int) int {
	trend := types.TREND_NO

	if len(bars) < 10 || bars[len(bars)-1].Open == 0 || period <= 0 {
		return trend
	}

	b_0 := bars[len(bars)-1]
	b_1 := bars[len(bars)-2]

	o_0 := b_0.Open

	h_0 := b_0.High
	h_1 := b_1.High

	l_0 := b_0.Low
	l_1 := b_1.Low

	c_0 := b_0.Close

	var h, l, c []float64
	for _, b := range bars {
		h = append(h, b.High)
		l = append(l, b.Low)
		c = append(c, b.Close)
	}

	hwma := talib.WMA(h, period)
	hma_0 := hwma[len(hwma)-1]

	lwma := talib.WMA(l, period)
	lma_0 := lwma[len(lwma)-1]

	cwma := talib.WMA(c, period)
	cma_0 := cwma[len(cwma)-1]
	cma_1 := cwma[len(cwma)-2]

	// Not the J. Welles Wilder Jr.'s ATR
	atr := hma_0 - lma_0

	// Positive slope
	if cma_1 < cma_0 {
		trend = types.TREND_UP_1
		// Higher low
		if l_1 < l_0 {
			trend = types.TREND_UP_2
			// Green bar
			if o_0 < c_0 {
				trend = types.TREND_UP_3
				// Current low is greater than the average close, or it's a long green bar
				if l_0 > cma_0 || h_0-l_0 > atr {
					trend = types.TREND_UP_4
					// Current low is greater than the average high, or it's a very long green bar
					if l_0 > hma_0 || h_0-l_0 > 1.5*atr {
						trend = types.TREND_UP_5
					}
				}
			}
		}
	}
	// Negative slope
	if cma_1 > cma_0 {
		trend = types.TREND_DOWN_1
		// Lower high
		if h_1 > h_0 {
			trend = types.TREND_DOWN_2
			// Red bar
			if o_0 > c_0 {
				trend = types.TREND_DOWN_3
				// Current high is less than the average close, or it's a long red bar
				if h_0 < cma_0 || h_0-l_0 > atr {
					trend = types.TREND_DOWN_4
					// Current high is less than the average low, or it's a very long red bar
					if h_0 < lma_0 || h_0-l_0 > 1.5*atr {
						trend = types.TREND_DOWN_5
					}
				}
			}
		}
	}

	return trend
}

// GetATR returns my ATR that is not the J. Welles Wilder Jr.'s ATR :-P
func GetATR(bars []types.HistoricalPrice, period int) float64 {
	var h, l []float64
	for _, b := range bars {
		h = append(h, b.High)
		l = append(l, b.Low)
	}
	hwma := talib.WMA(h, period)
	hma_0 := hwma[len(hwma)-1]
	lwma := talib.WMA(l, period)
	lma_0 := lwma[len(lwma)-1]
	return hma_0 - lma_0
}
