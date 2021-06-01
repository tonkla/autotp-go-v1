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
	cma_2 := cwma[len(cwma)-3]

	// Not the J. Welles Wilder Jr.'s ATR
	atr := hma_0 - lma_0

	// Positive slope
	if cma_1 < cma_0 {
		trend = types.TREND_UP_1
		// Higher low, and continued positive slope
		if l_1 < l_0 && cma_2 < cma_1 {
			trend = types.TREND_UP_2
			// Green bar, or moving to top
			if o_0 < c_0 || h_0-c_0 < (c_0-l_0)*0.5 {
				trend = types.TREND_UP_3
				// Low is greater than average close, or long green bar, or narrow upper band
				if l_0 > cma_0 || h_0-l_0 > atr || hma_0-cma_0 < (cma_0-lma_0)*0.7 {
					trend = types.TREND_UP_4
					// Low is greater than average high, or very long green bar
					if l_0 > hma_0 || h_0-l_0 > 1.25*atr {
						trend = types.TREND_UP_5
					}
				}
			}
		}
	}
	// Negative slope
	if cma_1 > cma_0 {
		trend = types.TREND_DOWN_1
		// Lower high, and continued negative slope
		if h_1 > h_0 && cma_2 > cma_1 {
			trend = types.TREND_DOWN_2
			// Red bar, or moving to bottom
			if o_0 > c_0 || (h_0-c_0)*0.5 > c_0-l_0 {
				trend = types.TREND_DOWN_3
				// High is less than average close, or long red bar, or narrow lower band
				if h_0 < cma_0 || h_0-l_0 > atr || (hma_0-cma_0)*0.7 > cma_0-lma_0 {
					trend = types.TREND_DOWN_4
					// High is less than average low, or very long red bar
					if h_0 < lma_0 || h_0-l_0 > 1.25*atr {
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
