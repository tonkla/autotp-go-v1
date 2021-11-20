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

		if s.BP.GridTP > 0 {
			o := s.DB.GetLowestFilledBuyOrder(qo)
			if o != nil && s.DB.GetTPOrder(o.ID) == nil {
				tpPrice := h.NormalizeDouble(o.ZonePrice+gridWidth*s.BP.GridTP, s.BP.PriceDigits)
				stopPrice := h.CalcTPStop(o.Side, tpPrice, float64(s.BP.Gap.TPStop), s.BP.PriceDigits)
				if ticker.Price+(tpPrice-stopPrice) > stopPrice {
					if ticker.Price > stopPrice {
						tpPrice = h.CalcStopUpperTicker(ticker.Price, float64(s.BP.Gap.TPLimit), s.BP.PriceDigits)
						stopPrice = h.CalcStopUpperTicker(ticker.Price, float64(s.BP.Gap.TPStop), s.BP.PriceDigits)
					}
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
						OpenPrice:   tpPrice,
						StopPrice:   stopPrice,
					}
					closeOrders = append(closeOrders, tpo)
				}
			}
		}

		if s.BP.ApplyTA {
			const numberOfBars = 50
			prices2nd := s.EX.GetHistoricalPrices(s.BP.Symbol, s.BP.MATf2nd, numberOfBars)
			if len(prices2nd) < numberOfBars {
				return &t.TradeOrders{
					CloseOrders: closeOrders,
				}
			}
			atr2nd := common.GetSimpleATR(prices2nd, int(s.BP.MAPeriod2nd))
			if !common.IsLowerMA(ticker.Price, prices2nd, s.BP.MAPeriod2nd, 0.4*atr2nd) {
				return &t.TradeOrders{
					CloseOrders: closeOrders,
				}
			}
			prices3rd := s.EX.GetHistoricalPrices(s.BP.Symbol, s.BP.MATf3rd, numberOfBars)
			if len(prices3rd) < numberOfBars {
				return &t.TradeOrders{
					CloseOrders: closeOrders,
				}
			}
			atr3rd := common.GetSimpleATR(prices3rd, int(s.BP.MAPeriod3rd))
			if !common.IsLowerMA(ticker.Price, prices3rd, s.BP.MAPeriod3rd, 0.4*atr3rd) {
				return &t.TradeOrders{
					CloseOrders: closeOrders,
				}
			}
		}

		var count int64 = 0

		openPrice := h.NormalizeDouble(lowerPrice, s.BP.PriceDigits)
		zones, _ := common.GetGridZones(ticker.Price, s.BP.LowerPrice, s.BP.UpperPrice, s.BP.GridSize)
		for _, zone := range zones {
			zonePrice := h.NormalizeDouble(zone, s.BP.PriceDigits)
			qo.ZonePrice = zonePrice
			qo.Side = t.OrderSideBuy
			if s.DB.IsEmptyZone(qo) {
				o := t.Order{
					ID:        fmt.Sprintf("%s%d", h.GenID(), count),
					Exchange:  s.BP.Exchange,
					Symbol:    s.BP.Symbol,
					BotID:     s.BP.BotID,
					Qty:       s.BP.BaseQty,
					Status:    t.OrderStatusNew,
					Type:      t.OrderTypeLimit,
					Side:      t.OrderSideBuy,
					OpenPrice: openPrice,
					ZonePrice: zonePrice,
				}
				qty := h.NormalizeDouble(s.BP.QuoteQty/o.OpenPrice, s.BP.QtyDigits)
				if qty > o.Qty {
					o.Qty = qty
				}
				openOrders = append(openOrders, o)
			}
			if count++; count == s.BP.OpenZones {
				break
			}
		}
	}

	return &t.TradeOrders{
		OpenOrders:  openOrders,
		CloseOrders: closeOrders,
	}
}
