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

func OnTick(params OnTickParams) *t.TradeOrders {
	ticker := params.Ticker
	p := params.BotParams
	db := params.DB

	var orders []t.Order

	buyPrice, sellPrice, gridWidth := strategy.GetGridRange(ticker.Price, p.LowerPrice, p.UpperPrice, p.GridSize)

	order := t.Order{
		BotID:    p.BotID,
		Exchange: ticker.Exchange,
		Symbol:   ticker.Symbol,
		Qty:      p.Qty,
		Status:   t.OrderStatusNew,
		Type:     t.OrderTypeLimit,
	}

	if p.OpenZones < 1 {
		p.OpenZones = 1
	}

	if p.View == t.ViewLong || p.View == t.ViewNeutral {
		var count int64 = 0
		zones, _ := strategy.GetGridZones(ticker.Price, p.LowerPrice, p.UpperPrice, p.GridSize)
		for _, zone := range zones {
			_order := order
			_order.Side = t.OrderSideBuy
			_order.OpenPrice = h.NormalizeDouble(buyPrice, p.PriceDigits)
			_order.ZonePrice = h.NormalizeDouble(zone, p.PriceDigits)
			if db.IsEmptyZone(_order) {
				if p.GridTP > 0 {
					_order.TPPrice = h.NormalizeDouble(zone+gridWidth*p.GridTP, p.PriceDigits)
				}
				orders = append(orders, _order)
			}
			if count++; count == p.OpenZones {
				break
			}
		}
	}

	if p.View == t.ViewShort || p.View == t.ViewNeutral {
		var count int64 = 0
		zones, _ := strategy.GetGridZones(ticker.Price, p.LowerPrice, p.UpperPrice, p.GridSize)
		// Sort DESC
		sort.Slice(zones, func(i, j int) bool {
			return zones[i] > zones[j]
		})
		for _, zone := range zones {
			_order := order
			_order.Side = t.OrderSideSell
			_order.OpenPrice = h.NormalizeDouble(sellPrice, p.PriceDigits)
			_order.ZonePrice = h.NormalizeDouble(zone, p.PriceDigits)
			if db.IsEmptyZone(_order) {
				if p.GridTP > 0 {
					_order.TPPrice = h.NormalizeDouble(zone-gridWidth*p.GridTP, p.PriceDigits)
				}
				orders = append(orders, _order)
			}
			if count++; count == p.OpenZones {
				break
			}
		}
	}

	return &t.TradeOrders{
		OpenOrders: orders,
	}
}
