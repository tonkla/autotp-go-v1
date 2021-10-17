package grid

import (
	"sort"

	"github.com/tonkla/autotp/db"
	h "github.com/tonkla/autotp/helper"
	"github.com/tonkla/autotp/strategy"
	t "github.com/tonkla/autotp/types"
)

type OnTickParams struct {
	DB        *db.DB
	Ticker    *t.Ticker
	OrderBook t.OrderBook
	BotParams t.BotParams
}

func OnTick(params OnTickParams) *t.TradeOrder {
	ticker := params.Ticker
	bp := params.BotParams
	db := params.DB

	var orders []t.Order

	buyPrice, sellPrice, gridWidth := strategy.GetGridRange(ticker.Price, bp.LowerPrice, bp.UpperPrice, bp.GridSize)

	qo := t.QueryOrder{
		BotID:    bp.BotID,
		Exchange: ticker.Exchange,
		Symbol:   ticker.Symbol,
	}

	order := t.Order{
		BotID:    bp.BotID,
		Exchange: ticker.Exchange,
		Symbol:   ticker.Symbol,
		Status:   t.OrderStatusNew,
		Type:     t.OrderTypeLimit,
		Qty:      bp.BaseQty,
	}

	if bp.OpenZones < 1 {
		bp.OpenZones = 1
	}

	if bp.View == t.ViewLong || bp.View == t.ViewNeutral {
		var count int64 = 0
		zones, _ := strategy.GetGridZones(ticker.Price, bp.LowerPrice, bp.UpperPrice, bp.GridSize)
		for _, zone := range zones {
			o := order
			o.Side = t.OrderSideBuy
			o.OpenPrice = h.NormalizeDouble(buyPrice, bp.PriceDigits)
			o.ZonePrice = h.NormalizeDouble(zone, bp.PriceDigits)

			qo.Side = t.OrderSideBuy
			qo.ZonePrice = o.ZonePrice

			if db.IsEmptyZone(qo) {
				o.ID = h.GenID()
				_qty := h.NormalizeDouble(bp.QuoteQty/o.OpenPrice, bp.QtyDigits)
				if _qty > o.Qty {
					o.Qty = _qty
				}
				if bp.GridTP > 0 {
					o.TPPrice = h.NormalizeDouble(zone+gridWidth*bp.GridTP, bp.PriceDigits)
				}
				orders = append(orders, o)
			}
			if count++; count == bp.OpenZones {
				break
			}
		}
	}

	if bp.View == t.ViewShort || bp.View == t.ViewNeutral {

		var count int64 = 0
		zones, _ := strategy.GetGridZones(ticker.Price, bp.LowerPrice, bp.UpperPrice, bp.GridSize)
		// Sort DESC
		sort.Slice(zones, func(i, j int) bool {
			return zones[i] > zones[j]
		})
		for _, zone := range zones {
			o := order
			o.Side = t.OrderSideSell
			o.OpenPrice = h.NormalizeDouble(sellPrice, bp.PriceDigits)
			o.ZonePrice = h.NormalizeDouble(zone, bp.PriceDigits)

			qo.Side = t.OrderSideSell
			qo.ZonePrice = o.ZonePrice

			if db.IsEmptyZone(qo) {
				o.ID = h.GenID()
				_qty := h.NormalizeDouble(bp.QuoteQty/o.OpenPrice, bp.QtyDigits)
				if _qty > o.Qty {
					o.Qty = _qty
				}
				if bp.GridTP > 0 {
					o.TPPrice = h.NormalizeDouble(zone-gridWidth*bp.GridTP, bp.PriceDigits)
				}
				orders = append(orders, o)
			}
			if count++; count == bp.OpenZones {
				break
			}
		}
	}

	return &t.TradeOrder{
		OpenOrders: orders,
	}
}
