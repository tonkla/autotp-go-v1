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
			CloseOrders:  closeOrders,
			CancelOrders: cancelOrders,
		}
	}

	const numberOfBars = 50
	prices := s.EX.GetHistoricalPrices(s.BP.Symbol, s.BP.MATf1st, numberOfBars)

	if len(prices) == 0 || prices[len(prices)-1].Open == 0 {
		return nil
	}

	highs, lows := common.GetHighsLows(prices)
	hma := talib.WMA(highs, int(s.BP.MAPeriod1st))
	hma_0 := hma[len(hma)-1]
	hma_1 := hma[len(hma)-2]
	lma := talib.WMA(lows, int(s.BP.MAPeriod1st))
	lma_0 := lma[len(lma)-1]
	lma_1 := lma[len(lma)-2]

	atr := 0.0
	if s.BP.AtrSL > 0 || s.BP.AtrTP > 0 {
		atr = common.GetSimpleATR(prices, int(s.BP.MAPeriod1st))
	}

	qo.Qty = h.NormalizeDouble(s.BP.BaseQty, s.BP.QtyDigits)
	qty := h.NormalizeDouble(s.BP.QuoteQty/ticker.Price, s.BP.QtyDigits)
	if qty > qo.Qty {
		qo.Qty = qty
	}

	if s.BP.AutoSL {
		closeOrders = append(closeOrders, common.SLLong(s.DB, s.BP, qo, ticker, atr)...)
		closeOrders = append(closeOrders, common.SLShort(s.DB, s.BP, qo, ticker, atr)...)
	}

	if s.BP.AutoTP {
		closeOrders = append(closeOrders, common.TPLong(s.DB, s.BP, qo, ticker, atr)...)
		closeOrders = append(closeOrders, common.TPShort(s.DB, s.BP, qo, ticker, atr)...)
	}

	if len(closeOrders) > 0 {
		return &t.TradeOrders{
			CloseOrders: closeOrders,
		}
	}

	shouldCloseLong := hma_1 > hma_0 && lma_1 > lma_0 && s.BP.AutoSL
	shouldCloseShort := hma_1 < hma_0 && lma_1 < lma_0 && s.BP.AutoSL

	if shouldCloseLong {
		cancelOrders = append(cancelOrders, s.DB.GetNewLimitLongOrders(qo)...)
		closeOrders = append(closeOrders, common.CloseLong(s.DB, s.BP, qo, ticker)...)
	}

	if shouldCloseShort {
		cancelOrders = append(cancelOrders, s.DB.GetNewLimitShortOrders(qo)...)
		closeOrders = append(closeOrders, common.CloseShort(s.DB, s.BP, qo, ticker)...)
	}

	if len(closeOrders) > 0 || len(cancelOrders) > 0 {
		return &t.TradeOrders{
			CloseOrders:  closeOrders,
			CancelOrders: cancelOrders,
		}
	}

	shouldOpenLong := lma_1 < lma_0 && (s.BP.View == t.ViewNeutral || s.BP.View == t.ViewLong)
	shouldOpenShort := hma_1 > hma_0 && (s.BP.View == t.ViewNeutral || s.BP.View == t.ViewShort)

	if shouldOpenLong && shouldOpenShort {
		return nil
	}

	if shouldOpenLong {
		openPrice := h.CalcStopLowerTicker(ticker.Price, float64(s.BP.Gap.OpenLimit), s.BP.PriceDigits)
		if openPrice < hma_0 {
			_qo := qo
			_qo.Side = t.OrderSideBuy
			_qo.OpenPrice = openPrice
			norder := s.DB.GetNearestOrder(_qo)
			if norder == nil || math.Abs(norder.OpenPrice-openPrice) >= s.BP.OrderGap {
				o := t.Order{
					ID:        h.GenID(),
					BotID:     s.BP.BotID,
					Exchange:  s.BP.Exchange,
					Symbol:    s.BP.Symbol,
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
			}
		}
	}

	if shouldOpenShort {
		openPrice := h.CalcStopUpperTicker(ticker.Price, float64(s.BP.Gap.OpenLimit), s.BP.PriceDigits)
		if openPrice > lma_0 {
			_qo := qo
			_qo.Side = t.OrderSideSell
			_qo.OpenPrice = openPrice
			norder := s.DB.GetNearestOrder(_qo)
			if norder == nil || math.Abs(openPrice-norder.OpenPrice) >= s.BP.OrderGap {
				o := t.Order{
					ID:        h.GenID(),
					BotID:     s.BP.BotID,
					Exchange:  s.BP.Exchange,
					Symbol:    s.BP.Symbol,
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
			}
		}
	}

	return &t.TradeOrders{
		OpenOrders:   openOrders,
		CloseOrders:  closeOrders,
		CancelOrders: cancelOrders,
	}
}
