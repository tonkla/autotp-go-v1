package daily

import (
	h "github.com/tonkla/autotp/helper"
	"github.com/tonkla/autotp/rdb"
	"github.com/tonkla/autotp/strategy/common"
	t "github.com/tonkla/autotp/types"
)

type Strategy struct {
	DB *rdb.DB
	BP *t.BotParams
}

func New(db *rdb.DB, bp *t.BotParams) Strategy {
	return Strategy{
		DB: db,
		BP: bp,
	}
}

func (s Strategy) OnTick(ticker *t.Ticker) t.TradeOrders {
	var openOrders, closeOrders []t.Order

	prices := s.BP.HPrices

	slStop := float64(s.BP.SLim.SLStop)
	slLimit := float64(s.BP.SLim.SLLimit)
	tpStop := float64(s.BP.SLim.TPStop)
	tpLimit := float64(s.BP.SLim.TPLimit)
	openLimit := float64(s.BP.SLim.OpenLimit)

	p_0 := prices[len(prices)-1]
	if p_0.Open == 0 || p_0.High == 0 || p_0.Low == 0 || p_0.Close == 0 {
		return t.TradeOrders{}
	}
	p_1 := prices[len(prices)-2]
	c_1 := p_1.Close
	h_1 := p_1.High
	l_1 := p_1.Low

	trend := common.GetTrend(prices, int(s.BP.MAPeriod))
	atr := common.GetATR(prices, int(s.BP.MAPeriod))
	mos := (h_1 - l_1) * s.BP.MoS // The Margin of Safety

	qo := t.QueryOrder{
		BotID:    s.BP.BotID,
		Exchange: ticker.Exchange,
		Symbol:   ticker.Symbol,
		Qty:      h.NormalizeDouble(s.BP.BaseQty, s.BP.QtyDigits),
	}

	_qty := h.NormalizeDouble(s.BP.QuoteQty/ticker.Price, s.BP.QtyDigits)
	if _qty > s.BP.BaseQty {
		qo.Qty = _qty
	}

	// Uptrend -------------------------------------------------------------------
	if trend >= t.TrendUp1 {
		// Take Profit, by the configured Volatility Stop (TP)
		if s.BP.AutoTP {
			for _, o := range s.DB.GetFilledLimitBuyOrders(qo) {
				if ticker.Price > o.OpenPrice+atr*s.BP.AtrTP && s.DB.GetTPOrder(o.ID) == nil {
					tpo := t.Order{
						ID:          h.GenID(),
						BotID:       s.BP.BotID,
						Exchange:    o.Exchange,
						Symbol:      o.Symbol,
						Side:        t.OrderSideSell,
						Type:        t.OrderTypeTP,
						Status:      t.OrderStatusNew,
						Qty:         h.NormalizeDouble(o.Qty, s.BP.QtyDigits),
						StopPrice:   h.CalcTPStop(t.OrderSideSell, ticker.Price, tpStop, s.BP.PriceDigits),
						OpenPrice:   h.CalcTPStop(t.OrderSideSell, ticker.Price, tpLimit, s.BP.PriceDigits),
						OpenOrderID: o.ID,
					}
					closeOrders = append(closeOrders, tpo)
				}
			}
		}

		// Open a new limit order, when no active BUY order
		if (s.BP.View == t.ViewLong || s.BP.View == t.ViewNeutral) && ticker.Price < h_1-mos && ticker.Price < c_1 {
			qo.Side = t.OrderSideBuy
			qo.OpenTime = p_0.Time
			activeOrders := s.DB.GetLimitOrdersBySide(qo)
			maxOrders := 3
			if len(activeOrders) == 0 || (activeOrders[0].OpenTime < p_0.Time && len(activeOrders) < maxOrders) {
				o := t.Order{
					ID:        h.GenID(),
					BotID:     s.BP.BotID,
					Exchange:  qo.Exchange,
					Symbol:    qo.Symbol,
					Side:      t.OrderSideBuy,
					Type:      t.OrderTypeLimit,
					Status:    t.OrderStatusNew,
					Qty:       qo.Qty,
					OpenPrice: h.CalcStopBehindTicker(t.OrderSideBuy, ticker.Price, openLimit, s.BP.PriceDigits),
				}
				openOrders = append(openOrders, o)
			}
		}
	}

	// Downtrend -----------------------------------------------------------------
	if trend <= t.TrendDown1 {
		// Stop Loss, for BUY orders
		if s.BP.AutoSL {
			for _, o := range s.DB.GetFilledLimitBuyOrders(qo) {
				if s.DB.GetSLOrder(o.ID) != nil {
					continue
				}
				slo := t.Order{
					ID:          h.GenID(),
					BotID:       s.BP.BotID,
					Exchange:    o.Exchange,
					Symbol:      o.Symbol,
					Side:        t.OrderSideSell,
					Type:        t.OrderTypeSL,
					Status:      t.OrderStatusNew,
					Qty:         h.NormalizeDouble(o.Qty, s.BP.QtyDigits),
					StopPrice:   h.CalcSLStop(t.OrderSideSell, ticker.Price, slStop, s.BP.PriceDigits),
					OpenPrice:   h.CalcSLStop(t.OrderSideSell, ticker.Price, slLimit, s.BP.PriceDigits),
					OpenOrderID: o.ID,
				}
				closeOrders = append(closeOrders, slo)
			}
		}

		// Open a new limit order, when no active SELL order
		if (s.BP.View == t.ViewShort || s.BP.View == t.ViewNeutral) && ticker.Price > l_1+mos && ticker.Price > c_1 {
			qo.Side = t.OrderSideSell
			qo.OpenTime = p_0.Time
			activeOrders := s.DB.GetLimitOrdersBySide(qo)
			maxOrders := 3
			if len(activeOrders) == 0 || (activeOrders[0].OpenTime < p_0.Time && len(activeOrders) < maxOrders) {
				o := t.Order{
					ID:        h.GenID(),
					BotID:     s.BP.BotID,
					Exchange:  qo.Exchange,
					Symbol:    qo.Symbol,
					Side:      t.OrderSideSell,
					Type:      t.OrderTypeLimit,
					Status:    t.OrderStatusNew,
					Qty:       qo.Qty,
					OpenPrice: h.CalcStopBehindTicker(t.OrderSideSell, ticker.Price, openLimit, s.BP.PriceDigits),
				}
				openOrders = append(openOrders, o)
			}
		}
	}

	return t.TradeOrders{
		OpenOrders:  openOrders,
		CloseOrders: closeOrders,
	}
}
