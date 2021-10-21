package common

import (
	"math"

	"github.com/tonkla/autotp/rdb"
	"github.com/tonkla/autotp/talib"
	t "github.com/tonkla/autotp/types"
)

type OnTickParams struct {
	DB rdb.DB
	TK t.Ticker
	BP t.BotParams
}

// IsDown returns true when the Close price is lower than the Open price
func IsDown(hprices t.HistoricalPrice) bool {
	return hprices.Close < hprices.Open
}

// GetTrend returns a stupid trend, do not trust him
func GetTrend(hprices []t.HistoricalPrice, period int) int {
	trend := t.TrendNo

	if len(hprices) < 10 || hprices[len(hprices)-1].Open == 0 || period <= 0 {
		return trend
	}

	p_0 := hprices[len(hprices)-1]
	p_1 := hprices[len(hprices)-2]

	o_0 := p_0.Open

	h_0 := p_0.High
	h_1 := p_1.High

	l_0 := p_0.Low
	l_1 := p_1.Low

	c_0 := p_0.Close

	var h, l, c []float64
	for _, p := range hprices {
		h = append(h, p.High)
		l = append(l, p.Low)
		c = append(c, p.Close)
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
		trend = t.TrendUp1
		// Higher low, and continued positive slope
		if l_1 < l_0 && cma_2 < cma_1 {
			trend = t.TrendUp2
			// Green bar, or moving to top
			if o_0 < c_0 || h_0-c_0 < (c_0-l_0)*0.5 {
				trend = t.TrendUp3
				// Low is greater than average close, or long green bar, or narrow upper band
				if l_0 > cma_0 || h_0-l_0 > atr || hma_0-cma_0 < (cma_0-lma_0)*0.6 {
					trend = t.TrendUp4
					// Low is greater than average high, or very long green bar
					if l_0 > hma_0 || h_0-l_0 > 1.25*atr {
						trend = t.TrendUp5
					}
				}
			}
		}
	}
	// Negative slope
	if cma_1 > cma_0 {
		trend = t.TrendDown1
		// Lower high, and continued negative slope
		if h_1 > h_0 && cma_2 > cma_1 {
			trend = t.TrendDown2
			// Red bar, or moving to bottom
			if o_0 > c_0 || (h_0-c_0)*0.5 > c_0-l_0 {
				trend = t.TrendDown3
				// High is less than average close, or long red bar, or narrow lower band
				if h_0 < cma_0 || h_0-l_0 > atr || (hma_0-cma_0)*0.6 > cma_0-lma_0 {
					trend = t.TrendDown4
					// High is less than average low, or very long red bar
					if h_0 < lma_0 || h_0-l_0 > 1.25*atr {
						trend = t.TrendDown5
					}
				}
			}
		}
	}
	return trend
}

// GetATR returns @tonkla's ATR, that is not the "J. Welles Wilder Jr."'s ATR :-P
func GetATR(hprices []t.HistoricalPrice, period int) float64 {
	var h, l []float64
	for _, p := range hprices {
		h = append(h, p.High)
		l = append(l, p.Low)
	}
	hwma := talib.WMA(h, period)
	hma_0 := hwma[len(hwma)-1]
	lwma := talib.WMA(l, period)
	lma_0 := lwma[len(lwma)-1]
	return hma_0 - lma_0
}

// GetGridRange returns the lower number and the upper number that closed to the target number
func GetGridRange(target float64, lowerNum float64, upperNum float64, gridSize float64) (float64, float64, float64) {
	if target <= lowerNum || lowerNum >= upperNum || gridSize < 2 {
		return 0, 0, 0
	}

	lower := lowerNum
	upper := upperNum
	zone := upper - lower
	gridWidth := zone / gridSize

	if math.Mod(gridSize, 5) == 0 {
		div := zone / 5

		if lower+div*4 < target {
			lower += div * 4
		} else if lower+div*3 < target {
			lower += div * 3
		} else if lower+div*2 < target {
			lower += div * 2
		} else if lower+div < target {
			lower += div
		}

		if upper-div*4 > target {
			upper -= div * 4
		} else if upper-div*3 > target {
			upper -= div * 3
		} else if upper-div*2 > target {
			upper -= div * 2
		} else if upper-div > target {
			upper -= div
		}
	} else if math.Mod(gridSize, 4) == 0 {
		div := zone / 4

		if lower+div*3 < target {
			lower += div * 3
		} else if lower+div*2 < target {
			lower += div * 2
		} else if lower+div < target {
			lower += div
		}

		if upper-div*3 > target {
			upper -= div * 3
		} else if upper-div*2 > target {
			upper -= div * 2
		} else if upper-div > target {
			upper -= div
		}
	} else if math.Mod(gridSize, 3) == 0 {
		div := zone / 3

		if lower+div*2 < target {
			lower += div * 2
		} else if lower+div < target {
			lower += div
		}

		if upper-div*2 > target {
			upper -= div * 2
		} else if upper-div > target {
			upper -= div
		}
	} else if math.Mod(gridSize, 2) == 0 {
		div := zone / 2

		if lower+div < target {
			lower += div
		}

		if upper-div > target {
			upper -= div
		}
	}

	for i := 0; i < int(gridSize); i++ {
		if lower+gridWidth < target {
			lower += gridWidth
		} else {
			break
		}
	}

	for i := 0; i < int(gridSize); i++ {
		if upper-gridWidth > target {
			upper -= gridWidth
		} else {
			break
		}
	}

	return lower, upper, gridWidth
}

// GetGridZones returns all buyable zones of the grid
func GetGridZones(target float64, lowerNum float64, upperNum float64, grids float64) ([]float64, float64) {
	if target <= lowerNum || lowerNum >= upperNum || grids < 2 {
		return nil, 0
	}

	start, _, gridWidth := GetGridRange(target, lowerNum, upperNum, grids)

	var zones []float64
	for i := 0.0; i < grids; i++ {
		num := start + i*gridWidth
		if num >= upperNum {
			break
		}
		zones = append(zones, num)
	}
	return zones, gridWidth
}
