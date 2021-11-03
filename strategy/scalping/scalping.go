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
	var openOrders, closeOrders []t.Order

	qo := t.QueryOrder{
		BotID:    s.BP.BotID,
		Exchange: s.BP.Exchange,
		Symbol:   s.BP.Symbol,
	}

	if s.BP.CloseLong || s.BP.CloseShort {
		if s.BP.CloseLong {
			closeOrders = append(closeOrders, common.CloseLong(s.DB, s.BP, qo, ticker)...)
		}
		if s.BP.CloseShort {
			closeOrders = append(closeOrders, common.CloseShort(s.DB, s.BP, qo, ticker)...)
		}
		return &t.TradeOrders{
			CloseOrders: closeOrders,
		}
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

	prices2nd := s.EX.GetHistoricalPrices(s.BP.Symbol, s.BP.MATf2nd, numberOfBars)
	prices3rd := s.EX.GetHistoricalPrices(s.BP.Symbol, s.BP.MATf3rd, numberOfBars)

	if len(prices2nd) == 0 || prices2nd[0].Close == 0 ||
		len(prices3rd) == 0 || prices3rd[0].Close == 0 {
		return nil
	}

	p2nd_0 := prices2nd[len(prices2nd)-1]
	p2nd_1 := prices2nd[len(prices2nd)-2]
	p3rd_0 := prices3rd[len(prices3rd)-1]
	p3rd_1 := prices3rd[len(prices3rd)-2]

	closes2nd := common.GetCloses(prices2nd)
	closes3rd := common.GetCloses(prices3rd)

	period := int(s.BP.MAPeriod1st)
	macl2nd := talib.WMA(closes2nd, period)
	macl3rd := talib.WMA(closes3rd, period)

	isUp := macl2nd[len(macl2nd)-2] < macl2nd[len(macl2nd)-1] &&
		macl3rd[len(macl3rd)-2] < macl3rd[len(macl3rd)-1] &&
		p2nd_1.High < p2nd_0.High &&
		p3rd_1.High < p3rd_0.High

	isDown := macl2nd[len(macl2nd)-2] > macl2nd[len(macl2nd)-1] &&
		macl3rd[len(macl3rd)-2] > macl3rd[len(macl3rd)-1] &&
		p2nd_1.Low > p2nd_0.Low &&
		p3rd_1.Low > p3rd_0.Low

	shouldCloseLong := macl3rd[len(macl3rd)-2] > macl3rd[len(macl3rd)-1] &&
		p3rd_1.Low > p3rd_0.Low

	shouldCloseShort := macl3rd[len(macl3rd)-2] < macl3rd[len(macl3rd)-1] &&
		p3rd_1.High < p3rd_0.High

	if shouldCloseLong {
		for _, o := range s.DB.GetFilledLimitLongOrders(qo) {
			if ticker.Price > o.OpenPrice && s.BP.AutoTP {
				tpo := common.TPLongNow(s.DB, s.BP, ticker, o)
				if tpo != nil {
					closeOrders = append(closeOrders, *tpo)
				}
			} else if s.BP.AutoSL {
				slo := common.SLLongNow(s.DB, s.BP, ticker, o)
				if slo != nil {
					closeOrders = append(closeOrders, *slo)
				}
			}
		}
	}

	if shouldCloseShort {
		for _, o := range s.DB.GetFilledLimitShortOrders(qo) {
			if ticker.Price < o.OpenPrice && s.BP.AutoTP {
				tpo := common.TPShortNow(s.DB, s.BP, ticker, o)
				if tpo != nil {
					closeOrders = append(closeOrders, *tpo)
				}
			} else if s.BP.AutoSL {
				slo := common.SLShortNow(s.DB, s.BP, ticker, o)
				if slo != nil {
					closeOrders = append(closeOrders, *slo)
				}
			}
		}
	}

	qo.Qty = h.NormalizeDouble(s.BP.BaseQty, s.BP.QtyDigits)
	qty := h.NormalizeDouble(s.BP.QuoteQty/ticker.Price, s.BP.QtyDigits)
	if qty > qo.Qty {
		qo.Qty = qty
	}

	if isUp {
		if s.BP.View == t.ViewNeutral || s.BP.View == t.ViewLong {
			openPrice := h.CalcStopLowerTicker(ticker.Price, openLimit, s.BP.PriceDigits)
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
