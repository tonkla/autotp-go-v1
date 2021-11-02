package daily

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
		Exchange: ticker.Exchange,
		Symbol:   ticker.Symbol,
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

	const numberOfBars = 50
	prices := s.EX.GetHistoricalPrices(s.BP.Symbol, s.BP.MATf1st, numberOfBars)

	if len(prices) == 0 || prices[len(prices)-1].Open == 0 {
		return nil
	}

	p_1 := prices[len(prices)-2]
	c_1 := p_1.Close
	h_1 := p_1.High
	l_1 := p_1.Low

	cma := talib.WMA(common.GetCloses(prices), int(s.BP.MAPeriod1st))
	cma_0 := cma[len(cma)-1]
	cma_1 := cma[len(cma)-2]

	atr := 0.0
	if s.BP.AtrSL > 0 || s.BP.AtrTP > 0 {
		_atr := common.GetATR(prices, int(s.BP.MAPeriod1st))
		if _atr == nil {
			return nil
		}
		atr = *_atr
	}

	mos := (h_1 - l_1) * s.BP.MoS // The Margin of Safety

	qo.Qty = h.NormalizeDouble(s.BP.BaseQty, s.BP.QtyDigits)
	qty := h.NormalizeDouble(s.BP.QuoteQty/ticker.Price, s.BP.QtyDigits)
	if qty > qo.Qty {
		qo.Qty = qty
	}

	openLimit := float64(s.BP.SLim.OpenLimit)
	isFutures := s.BP.Product == t.ProductFutures

	if s.BP.AutoSL {
		closeOrders = append(closeOrders, common.SLLong(s.DB, s.BP, qo, ticker, atr)...)
		closeOrders = append(closeOrders, common.SLShort(s.DB, s.BP, qo, ticker, atr)...)
	}

	if s.BP.AutoTP {
		closeOrders = append(closeOrders, common.TPLong(s.DB, s.BP, qo, ticker, atr)...)
		closeOrders = append(closeOrders, common.TPShort(s.DB, s.BP, qo, ticker, atr)...)
	}

	// Uptrend -------------------------------------------------------------------
	if cma_1 < cma_0 {
		if (s.BP.View == t.ViewLong || s.BP.View == t.ViewNeutral) && ticker.Price < h_1-mos && ticker.Price < c_1 {
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
					OpenPrice: openPrice,
				}
				if isFutures {
					o.PosSide = t.OrderPosSideLong
				}
				openOrders = append(openOrders, o)
			}
		}
	}

	// Downtrend -----------------------------------------------------------------
	if cma_1 > cma_0 {
		if (s.BP.View == t.ViewShort || s.BP.View == t.ViewNeutral) && ticker.Price > l_1+mos && ticker.Price > c_1 {
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
					OpenPrice: openPrice,
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
