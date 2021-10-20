package grid

import (
	"sort"

	rdb "github.com/tonkla/autotp/db"
	h "github.com/tonkla/autotp/helper"
	c "github.com/tonkla/autotp/strategy/common"
	t "github.com/tonkla/autotp/types"
)

type Strategy struct {
	DB rdb.DB
	BP t.BotParams
}

func New(db rdb.DB, bp t.BotParams) c.Repository {
	return Strategy{
		DB: db,
		BP: bp,
	}
}

func (s Strategy) OnTick(ticker t.Ticker) t.TradeOrders {
	var orders []t.Order

	buyPrice, sellPrice, gridWidth := c.GetGridRange(ticker.Price, s.BP.LowerPrice, s.BP.UpperPrice, s.BP.GridSize)

	qo := t.QueryOrder{
		BotID:    s.BP.BotID,
		Exchange: ticker.Exchange,
		Symbol:   ticker.Symbol,
	}

	order := t.Order{
		BotID:    s.BP.BotID,
		Exchange: ticker.Exchange,
		Symbol:   ticker.Symbol,
		Status:   t.OrderStatusNew,
		Type:     t.OrderTypeLimit,
		Qty:      s.BP.BaseQty,
	}

	if s.BP.OpenZones < 1 {
		s.BP.OpenZones = 1
	}

	if s.BP.View == t.ViewLong || s.BP.View == t.ViewNeutral {
		var count int64 = 0
		zones, _ := c.GetGridZones(ticker.Price, s.BP.LowerPrice, s.BP.UpperPrice, s.BP.GridSize)
		for _, zone := range zones {
			o := order
			o.Side = t.OrderSideBuy
			o.OpenPrice = h.NormalizeDouble(buyPrice, s.BP.PriceDigits)
			o.ZonePrice = h.NormalizeDouble(zone, s.BP.PriceDigits)

			qo.Side = t.OrderSideBuy
			qo.ZonePrice = o.ZonePrice

			if s.DB.IsEmptyZone(qo) {
				o.ID = h.GenID()
				_qty := h.NormalizeDouble(s.BP.QuoteQty/o.OpenPrice, s.BP.QtyDigits)
				if _qty > o.Qty {
					o.Qty = _qty
				}
				if s.BP.GridTP > 0 {
					o.TPPrice = h.NormalizeDouble(zone+gridWidth*s.BP.GridTP, s.BP.PriceDigits)
				}
				orders = append(orders, o)
			}
			if count++; count == s.BP.OpenZones {
				break
			}
		}
	}

	if s.BP.View == t.ViewShort || s.BP.View == t.ViewNeutral {
		var count int64 = 0
		zones, _ := c.GetGridZones(ticker.Price, s.BP.LowerPrice, s.BP.UpperPrice, s.BP.GridSize)
		// Sort DESC
		sort.Slice(zones, func(i, j int) bool {
			return zones[i] > zones[j]
		})
		for _, zone := range zones {
			o := order
			o.Side = t.OrderSideSell
			o.OpenPrice = h.NormalizeDouble(sellPrice, s.BP.PriceDigits)
			o.ZonePrice = h.NormalizeDouble(zone, s.BP.PriceDigits)

			qo.Side = t.OrderSideSell
			qo.ZonePrice = o.ZonePrice

			if s.DB.IsEmptyZone(qo) {
				o.ID = h.GenID()
				_qty := h.NormalizeDouble(s.BP.QuoteQty/o.OpenPrice, s.BP.QtyDigits)
				if _qty > o.Qty {
					o.Qty = _qty
				}
				if s.BP.GridTP > 0 {
					o.TPPrice = h.NormalizeDouble(zone-gridWidth*s.BP.GridTP, s.BP.PriceDigits)
				}
				orders = append(orders, o)
			}
			if count++; count == s.BP.OpenZones {
				break
			}
		}
	}

	return t.TradeOrders{
		OpenOrders: orders,
	}
}
