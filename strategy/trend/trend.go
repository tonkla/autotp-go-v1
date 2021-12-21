package trend

import (
	"math"

	"github.com/tonkla/autotp/exchange"
	h "github.com/tonkla/autotp/helper"
	"github.com/tonkla/autotp/rdb"
	"github.com/tonkla/autotp/strategy/common"
	"github.com/tonkla/autotp/talib"
	t "github.com/tonkla/autotp/types"
)

type Strategy struct {
	DB *rdb.DB
	BP *t.BotParams
	EX exchange.Repository
}

func New(db *rdb.DB, bp *t.BotParams, ex exchange.Repository) Strategy {
	return Strategy{
		DB: db,
		BP: bp,
		EX: ex,
	}
}

func (s Strategy) OnTick(ticker t.Ticker) *t.TradeOrders {
	var openOrders, closeOrders, cancelOrders []t.Order

	qo := t.QueryOrder{
		BotID:    s.BP.BotID,
		Exchange: s.BP.Exchange,
		Symbol:   s.BP.Symbol,
	}

	if s.BP.CloseLong || s.BP.CloseShort {
		if s.BP.CloseLong {
			cancelOrders = append(cancelOrders, s.DB.GetNewStopLongOrders(qo)...)
			cancelOrders = append(cancelOrders, s.DB.GetNewLimitLongOrders(qo)...)
			closeOrders = append(closeOrders, common.CloseLong(s.DB, s.BP, qo, ticker)...)
		}
		if s.BP.CloseShort {
			cancelOrders = append(cancelOrders, s.DB.GetNewStopShortOrders(qo)...)
			cancelOrders = append(cancelOrders, s.DB.GetNewLimitShortOrders(qo)...)
			closeOrders = append(closeOrders, common.CloseShort(s.DB, s.BP, qo, ticker)...)
		}
		return &t.TradeOrders{
			CancelOrders: cancelOrders,
			CloseOrders:  closeOrders,
		}
	}

	const numberOfBars = 30

	prices1st := s.EX.GetHistoricalPrices(s.BP.Symbol, s.BP.MATf1st, numberOfBars)
	if len(prices1st) == 0 || prices1st[len(prices1st)-1].Open == 0 {
		return nil
	}

	prices2nd := s.EX.GetHistoricalPrices(s.BP.Symbol, s.BP.MATf2nd, numberOfBars)
	if len(prices2nd) == 0 || prices2nd[len(prices2nd)-1].Open == 0 {
		return nil
	}

	prices3rd := s.EX.GetHistoricalPrices(s.BP.Symbol, s.BP.MATf3rd, numberOfBars)
	if len(prices3rd) == 0 || prices3rd[len(prices3rd)-1].Open == 0 {
		return nil
	}

	highs1st, lows1st := common.GetHighsLows(prices1st)
	hma1st := talib.WMA(highs1st, int(s.BP.MAPeriod1st))
	hma1st_0 := hma1st[len(hma1st)-1]
	hma1st_1 := hma1st[len(hma1st)-2]
	lma1st := talib.WMA(lows1st, int(s.BP.MAPeriod1st))
	lma1st_0 := lma1st[len(lma1st)-1]
	lma1st_1 := lma1st[len(lma1st)-2]
	mma1st_0 := lma1st_0 + ((hma1st_0 - lma1st_0) / 2)
	mma1st_1 := lma1st_1 + ((hma1st_1 - lma1st_1) / 2)

	h1st_1 := highs1st[len(highs1st)-2]
	l1st_1 := lows1st[len(lows1st)-2]

	highs2nd, lows2nd := common.GetHighsLows(prices2nd)
	hma2nd := talib.WMA(highs2nd, int(s.BP.MAPeriod2nd))
	hma2nd_0 := hma2nd[len(hma2nd)-1]
	hma2nd_1 := hma2nd[len(hma2nd)-2]
	lma2nd := talib.WMA(lows2nd, int(s.BP.MAPeriod2nd))
	lma2nd_0 := lma2nd[len(lma2nd)-1]
	lma2nd_1 := lma2nd[len(lma2nd)-2]
	mma2nd_0 := lma2nd_0 + ((hma2nd_0 - lma2nd_0) / 2)
	mma2nd_1 := lma2nd_1 + ((hma2nd_1 - lma2nd_1) / 2)

	h2nd_1 := highs2nd[len(highs2nd)-2]
	l2nd_1 := lows2nd[len(lows2nd)-2]
	h2nd_2 := highs2nd[len(highs2nd)-3]
	l2nd_2 := lows2nd[len(lows2nd)-3]

	hh2nd := h2nd_1
	if h2nd_2 > hh2nd {
		hh2nd = h2nd_2
	}
	ll2nd := l2nd_1
	if l2nd_2 < ll2nd {
		ll2nd = l2nd_2
	}

	highs3rd, lows3rd, closes3rd := common.GetHighsLowsCloses(prices3rd)

	c3rd_1 := closes3rd[len(closes3rd)-2]
	c3rd_2 := closes3rd[len(closes3rd)-3]
	c3rd_3 := closes3rd[len(closes3rd)-4]

	cma3rd := talib.WMA(closes3rd, int(s.BP.MAPeriod3rd))
	cma3rd_0 := cma3rd[len(cma3rd)-1]
	cma3rd_1 := cma3rd[len(cma3rd)-2]

	hma3rd := talib.WMA(highs3rd, int(s.BP.MAPeriod3rd))
	hma3rd_0 := hma3rd[len(hma3rd)-1]
	hma3rd_2 := hma3rd[len(hma3rd)-3]

	lma3rd := talib.WMA(lows3rd, int(s.BP.MAPeriod3rd))
	lma3rd_0 := lma3rd[len(lma3rd)-1]
	lma3rd_2 := lma3rd[len(lma3rd)-3]

	atr3rd := hma3rd_0 - lma3rd_0

	qo.Qty = h.NormalizeDouble(s.BP.BaseQty, s.BP.QtyDigits)
	qty := h.NormalizeDouble(s.BP.QuoteQty/ticker.Price, s.BP.QtyDigits)
	if qty > qo.Qty {
		qo.Qty = qty
	}

	if s.BP.AutoSL {
		closeOrders = append(closeOrders, common.SLLong(s.DB, s.BP, qo, ticker, atr3rd)...)
		closeOrders = append(closeOrders, common.SLShort(s.DB, s.BP, qo, ticker, atr3rd)...)
		closeOrders = append(closeOrders, common.TimeSL(s.DB, s.BP, qo, ticker)...)
	}

	if s.BP.AutoTP {
		closeOrders = append(closeOrders, common.TPLong(s.DB, s.BP, qo, ticker, atr3rd)...)
		closeOrders = append(closeOrders, common.TPShort(s.DB, s.BP, qo, ticker, atr3rd)...)
		closeOrders = append(closeOrders, common.TimeTP(s.DB, s.BP, qo, ticker)...)

		if len(closeOrders) == 0 {
			t0 := prices3rd[len(prices3rd)-1].Time
			if c3rd_2 > hma3rd_2 && c3rd_2 > c3rd_3 && c3rd_2 > c3rd_1 && ticker.Price >= c3rd_1 {
				closeOrders = append(closeOrders, common.CloseProfitLong(s.DB, s.BP, qo, ticker, t0)...)
			}
			if c3rd_2 < lma3rd_2 && c3rd_2 < c3rd_3 && c3rd_2 < c3rd_1 && ticker.Price <= c3rd_1 {
				closeOrders = append(closeOrders, common.CloseProfitShort(s.DB, s.BP, qo, ticker, t0)...)
			}
		}
	}

	if len(closeOrders) > 0 {
		return &t.TradeOrders{
			CloseOrders: closeOrders,
		}
	}

	shouldCloseLong := mma1st_1 > mma1st_0 || l1st_1 > ticker.Price
	shouldCloseShort := mma1st_1 < mma1st_0 || h1st_1 < ticker.Price

	if shouldCloseLong && s.BP.ForceClose {
		cancelOrders = append(cancelOrders, s.DB.GetNewLimitLongOrders(qo)...)
		closeOrders = append(closeOrders, common.CloseLong(s.DB, s.BP, qo, ticker)...)
	}

	if shouldCloseShort && s.BP.ForceClose {
		cancelOrders = append(cancelOrders, s.DB.GetNewLimitShortOrders(qo)...)
		closeOrders = append(closeOrders, common.CloseShort(s.DB, s.BP, qo, ticker)...)
	}

	if len(cancelOrders) > 0 || len(closeOrders) > 0 {
		return &t.TradeOrders{
			CancelOrders: cancelOrders,
			CloseOrders:  closeOrders,
		}
	}

	shouldOpenLong := ll2nd < ticker.Price && mma2nd_1 < mma2nd_0 &&
		!shouldCloseLong && (s.BP.View == t.ViewNeutral || s.BP.View == t.ViewLong)
	shouldOpenShort := hh2nd > ticker.Price && mma2nd_1 > mma2nd_0 &&
		!shouldCloseShort && (s.BP.View == t.ViewNeutral || s.BP.View == t.ViewShort)

	if shouldOpenLong && shouldOpenShort {
		return nil
	}

	if shouldOpenLong {
		openPrice := h.CalcStopLowerTicker(ticker.Price, float64(s.BP.Gap.OpenLimit), s.BP.PriceDigits)
		if (cma3rd_1 < cma3rd_0 && openPrice < cma3rd_0-(s.BP.MoS*atr3rd)) || cma3rd_1 > cma3rd_0 {
			_qo := qo
			_qo.Side = t.OrderSideBuy
			_qo.OpenPrice = openPrice
			norder := s.DB.GetNearestOrder(_qo)
			if norder == nil || math.Abs(norder.OpenPrice-openPrice) >= s.BP.OrderGapATR*atr3rd {
				o := t.Order{
					ID:        h.GenID(),
					BotID:     s.BP.BotID,
					Exchange:  s.BP.Exchange,
					Symbol:    s.BP.Symbol,
					Side:      t.OrderSideBuy,
					PosSide:   t.OrderPosSideLong,
					Type:      t.OrderTypeLimit,
					Status:    t.OrderStatusNew,
					Qty:       qo.Qty,
					OpenPrice: openPrice,
				}
				openOrders = append(openOrders, o)
			}
		}
	}

	if shouldOpenShort {
		openPrice := h.CalcStopUpperTicker(ticker.Price, float64(s.BP.Gap.OpenLimit), s.BP.PriceDigits)
		if (cma3rd_1 > cma3rd_0 && openPrice > cma3rd_0+(s.BP.MoS*atr3rd)) || cma3rd_1 < cma3rd_0 {
			_qo := qo
			_qo.Side = t.OrderSideSell
			_qo.OpenPrice = openPrice
			norder := s.DB.GetNearestOrder(_qo)
			if norder == nil || math.Abs(openPrice-norder.OpenPrice) >= s.BP.OrderGapATR*atr3rd {
				o := t.Order{
					ID:        h.GenID(),
					BotID:     s.BP.BotID,
					Exchange:  s.BP.Exchange,
					Symbol:    s.BP.Symbol,
					Side:      t.OrderSideSell,
					PosSide:   t.OrderPosSideShort,
					Type:      t.OrderTypeLimit,
					Status:    t.OrderStatusNew,
					Qty:       qo.Qty,
					OpenPrice: openPrice,
				}
				openOrders = append(openOrders, o)
			}
		}
	}

	return &t.TradeOrders{
		OpenOrders: openOrders,
	}
}
