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

	const numberOfBars = 30

	prices := s.EX.GetHistoricalPrices(s.BP.Symbol, s.BP.MATf1st, numberOfBars)
	if len(prices) == 0 || prices[len(prices)-1].Open == 0 {
		return nil
	}

	highs, lows := common.GetHighsLows(prices)

	close_1 := prices[len(prices)-2].Close

	hma := talib.WMA(highs, int(s.BP.MAPeriod1st))
	hma_0 := hma[len(hma)-1]

	lma := talib.WMA(lows, int(s.BP.MAPeriod1st))
	lma_0 := lma[len(lma)-1]

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
		if ticker.Price > hma_0 {
			closeOrders = append(closeOrders, common.TPSpot(s.DB, s.BP, qo, ticker, atr)...)
			if len(closeOrders) > 0 {
				return &t.TradeOrders{
					CloseOrders: closeOrders,
				}
			}
		}
	}

	if ticker.Price < lma_0 && ticker.Price < close_1 {
		openPrice := h.CalcStopLowerTicker(ticker.Price, float64(s.BP.Gap.OpenLimit), s.BP.PriceDigits)
		qo.Side = t.OrderSideBuy
		qo.OpenPrice = openPrice
		norder := s.DB.GetNearestOrder(qo)
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

	return &t.TradeOrders{
		OpenOrders: openOrders,
	}
}
