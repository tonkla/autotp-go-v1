package helper

import "testing"

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
