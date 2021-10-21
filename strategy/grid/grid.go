package grid

import (
	"fmt"
	"os"
	"sort"

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
	if s.BP.UpperPrice <= s.BP.LowerPrice {
		fmt.Fprintln(os.Stderr, "The upper price must be greater than the lower price")
		return t.TradeOrders{}
	} else if s.BP.GridSize < 2 {
		fmt.Fprintln(os.Stderr, "Grid size must be greater than 1")
		return t.TradeOrders{}
	} else if s.BP.BaseQty == 0 && s.BP.QuoteQty == 0 {
		fmt.Fprintln(os.Stderr, "Quantity per grid must be greater than 0")
		os.Exit(0)
		return t.TradeOrders{}
	}

	var openOrders, closeOrders []t.Order

	buyPrice, sellPrice, gridWidth := common.GetGridRange(ticker.Price, s.BP.LowerPrice, s.BP.UpperPrice, s.BP.GridSize)

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
		zones, _ := common.GetGridZones(ticker.Price, s.BP.LowerPrice, s.BP.UpperPrice, s.BP.GridSize)
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
				openOrders = append(openOrders, o)
			}
			if count++; count == s.BP.OpenZones {
				break
			}
		}
	}

	if s.BP.View == t.ViewShort || s.BP.View == t.ViewNeutral {
		var count int64 = 0
		zones, _ := common.GetGridZones(ticker.Price, s.BP.LowerPrice, s.BP.UpperPrice, s.BP.GridSize)
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
				openOrders = append(openOrders, o)
			}
			if count++; count == s.BP.OpenZones {
				break
			}
		}
	}

	return t.TradeOrders{
		OpenOrders:  openOrders,
		CloseOrders: closeOrders,
	}
}
