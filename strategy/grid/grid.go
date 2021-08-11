package grid

import (
	"strings"

	"github.com/tonkla/autotp/db"
	"github.com/tonkla/autotp/strategy"
	t "github.com/tonkla/autotp/types"
)

type OnTickParams struct {
	Ticker    t.Ticker
	OrderBook t.OrderBook
	BotParams t.BotParams
	HPrices   []t.HistoricalPrice
	DB        db.DB
}

func OnTick(params OnTickParams) *t.TradeOrders {
	ticker := params.Ticker
	p := params.BotParams
	db := params.DB

	var orders []t.Order

	lowerPrice, upperPrice, gridWidth := strategy.GetGridRange(ticker.Price, p.LowerPrice, p.UpperPrice, p.GridSize)

	trend := 100
	if len(params.HPrices) > 0 {
		trend = strategy.GetTrend(params.HPrices, int(p.MAPeriod))
	}

	order := t.Order{
		BotID:    p.BotID,
		Exchange: ticker.Exchange,
		Symbol:   ticker.Symbol,
		Qty:      p.Qty,
		Status:   t.OrderStatusNew,
		Type:     t.OrderTypeLimit,
	}

	view := strings.ToUpper(p.View)

	// A multiplier of deducted zones when a trend is strong
	const openGaps = 1

	if view == t.ViewLong || view == "L" || view == t.ViewNeutral || view == "N" {
		buyPrice := lowerPrice
		if p.ApplyTrend && trend < 100 && trend < t.TrendDown2 {
			buyPrice = lowerPrice - gridWidth*openGaps
		}
		if p.OpenAll {
			// Buy all available zones of the grid, please ensure your fund is plenty
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
			}
		} else {
			_order := order
			_order.Side = t.OrderSideBuy
			_order.OpenPrice = buyPrice
			_order.ZonePrice = buyPrice
			if db.IsEmptyZone(_order) {
				if p.GridTP > 0 {
					_order.TPPrice = buyPrice + gridWidth*p.GridTP
				}
				orders = append(orders, _order)
			}
		}
	}

	if view == t.ViewShort || view == "S" || view == t.ViewNeutral || view == "N" {
		sellPrice := upperPrice
		if p.ApplyTrend && trend < 100 && trend > t.TrendUp2 {
			sellPrice = upperPrice + gridWidth*openGaps
		}
		if p.OpenAll {
			// Sell all available zones of the grid, please ensure your fund is plenty
			zones, _ := strategy.GetGridZones(ticker.Price, p.LowerPrice, p.UpperPrice, p.GridSize)
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
			}
		} else {
			_order := order
			_order.Side = t.OrderSideSell
			_order.OpenPrice = sellPrice
			_order.ZonePrice = sellPrice
			if db.IsEmptyZone(_order) {
				if p.GridTP > 0 {
					_order.TPPrice = sellPrice - gridWidth*p.GridTP
				}
				orders = append(orders, _order)
			}
		}
	}

	return &t.TradeOrders{
		OpenOrders: orders,
	}
}
