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
	var openOrders, closeOrders, cancelOrders []t.Order

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

	lows1st, closes1st := common.GetLowsCloses(prices1st)

	l1st_1 := lows1st[len(lows1st)-2]

	cma1st := talib.WMA(closes1st, int(s.BP.MAPeriod1st))
	cma1st_0 := cma1st[len(cma1st)-1]
	cma1st_1 := cma1st[len(cma1st)-2]

	lows2nd := common.GetLows(prices2nd)
	lma2nd := talib.WMA(lows2nd, int(s.BP.MAPeriod2nd))
	lma2nd_0 := lma2nd[len(lma2nd)-1]
	lma2nd_1 := lma2nd[len(lma2nd)-2]

	highs, lows, closes := common.GetHighsLowsCloses(prices3rd)

	hma := talib.WMA(highs, int(s.BP.MAPeriod3rd))
	hma_0 := hma[len(hma)-1]
	hma_2 := hma[len(hma)-3]

	lma := talib.WMA(lows, int(s.BP.MAPeriod3rd))
	lma_0 := lma[len(lma)-1]

	c1 := closes[len(closes)-2]
	c2 := closes[len(closes)-3]
	c3 := closes[len(closes)-4]

	cma := talib.WMA(closes, int(s.BP.MAPeriod3rd))
	cma_0 := cma[len(cma)-1]
	cma_1 := cma[len(cma)-2]

	atr := hma_0 - lma_0

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

	if s.BP.AutoTP {
		closeOrders = append(closeOrders, common.TPSpot(s.DB, s.BP, qo, ticker, atr)...)

		if len(closeOrders) == 0 && c2 > hma_2 && c3 < c2 && c2 > c1 {
			closeOrders = append(closeOrders, common.CloseProfitSpot(s.DB, s.BP, qo, ticker)...)
		}

		if len(closeOrders) > 0 {
			return &t.TradeOrders{
				CloseOrders: closeOrders,
			}
		}
	}

	shouldClose := cma1st_1 > cma1st_0 || l1st_1 > ticker.Price

	if shouldClose && s.BP.ForceClose {
		cancelOrders = append(cancelOrders, s.DB.GetNewLimitOrders(qo)...)
		closeOrders = append(closeOrders, common.CloseProfitSpot(s.DB, s.BP, qo, ticker)...)

		if len(cancelOrders) > 0 || len(closeOrders) > 0 {
			return &t.TradeOrders{
				CancelOrders: cancelOrders,
				CloseOrders:  closeOrders,
			}
		}
	}

	shouldOpen := lma2nd_1 < lma2nd_0 && !shouldClose

	if shouldOpen {
		openPrice := h.CalcStopLowerTicker(ticker.Price, float64(s.BP.Gap.OpenLimit), s.BP.PriceDigits)
		if (cma_1 < cma_0 && openPrice < cma_0-(s.BP.MoS*atr)) || cma_1 > cma_0 {
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
