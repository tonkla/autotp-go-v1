package grid

import (
	"fmt"
	"os"

	"github.com/tonkla/autotp/exchange"
	h "github.com/tonkla/autotp/helper"
	"github.com/tonkla/autotp/rdb"
	"github.com/tonkla/autotp/strategy/common"
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
	if s.BP.UpperPrice <= s.BP.LowerPrice {
		fmt.Fprintln(os.Stderr, "The upper price must be greater than the lower price")
		return nil
	} else if s.BP.GridSize < 2 {
		fmt.Fprintln(os.Stderr, "Grid size must be greater than 1")
		return nil
	} else if s.BP.BaseQty == 0 && s.BP.QuoteQty == 0 {
		fmt.Fprintln(os.Stderr, "Quantity per grid must be greater than 0")
		return nil
	}

	if s.BP.ApplyTA {
		const numberOfBars = 50
		prices4h := s.EX.Get4hHistoricalPrices(s.BP.Symbol, numberOfBars)
		if !common.IsUnderMA(ticker.Price, prices4h, s.BP.MAPeriod, 0) {
			return nil
		}
		prices1h := s.EX.Get1hHistoricalPrices(s.BP.Symbol, numberOfBars)
		if !common.IsUnderMA(ticker.Price, prices1h, s.BP.MAPeriod, 0) {
			return nil
		}
	}

	var openOrders, closeOrders []t.Order

	qo := t.QueryOrder{
		Exchange: s.BP.Exchange,
		Symbol:   s.BP.Symbol,
		BotID:    s.BP.BotID,
	}

	lowerPrice, _, gridWidth := common.GetGridRange(ticker.Price, s.BP.LowerPrice, s.BP.UpperPrice, s.BP.GridSize)

	if s.BP.OpenZones < 1 {
		s.BP.OpenZones = 1
	}

	if s.BP.View == t.ViewLong || s.BP.View == t.ViewNeutral {
		if s.BP.StartPrice > 0 && ticker.Price > s.BP.StartPrice && len(s.DB.GetActiveLimitOrders(qo)) == 0 {
			return nil
		}

		var count int64 = 0

		zones, _ := common.GetGridZones(ticker.Price, s.BP.LowerPrice, s.BP.UpperPrice, s.BP.GridSize)
		for _, zone := range zones {
			zonePrice := h.NormalizeDouble(zone, s.BP.PriceDigits)
			qo.ZonePrice = zonePrice
			qo.Side = t.OrderSideBuy
			if s.DB.IsEmptyZone(qo) {
				o := t.Order{
					ID:        h.GenID(),
					Exchange:  s.BP.Exchange,
					Symbol:    s.BP.Symbol,
					BotID:     s.BP.BotID,
					Qty:       s.BP.BaseQty,
					Status:    t.OrderStatusNew,
					Type:      t.OrderTypeLimit,
					Side:      t.OrderSideBuy,
					OpenPrice: h.NormalizeDouble(lowerPrice, s.BP.PriceDigits),
					ZonePrice: zonePrice,
				}
				_qty := h.NormalizeDouble(s.BP.QuoteQty/o.OpenPrice, s.BP.QtyDigits)
				if _qty > o.Qty {
					o.Qty = _qty
				}
				openOrders = append(openOrders, o)
			}
			if count++; count == s.BP.OpenZones {
				break
			}
		}

		if s.BP.GridTP > 0 {
			o := s.DB.GetLowestFilledBuyOrder(qo)
			if o != nil && s.DB.GetTPOrder(o.ID) == nil {
				_tp := h.NormalizeDouble(o.ZonePrice+gridWidth*s.BP.GridTP, s.BP.PriceDigits)
				_stop := h.CalcTPStop(t.OrderSideBuy, _tp, float64(s.BP.SLim.TPStop), s.BP.PriceDigits)
				if ticker.Price+(_tp-_stop) > _stop {
					tpo := t.Order{
						ID:          h.GenID(),
						Exchange:    s.BP.Exchange,
						Symbol:      s.BP.Symbol,
						BotID:       s.BP.BotID,
						Side:        t.OrderSideSell,
						Type:        t.OrderTypeTP,
						Status:      t.OrderStatusNew,
						Qty:         o.Qty,
						OpenOrderID: o.ID,
						OpenPrice:   _tp,
						StopPrice:   _stop,
					}
					closeOrders = append(closeOrders, tpo)
				}
			}
		}
	}

	return &t.TradeOrders{
		OpenOrders:  openOrders,
		CloseOrders: closeOrders,
	}
}
