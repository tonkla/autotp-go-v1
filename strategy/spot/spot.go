package spot

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

	const numberOfBars = 50
	prices1st := s.EX.GetHistoricalPrices(s.BP.Symbol, s.BP.MATf1st, numberOfBars)
	if len(prices1st) == 0 || prices1st[len(prices1st)-1].Open == 0 {
		return nil
	}

	prices2nd := s.EX.GetHistoricalPrices(s.BP.Symbol, s.BP.MATf2nd, numberOfBars)
	if len(prices2nd) == 0 || prices2nd[len(prices2nd)-1].Open == 0 {
		return nil
	}

	lows1st := common.GetLows(prices1st)
	low1st_1 := lows1st[len(lows1st)-2]
	lma1st := talib.WMA(lows1st, int(s.BP.MAPeriod1st))
	lma1st_0 := lma1st[len(lma1st)-1]
	lma1st_1 := lma1st[len(lma1st)-2]

	highs, lows := common.GetHighsLows(prices2nd)
	hma := talib.WMA(highs, int(s.BP.MAPeriod2nd))
	hma_0 := hma[len(hma)-1]
	lma := talib.WMA(lows, int(s.BP.MAPeriod2nd))
	lma_0 := lma[len(lma)-1]
	mma_0 := lma_0 + ((hma_0 - lma_0) / 2)

	qo := t.QueryOrder{
		BotID:    s.BP.BotID,
		Exchange: s.BP.Exchange,
		Symbol:   s.BP.Symbol,
	}

	qo.Qty = h.NormalizeDouble(s.BP.BaseQty, s.BP.QtyDigits)
	qty := h.NormalizeDouble(s.BP.QuoteQty/ticker.Price, s.BP.QtyDigits)
	if qty > qo.Qty {
		qo.Qty = qty
	}

	atr := hma_0 - lma_0

	if s.BP.AutoTP {
		closeOrders = append(closeOrders, common.TPSpot(s.DB, s.BP, qo, ticker, atr)...)

		if ticker.Price < low1st_1 {
			closeOrders = append(closeOrders, common.CloseSpot(s.DB, s.BP, qo, ticker)...)
		}

		if len(closeOrders) > 0 {
			return &t.TradeOrders{
				CloseOrders: closeOrders,
			}
		}
	}

	if lma1st_1 < lma1st_0 {
		openPrice := h.CalcStopLowerTicker(ticker.Price, float64(s.BP.Gap.OpenLimit), s.BP.PriceDigits)
		if openPrice < mma_0 {
			_qo := qo
			_qo.Side = t.OrderSideBuy
			_qo.OpenPrice = openPrice
			norder := s.DB.GetNearestOrder(_qo)
			if norder == nil || math.Abs(norder.OpenPrice-openPrice) >= s.BP.OrderGapATR*atr {
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
				openOrders = append(openOrders, o)
			}
		}
	}

	return &t.TradeOrders{
		OpenOrders: openOrders,
	}
}
