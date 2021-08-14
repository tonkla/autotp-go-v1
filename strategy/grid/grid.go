package grid

import (
	"sort"
	"strings"

	"github.com/tonkla/autotp/db"
	"github.com/tonkla/autotp/strategy"
	t "github.com/tonkla/autotp/types"
)

type OnTickParams struct {
	Ticker    t.Ticker
	OrderBook t.OrderBook
	BotParams t.BotParams
	DB        db.DB
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

	view := strings.ToUpper(p.View)

	if view == t.ViewLong || view == "L" || view == t.ViewNeutral || view == "N" {
		var count int64 = 0
		zones, _ := strategy.GetGridZones(ticker.Price, p.LowerPrice, p.UpperPrice, p.GridSize)
		for _, zone := range zones {
			_order := order
			_order.Side = t.OrderSideBuy
			_order.OpenPrice = buyPrice
			_order.ZonePrice = zone
			if db.IsEmptyZone(_order) {
				if p.GridTP > 0 {
					_order.TPPrice = zone + gridWidth*p.GridTP
				}
				orders = append(orders, _order)
			}
			if count++; count == p.OpenZones {
				break
			}
		}
	}

	if view == t.ViewShort || view == "S" || view == t.ViewNeutral || view == "N" {
		var count int64 = 0
		zones, _ := strategy.GetGridZones(ticker.Price, p.LowerPrice, p.UpperPrice, p.GridSize)
		// Sort DESC
		sort.Slice(zones, func(i, j int) bool {
			return zones[i] > zones[j]
		})
		for _, zone := range zones {
			_order := order
			_order.Side = t.OrderSideSell
			_order.OpenPrice = sellPrice
			_order.ZonePrice = zone
			if db.IsEmptyZone(_order) {
				if p.GridTP > 0 {
					_order.TPPrice = zone - gridWidth*p.GridTP
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
