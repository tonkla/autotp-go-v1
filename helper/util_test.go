package helper

import (
	"testing"

	"github.com/tonkla/autotp/types"
)

func TestCalcSLStop(t *testing.T) {
	var gap float64 = 500

	side := "BUY"
	if CalcSLStop(side, 5, gap, 5) != 5.005 {
		t.Fail()
	}
	if CalcSLStop(side, 50, gap, 3) != 50.5 {
		t.Fail()
	}
	if CalcSLStop(side, 500, gap, 2) != 505 {
		t.Fail()
	}
	if CalcSLStop(side, 5000, gap, 2) != 5005 {
		t.Fail()
	}

	side = "SELL"
	if CalcSLStop(side, 5, gap, 5) != 4.995 {
		t.Fail()
	}
	if CalcSLStop(side, 50, gap, 3) != 49.5 {
		t.Fail()
	}
	if CalcSLStop(side, 500, gap, 2) != 495 {
		t.Fail()
	}
	if CalcSLStop(side, 5000, gap, 2) != 4995 {
		t.Fail()
	}
}

func TestCalcTPStop(t *testing.T) {
	var gap float64 = 500

	side := "BUY"
	if CalcTPStop(side, 5, gap, 5) != 4.995 {
		t.Fail()
	}
	if CalcTPStop(side, 50, gap, 3) != 49.5 {
		t.Fail()
	}
	if CalcTPStop(side, 500, gap, 2) != 495 {
		t.Fail()
	}
	if CalcTPStop(side, 5000, gap, 2) != 4995 {
		t.Fail()
	}

	side = "SELL"
	if CalcTPStop(side, 5, gap, 5) != 5.005 {
		t.Fail()
	}
	if CalcTPStop(side, 50, gap, 3) != 50.5 {
		t.Fail()
	}
	if CalcTPStop(side, 500, gap, 2) != 505 {
		t.Fail()
	}
	if CalcTPStop(side, 5000, gap, 2) != 5005 {
		t.Fail()
	}
}

func TestReverse(t *testing.T) {
	if Reverse(types.OrderSideBuy) != types.OrderSideSell ||
		Reverse(types.OrderSideSell) != types.OrderSideBuy {
		t.Fail()
	}
}

func TestNormalizeDouble(t *testing.T) {
	type test struct {
		number   float64
		digits   int64
		expected float64
	}

	data := []test{
		{0.12345, 1, 0.1},
		{0.12345, 2, 0.12},
		{0.12345, 3, 0.123},
		{0.12344, 4, 0.1234},
		{0.12345, 4, 0.1235},
		{0.12345, 5, 0.12345},
		{1.001, 2, 1.0},
		{1.504, 2, 1.5},
		{1.505, 2, 1.51},
		{1.999, 2, 2.0},
	}

	for _, d := range data {
		if NormalizeDouble(d.number, d.digits) != d.expected {
			t.Fail()
		}
	}
}
