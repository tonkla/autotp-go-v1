package trend

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

	const numberOfBars = 50
	prices := s.EX.GetHistoricalPrices(s.BP.Symbol, s.BP.MATimeframe, numberOfBars)

	if len(prices) == 0 || prices[len(prices)-1].Open == 0 {
		return nil
	}

	cma := talib.WMA(common.GetCloses(prices), int(s.BP.MAPeriod))
	cma_0 := cma[len(cma)-1]
	cma_1 := cma[len(cma)-2]

	_atr := common.GetATR(prices, int(s.BP.MAPeriod))
	if _atr == nil {
		return nil
	}
	atr := *_atr

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

	// Uptrend: Open Long --------------------------------------------------------
	if cma_1 < cma_0 {
		if s.BP.View == t.ViewNeutral || s.BP.View == t.ViewLong {
			openPrice := h.CalcStopLowerTicker(ticker.Price, openLimit, s.BP.PriceDigits)
			if openPrice < cma_0 {
				qo.Side = t.OrderSideBuy
				qo.OpenPrice = openPrice
				norder := s.DB.GetNearestOrder(qo)
				// Open a new limit order with safe minimum price gap
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
	}

	// Downtrend: Open Short -----------------------------------------------------
	if cma_1 > cma_0 {
		if s.BP.View == t.ViewNeutral || s.BP.View == t.ViewShort {
			openPrice := h.CalcStopUpperTicker(ticker.Price, openLimit, s.BP.PriceDigits)
			if openPrice > cma_0 {
				qo.Side = t.OrderSideSell
				qo.OpenPrice = openPrice
				norder := s.DB.GetNearestOrder(qo)
				// Open a new limit order with safe minimum price gap
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
	}

	return &t.TradeOrders{
		OpenOrders:  openOrders,
		CloseOrders: closeOrders,
	}
}
