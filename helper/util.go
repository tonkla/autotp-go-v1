package helper

import (
	"time"

	t "github.com/tonkla/autotp/types"
)

// Now13 returns a millisecond Unix timestamp (13 digits)
func Now13() int64 {
	return time.Now().UnixNano() / 1e6
}

// CalSLStop calculates a stop price of the stop loss price
func CalSLStop(side string, sl float64, trigger float64) float64 {
	if trigger == 0 {
		trigger = 0.002
	}
	if side == t.OrderSideBuy {
		return sl + sl*trigger
	}
	return sl - sl*trigger
}

// CalTPStop calculates a stop price of the take profit price
func CalTPStop(side string, tp float64, trigger float64) float64 {
	if trigger == 0 {
		trigger = 0.002
	}
	if side == t.OrderSideBuy {
		return tp - tp*trigger
	}
	return tp + tp*trigger
}

// Reverse reverses an order side to opposite
func Reverse(side string) string {
	if side == t.OrderSideBuy {
		return t.OrderSideSell
	}
	return t.OrderSideBuy
}
