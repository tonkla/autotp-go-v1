package scalping

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
	var openOrders, closeOrders []t.Order

	qo := t.QueryOrder{
		BotID:    s.BP.BotID,
		Exchange: ticker.Exchange,
		Symbol:   ticker.Symbol,
		Qty:      h.NormalizeDouble(s.BP.BaseQty, s.BP.QtyDigits),
	}
	qty := h.NormalizeDouble(s.BP.QuoteQty/ticker.Price, s.BP.QtyDigits)
	if qty > qo.Qty {
		qo.Qty = qty
	}

	if s.BP.AutoSL {
		closeOrders = append(closeOrders, common.SLLong(s.DB, s.BP, qo, ticker, 0)...)
		closeOrders = append(closeOrders, common.SLShort(s.DB, s.BP, qo, ticker, 0)...)
	}

	if s.BP.AutoTP {
		closeOrders = append(closeOrders, common.TPLong(s.DB, s.BP, qo, ticker, 0)...)
		closeOrders = append(closeOrders, common.TPShort(s.DB, s.BP, qo, ticker, 0)...)
	}

	const numberOfBars = 50

	openLimit := float64(s.BP.SLim.OpenLimit)
	isFutures := s.BP.Product == t.ProductFutures

	prices1h := s.EX.Get1hHistoricalPrices(s.BP.Symbol, numberOfBars)
	prices15m := s.EX.Get15mHistoricalPrices(s.BP.Symbol, numberOfBars)

	if len(prices1h) == 0 || prices1h[0].Close == 0 ||
		len(prices15m) == 0 || prices15m[0].Close == 0 {
		return nil
	}

	p1h_0 := prices1h[len(prices1h)-1]
	p1h_1 := prices1h[len(prices1h)-2]
	p15m_0 := prices15m[len(prices15m)-1]
	p15m_1 := prices15m[len(prices15m)-2]

	closes1h := common.GetCloses(prices1h)
	closes15m := common.GetCloses(prices15m)

	period := int(s.BP.MAPeriod)
	macl1h := talib.WMA(closes1h, period)
	macl15m := talib.WMA(closes15m, period)

	isUp := macl1h[len(macl1h)-2] < macl1h[len(macl1h)-1] &&
		macl15m[len(macl15m)-2] < macl15m[len(macl15m)-1] &&
		p1h_1.Low < p1h_0.Low &&
		p15m_1.Low < p15m_0.Low

	isDown := macl1h[len(macl1h)-2] > macl1h[len(macl1h)-1] &&
		macl15m[len(macl15m)-2] > macl15m[len(macl15m)-1] &&
		p1h_1.High > p1h_0.High &&
		p15m_1.High > p15m_0.High

	shouldCloseLong := macl15m[len(macl15m)-2] > macl15m[len(macl15m)-1] &&
		p15m_1.High > p15m_0.High

	shouldCloseShort := macl15m[len(macl15m)-2] < macl15m[len(macl15m)-1] &&
		p15m_1.Low < p15m_0.Low

	if shouldCloseLong {
		for _, o := range s.DB.GetFilledLimitLongOrders(qo) {
			if ticker.Price > o.OpenPrice && s.BP.AutoTP {
				closeOrders = append(closeOrders, common.TPLong(s.DB, s.BP, qo, ticker, 0)...)
			} else if s.BP.AutoSL {
				closeOrders = append(closeOrders, common.SLLong(s.DB, s.BP, qo, ticker, 0)...)
			}
		}
	}

	if shouldCloseShort {
		for _, o := range s.DB.GetFilledLimitShortOrders(qo) {
			if ticker.Price < o.OpenPrice && s.BP.AutoTP {
				closeOrders = append(closeOrders, common.TPShort(s.DB, s.BP, qo, ticker, 0)...)
			} else if s.BP.AutoSL {
				closeOrders = append(closeOrders, common.SLShort(s.DB, s.BP, qo, ticker, 0)...)
			}
		}
	}

	if isUp {
		if s.BP.View == t.ViewNeutral || s.BP.View == t.ViewLong {
			openPrice := h.CalcStopLowerTicker(ticker.Price, openLimit, s.BP.PriceDigits)
			qo.OpenPrice = openPrice
			qo.Side = t.OrderSideBuy
			norder := s.DB.GetNearestOrder(qo)
			if norder == nil || math.Abs(norder.OpenPrice-openPrice) >= s.BP.OrderGap {
				o := t.Order{
					ID:        h.GenID(),
					BotID:     s.BP.BotID,
					Exchange:  qo.Exchange,
					Symbol:    qo.Symbol,
					Side:      t.OrderSideBuy,
					Type:      t.OrderTypeLimit,
					Status:    t.OrderStatusNew,
					Qty:       qo.Qty,
					OpenPrice: qo.OpenPrice,
				}
				if isFutures {
					o.PosSide = t.OrderPosSideLong
				}
				openOrders = append(openOrders, o)
			}
		}
	}

	if isDown {
		if s.BP.View == t.ViewNeutral || s.BP.View == t.ViewShort {
			openPrice := h.CalcStopUpperTicker(ticker.Price, openLimit, s.BP.PriceDigits)
			qo.OpenPrice = openPrice
			qo.Side = t.OrderSideSell
			norder := s.DB.GetNearestOrder(qo)
			if norder == nil || math.Abs(norder.OpenPrice-openPrice) >= s.BP.OrderGap {
				o := t.Order{
					ID:        h.GenID(),
					BotID:     s.BP.BotID,
					Exchange:  qo.Exchange,
					Symbol:    qo.Symbol,
					Side:      t.OrderSideSell,
					Type:      t.OrderTypeLimit,
					Status:    t.OrderStatusNew,
					Qty:       qo.Qty,
					OpenPrice: qo.OpenPrice,
				}
				if isFutures {
					o.PosSide = t.OrderPosSideShort
				}
				openOrders = append(openOrders, o)
			}
		}
	}

	return &t.TradeOrders{
		OpenOrders:  openOrders,
		CloseOrders: closeOrders,
	}
}
