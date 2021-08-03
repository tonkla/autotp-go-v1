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

const openGaps = 2

func OnTick(params OnTickParams) *t.TradeOrders {
	ticker := params.Ticker
	p := params.BotParams
	db := params.DB

	var orders []t.Order

	buyPrice, sellPrice, gridWidth := strategy.GetGridRange(ticker.Price, p.LowerPrice, p.UpperPrice, p.GridSize)
	trend := strategy.GetTrend(params.HPrices, int(p.MAPeriod))

	order := t.Order{
		BotID:    p.BotID,
		Exchange: ticker.Exchange,
		Symbol:   ticker.Symbol,
		Qty:      p.Qty,
		Status:   t.OrderStatusNew,
		Type:     t.OrderTypeLimit,
	}

	view := strings.ToUpper(p.View)

	if view == t.ViewNeutral || view == "N" || view == t.ViewLong || view == "L" {
		lowerPrice := buyPrice
		if p.FollowTrend && trend < t.TrendDown2 {
			lowerPrice = buyPrice - gridWidth*openGaps
		}
		if p.OpenAll {
			// Buy all available zones of the grid, please ensure your fund is plenty
			zones, _ := strategy.GetGridZones(ticker.Price, p.LowerPrice, p.UpperPrice, p.GridSize)
			for _, zone := range zones {
				_order := order
				_order.Side = t.OrderSideBuy
				_order.OpenPrice = zone
				if db.GetLimitOrder(_order, p.Slippage) == nil {
					_order.OpenPrice = lowerPrice
					if p.GridTP > 0 {
						_order.TPPrice = zone + gridWidth*p.GridTP
					}
					orders = append(orders, _order)
				}
			}
		} else {
			_order := order
			_order.Side = t.OrderSideBuy
			_order.OpenPrice = lowerPrice
			if db.GetLimitOrder(_order, p.Slippage) == nil {
				if p.GridTP > 0 {
					_order.TPPrice = lowerPrice + gridWidth*p.GridTP
				}
				orders = append(orders, _order)
			}
		}
	}

	if view == t.ViewNeutral || view == "N" || view == t.ViewShort || view == "S" {
		upperPrice := sellPrice
		if p.FollowTrend && trend > t.TrendUp2 {
			upperPrice = sellPrice + gridWidth*openGaps
		}
		if p.OpenAll {
			// Sell all available zones of the grid, please ensure your fund is plenty
			zones, _ := strategy.GetGridZones(ticker.Price, p.LowerPrice, p.UpperPrice, p.GridSize)
			for _, zone := range zones {
				_order := order
				_order.Side = t.OrderSideSell
				_order.OpenPrice = zone
				if db.GetLimitOrder(_order, p.Slippage) == nil {
					_order.OpenPrice = upperPrice
					if p.GridTP > 0 {
						_order.TPPrice = zone - gridWidth*p.GridTP
					}
					orders = append(orders, _order)
				}
			}
		} else {
			_order := order
			_order.Side = t.OrderSideSell
			_order.OpenPrice = upperPrice
			if db.GetLimitOrder(_order, p.Slippage) == nil {
				if p.GridTP > 0 {
					_order.TPPrice = upperPrice - gridWidth*p.GridTP
				}
				orders = append(orders, _order)
			}
		}
	}

	return &t.TradeOrders{
		OpenOrders: orders,
	}
}
