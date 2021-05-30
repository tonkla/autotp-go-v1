/*
Copyright 2016 Mark Chenoweth
Licensed under terms of MIT license (see LICENSE)
https://github.com/markcheno/go-talib/blob/master/talib.go
*/

package talib

import "math"

func SMA(inReal []float64, inTimePeriod int) []float64 {
	outReal := make([]float64, len(inReal))

	lookbackTotal := inTimePeriod - 1
	startIdx := lookbackTotal
	periodTotal := 0.0
	trailingIdx := startIdx - lookbackTotal
	i := trailingIdx
	if inTimePeriod > 1 {
		for i < startIdx {
			periodTotal += inReal[i]
			i++
		}
	}
	outIdx := startIdx
	for ok := true; ok; {
		periodTotal += inReal[i]
		tempReal := periodTotal
		periodTotal -= inReal[trailingIdx]
		outReal[outIdx] = tempReal / float64(inTimePeriod)
		trailingIdx++
		i++
		outIdx++
		ok = i < len(outReal)
	}
	return outReal
}

func EMA(inReal []float64, inTimePeriod int) []float64 {
	multiplier := 2.0 / float64(inTimePeriod+1)
	return EMAk(inReal, inTimePeriod, multiplier)
}

func EMAk(inReal []float64, inTimePeriod int, multiplier float64) []float64 {
	outReal := make([]float64, len(inReal))

	lookbackTotal := inTimePeriod - 1
	startIdx := lookbackTotal
	today := startIdx - lookbackTotal
	i := inTimePeriod
	tempReal := 0.0
	for i > 0 {
		tempReal += inReal[today]
		today++
		i--
	}

	prevMA := tempReal / float64(inTimePeriod)
	for today <= startIdx {
		prevMA = ((inReal[today] - prevMA) * multiplier) + prevMA
		today++
	}
	outReal[startIdx] = prevMA
	outIdx := startIdx + 1
	for today < len(inReal) {
		prevMA = ((inReal[today] - prevMA) * multiplier) + prevMA
		outReal[outIdx] = prevMA
		today++
		outIdx++
	}
	return outReal
}

func WMA(inReal []float64, inTimePeriod int) []float64 {
	outReal := make([]float64, len(inReal))

	lookbackTotal := inTimePeriod - 1
	startIdx := lookbackTotal

	if inTimePeriod == 1 {
		copy(outReal, inReal)
		return outReal
	}
	divider := (inTimePeriod * (inTimePeriod + 1)) >> 1
	outIdx := inTimePeriod - 1
	trailingIdx := startIdx - lookbackTotal
	periodSum, periodSub := 0.0, 0.0
	inIdx := trailingIdx
	i := 1
	for inIdx < startIdx {
		tempReal := inReal[inIdx]
		periodSub += tempReal
		periodSum += tempReal * float64(i)
		inIdx++
		i++
	}
	trailingValue := 0.0
	for inIdx < len(inReal) {
		tempReal := inReal[inIdx]
		periodSub += tempReal
		periodSub -= trailingValue
		periodSum += tempReal * float64(inTimePeriod)
		trailingValue = inReal[trailingIdx]
		outReal[outIdx] = periodSum / float64(divider)
		periodSum -= periodSub
		inIdx++
		trailingIdx++
		outIdx++
	}
	return outReal
}

func MACD(inReal []float64, inFastPeriod int, inSlowPeriod int, inSignalPeriod int) ([]float64, []float64, []float64) {
	if inSlowPeriod < inFastPeriod {
		inSlowPeriod, inFastPeriod = inFastPeriod, inSlowPeriod
	}

	mFast := 0.0
	if inFastPeriod != 0 {
		mFast = 2.0 / float64(inFastPeriod+1)
	} else {
		inFastPeriod = 12
		mFast = 0.15
	}

	mSlow := 0.0
	if inSlowPeriod != 0 {
		mSlow = 2.0 / float64(inSlowPeriod+1)
	} else {
		inSlowPeriod = 26
		mSlow = 0.075
	}

	lookbackSignal := inSignalPeriod - 1
	lookbackTotal := lookbackSignal
	lookbackTotal += (inSlowPeriod - 1)

	fastEMABuffer := EMAk(inReal, inFastPeriod, mFast)
	slowEMABuffer := EMAk(inReal, inSlowPeriod, mSlow)
	for i := 0; i < len(fastEMABuffer); i++ {
		fastEMABuffer[i] = fastEMABuffer[i] - slowEMABuffer[i]
	}

	outMACD := make([]float64, len(inReal))
	for i := lookbackTotal - 1; i < len(fastEMABuffer); i++ {
		outMACD[i] = fastEMABuffer[i]
	}
	outMACDSignal := EMAk(outMACD, inSignalPeriod, (2.0 / float64(inSignalPeriod+1)))

	outMACDHist := make([]float64, len(inReal))
	for i := lookbackTotal; i < len(outMACDHist); i++ {
		outMACDHist[i] = outMACD[i] - outMACDSignal[i]
	}
	return outMACD, outMACDSignal, outMACDHist
}

func ATR(inHigh []float64, inLow []float64, inClose []float64, inTimePeriod int) []float64 {
	outReal := make([]float64, len(inClose))

	inTimePeriodF := float64(inTimePeriod)

	if inTimePeriod < 1 {
		return outReal
	}

	if inTimePeriod <= 1 {
		return TRange(inHigh, inLow, inClose)
	}

	outIdx := inTimePeriod
	today := inTimePeriod + 1

	tr := TRange(inHigh, inLow, inClose)
	prevATRTemp := SMA(tr, inTimePeriod)
	prevATR := prevATRTemp[inTimePeriod]
	outReal[inTimePeriod] = prevATR

	for outIdx = inTimePeriod + 1; outIdx < len(inClose); outIdx++ {
		prevATR *= inTimePeriodF - 1.0
		prevATR += tr[today]
		prevATR /= inTimePeriodF
		outReal[outIdx] = prevATR
		today++
	}
	return outReal
}

func TRange(inHigh []float64, inLow []float64, inClose []float64) []float64 {
	outReal := make([]float64, len(inClose))

	startIdx := 1
	outIdx := startIdx
	today := startIdx
	for today < len(inClose) {
		tempLT := inLow[today]
		tempHT := inHigh[today]
		tempCY := inClose[today-1]
		greatest := tempHT - tempLT
		val2 := math.Abs(tempCY - tempHT)
		if val2 > greatest {
			greatest = val2
		}
		val3 := math.Abs(tempCY - tempLT)
		if val3 > greatest {
			greatest = val3
		}
		outReal[outIdx] = greatest
		outIdx++
		today++
	}
	return outReal
}
