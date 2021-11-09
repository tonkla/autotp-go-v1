package scalping

import (
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
			CloseOrders:  closeOrders,
			CancelOrders: cancelOrders,
		}
	}

	closeOrders = append(closeOrders, common.CloseOpposite(s.DB, s.BP, qo, ticker)...)
	if len(closeOrders) > 0 {
		if closeOrders[0].Side == t.OrderSideBuy {
			cancelOrders = append(cancelOrders, s.DB.GetNewStopLongOrders(qo)...)
			cancelOrders = append(cancelOrders, s.DB.GetNewLimitLongOrders(qo)...)
		} else {
			cancelOrders = append(cancelOrders, s.DB.GetNewStopShortOrders(qo)...)
			cancelOrders = append(cancelOrders, s.DB.GetNewLimitShortOrders(qo)...)
		}
		return &t.TradeOrders{
			CloseOrders:  closeOrders,
			CancelOrders: cancelOrders,
		}
	}

	const numberOfBars = 50

	prices2nd := s.EX.GetHistoricalPrices(s.BP.Symbol, s.BP.MATf2nd, numberOfBars)
	if len(prices2nd) < numberOfBars || prices2nd[len(prices2nd)-1].Open == 0 || prices2nd[len(prices2nd)-2].Open == 0 {
		return nil
	}

	prices := s.EX.GetHistoricalPrices(s.BP.Symbol, s.BP.MATf3rd, numberOfBars)
	if len(prices) < numberOfBars || prices[len(prices)-1].Open == 0 || prices[len(prices)-2].Open == 0 {
		return nil
	}

	highs, lows := common.GetHighsLows(prices)
	hma := talib.WMA(highs, int(s.BP.MAPeriod3rd))
	hma_0 := hma[len(hma)-1]
	hma_1 := hma[len(hma)-2]
	lma := talib.WMA(lows, int(s.BP.MAPeriod3rd))
	lma_0 := lma[len(lma)-1]
	lma_1 := lma[len(lma)-2]

	atr := hma_0 - lma_0

	qo.Qty = h.NormalizeDouble(s.BP.BaseQty, s.BP.QtyDigits)
	qty := h.NormalizeDouble(s.BP.QuoteQty/ticker.Price, s.BP.QtyDigits)
	if qty > qo.Qty {
		qo.Qty = qty
	}

	if s.BP.AutoSL {
		closeOrders = append(closeOrders, common.SLLong(s.DB, s.BP, qo, ticker, atr)...)
		closeOrders = append(closeOrders, common.SLShort(s.DB, s.BP, qo, ticker, atr)...)
		closeOrders = append(closeOrders, common.TimeSL(s.DB, s.BP, qo, ticker)...)
	}

	if s.BP.AutoTP {
		closeOrders = append(closeOrders, common.TPLong(s.DB, s.BP, qo, ticker, atr)...)
		closeOrders = append(closeOrders, common.TPShort(s.DB, s.BP, qo, ticker, atr)...)
		closeOrders = append(closeOrders, common.TimeTP(s.DB, s.BP, qo, ticker)...)
	}

	p_0 := prices[len(prices)-1]
	t_0 := p_0.Time

	percent := common.GetPercentHL(prices, ticker)
	if percent == nil {
		return nil
	}

	closes2nd := common.GetCloses(prices2nd)
	cma2nd := talib.WMA(closes2nd, int(s.BP.MAPeriod2nd))
	cma2nd_0 := cma2nd[len(cma2nd)-1]
	cma2nd_1 := cma2nd[len(cma2nd)-2]

	shouldCloseLong := cma2nd_1 > cma2nd_0 && hma_1 > hma_0
	shouldCloseShort := cma2nd_1 < cma2nd_0 && lma_1 < lma_0

	if shouldCloseLong || shouldCloseShort {
		if shouldCloseLong {
			cancelOrders = append(cancelOrders, s.DB.GetNewLimitLongOrders(qo)...)
			closeOrders = append(closeOrders, common.CloseLong(s.DB, s.BP, qo, ticker)...)
		}
		if shouldCloseShort {
			cancelOrders = append(cancelOrders, s.DB.GetNewLimitShortOrders(qo)...)
			closeOrders = append(closeOrders, common.CloseShort(s.DB, s.BP, qo, ticker)...)
		}
		if len(cancelOrders) > 0 || len(closeOrders) > 0 {
			return &t.TradeOrders{
				CloseOrders:  closeOrders,
				CancelOrders: cancelOrders,
			}
		}
	}

	shouldOpenLong := cma2nd_1 < cma2nd_0 && lma_1 < lma_0 && *percent < 0.2 && ticker.Price < hma_0
	shouldOpenShort := cma2nd_1 > cma2nd_0 && hma_1 > hma_0 && *percent > 0.8 && ticker.Price > lma_0

	if shouldOpenLong && shouldOpenShort {
		return &t.TradeOrders{
			CloseOrders:  closeOrders,
			CancelOrders: cancelOrders,
		}
	}

	if shouldOpenLong && (s.BP.View == t.ViewNeutral || s.BP.View == t.ViewLong) {
		openPrice := h.CalcStopLowerTicker(ticker.Price, float64(s.BP.Gap.OpenLimit), s.BP.PriceDigits)
		qo.OpenPrice = openPrice
		qo.Side = t.OrderSideBuy
		norder := s.DB.GetNearestOrder(qo)
		if norder == nil || norder.OpenPrice-openPrice >= s.BP.OrderGap {
			o := t.Order{
				ID:        h.GenID(),
				BotID:     s.BP.BotID,
				Exchange:  qo.Exchange,
				Symbol:    qo.Symbol,
				Side:      t.OrderSideBuy,
				Type:      t.OrderTypeLimit,
				Status:    t.OrderStatusNew,
				Qty:       qo.Qty,
				OpenPrice: openPrice,
			}
			if s.BP.Product == t.ProductFutures {
				o.PosSide = t.OrderPosSideLong
			}
			openOrders = append(openOrders, o)
		} else if norder.Status == t.OrderStatusNew && norder.OpenTime < t_0 {
			cancelOrders = append(cancelOrders, *norder)
		}
	}

	if shouldOpenShort && (s.BP.View == t.ViewNeutral || s.BP.View == t.ViewShort) {
		openPrice := h.CalcStopUpperTicker(ticker.Price, float64(s.BP.Gap.OpenLimit), s.BP.PriceDigits)
		qo.OpenPrice = openPrice
		qo.Side = t.OrderSideSell
		norder := s.DB.GetNearestOrder(qo)
		if norder == nil || openPrice-norder.OpenPrice >= s.BP.OrderGap {
			o := t.Order{
				ID:        h.GenID(),
				BotID:     s.BP.BotID,
				Exchange:  qo.Exchange,
				Symbol:    qo.Symbol,
				Side:      t.OrderSideSell,
				Type:      t.OrderTypeLimit,
				Status:    t.OrderStatusNew,
				Qty:       qo.Qty,
				OpenPrice: openPrice,
			}
			if s.BP.Product == t.ProductFutures {
				o.PosSide = t.OrderPosSideShort
			}
			openOrders = append(openOrders, o)
		} else if norder.Status == t.OrderStatusNew && norder.OpenTime < t_0 {
			cancelOrders = append(cancelOrders, *norder)
		}
	}

	return &t.TradeOrders{
		OpenOrders:   openOrders,
		CloseOrders:  closeOrders,
		CancelOrders: cancelOrders,
	}
}
