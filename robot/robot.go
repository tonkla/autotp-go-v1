package robot

import (
	"fmt"
	"os"

	"github.com/tonkla/autotp/db"
	"github.com/tonkla/autotp/exchange"
	h "github.com/tonkla/autotp/helper"
	t "github.com/tonkla/autotp/types"
)

// Note: cannot move this to /types because of circular import of 'db'
type AppParams struct {
	DB     db.DB
	Repo   exchange.Repository
	Ticker t.Ticker
	TO     t.TradeOrder
	QO     t.QueryOrder
	BP     t.BotParams
}

func PlaceAsMaker(p *AppParams) {
}

func PlaceAsTaker(p *AppParams) {
}

func OpenLimitSpotOrders(p *AppParams) {
	for _, o := range p.TO.OpenOrders {
		if o.OpenPrice < h.CalcStopBehindTicker(t.OrderSideBuy, p.Ticker.Price,
			float64(p.BP.SLim.OpenLimit), p.BP.PriceDigits) {
			return
		}

		exo, err := p.Repo.OpenLimitOrder(o)
		if err != nil || exo == nil {
			h.Log("OpenOrder")
			os.Exit(1)
		}

		o.RefID = exo.RefID
		o.Status = exo.Status
		o.OpenPrice = exo.OpenPrice
		o.OpenTime = exo.OpenTime
		err = p.DB.CreateOrder(o)
		if err != nil {
			h.Log("CreateOrder", err)
			continue
		}
		h.LogNew(&o)
	}
}

func OpenLimitFuturesOrders(p *AppParams) {
	for _, o := range p.TO.OpenOrders {
		exo, err := p.Repo.OpenLimitOrder(o)
		if err != nil || exo == nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		o.RefID = exo.RefID
		o.Status = exo.Status
		o.OpenTime = exo.OpenTime
		o.OpenPrice = exo.OpenPrice
		err = p.DB.CreateOrder(o)
		if err != nil {
			h.Log(err)
			return
		}

		h.LogNewF(&o)
	}
}

func CloseOrders(p *AppParams) {
	for _, o := range p.TO.CloseOrders {
		exo, err := p.Repo.OpenStopOrder(o)
		if err != nil || exo == nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		o.RefID = exo.RefID
		o.OpenTime = exo.OpenTime
		err = p.DB.CreateOrder(o)
		if err != nil {
			h.Log(err)
			return
		}

		h.LogNewF(&o)
	}
}

func SyncNewLimitOrder(p *AppParams) {
	// Synchronize order status
	o := p.DB.GetHighestNewBuyOrder(p.QO)
	if o == nil {
		return
	}
	exo, err := p.Repo.GetOrder(*o)
	if err != nil || exo == nil {
		h.Log("GetOrder")
		os.Exit(1)
	}
	if exo.Status == t.OrderStatusNew {
		return
	}

	// Synchronize FILLED/CANCELED order
	if o.Status != exo.Status {
		o.Status = exo.Status
		o.UpdateTime = exo.UpdateTime
		err := p.DB.UpdateOrder(*o)
		if err != nil {
			h.Log("UpdateOrder", err)
			return
		}
		if exo.Status == t.OrderStatusFilled {
			h.LogFilled(o)
		}
		if exo.Status == t.OrderStatusCanceled {
			h.LogCanceled(o)
		}
	}
}

func SyncFilledLimitOrder(p *AppParams) {
	// Place a new Take Profit order
	o := p.DB.GetLowestFilledBuyOrder(p.QO)
	if o != nil && o.TPPrice > 0 && p.DB.GetTPOrder(o.ID) == nil {
		// Keep only one TP order, because Binance has a 'MAX_NUM_ALGO_ORDERS=5'
		tpOrders := p.DB.GetNewTPOrders(p.QO)
		if len(tpOrders) > 0 {
			tpo := tpOrders[0]
			// Ignore when the order TP price is so far, keep calm and waiting
			if tpo.OpenPrice < o.TPPrice {
				return
			}
			exo, err := p.Repo.CancelOrder(tpo)
			if err != nil || exo == nil {
				h.Log("CancelOrder")
				os.Exit(1)
			}

			tpo.Status = exo.Status
			tpo.UpdateTime = exo.UpdateTime
			err = p.DB.UpdateOrder(tpo)
			if err != nil {
				h.Log(err)
				return
			}
			h.LogCanceled(&tpo)
		}

		if p.Ticker.Price < h.CalcTPStop(o.Side, o.TPPrice, float64(p.BP.SLim.TPStop), p.BP.PriceDigits) {
			return
		}
		stopPrice := h.CalcTPStop(o.Side, o.TPPrice, float64(p.BP.SLim.TPLimit), p.BP.PriceDigits)

		// The price moves so fast
		if p.Ticker.Price > stopPrice && o.CloseOrderID == "" {
			o.CloseOrderID = "0"
			o.ClosePrice = o.TPPrice
			o.CloseTime = h.Now13()
			o.PL = h.NormalizeDouble(((o.ClosePrice - o.OpenPrice) * o.Qty), p.BP.PriceDigits)
			err := p.DB.UpdateOrder(*o)
			if err != nil {
				h.Log(err)
			}
			return
		}

		tpo := t.Order{
			BotID:       o.BotID,
			Exchange:    o.Exchange,
			Symbol:      o.Symbol,
			ID:          h.GenID(),
			OpenOrderID: o.ID,
			Qty:         o.Qty,
			Side:        h.Reverse(o.Side),
			Type:        t.OrderTypeTP,
			Status:      t.OrderStatusNew,
			StopPrice:   stopPrice,
			OpenPrice:   o.TPPrice,
		}
		exo, err := p.Repo.OpenStopOrder(tpo)
		if err != nil || exo == nil {
			h.Log("PlaceTPOrder")
			os.Exit(1)
		}

		tpo.RefID = exo.RefID
		tpo.OpenTime = exo.OpenTime
		err = p.DB.CreateOrder(tpo)
		if err != nil {
			h.Log(err)
			return
		}
		h.LogNew(&tpo)
	}
}

func SyncTPLimitOrder(p *AppParams) {
	tpo := p.DB.GetLowestTPOrder(p.QO)
	if tpo == nil {
		return
	}
	exo, err := p.Repo.GetOrder(*tpo)
	if err != nil || exo == nil {
		h.Log("GetTPOrder")
		os.Exit(1)
	}
	if exo.Status == t.OrderStatusNew {
		return
	}

	if tpo.Status != exo.Status {
		tpo.Status = exo.Status
		tpo.UpdateTime = exo.UpdateTime
		err := p.DB.UpdateOrder(*tpo)
		if err != nil {
			h.Log(err)
			return
		}
		if exo.Status == t.OrderStatusCanceled {
			h.LogCanceled(tpo)
			return
		}
	}
	if exo.Status == t.OrderStatusCanceled {
		return
	}

	oo := p.DB.GetOrderByID(tpo.OpenOrderID)
	if oo != nil && oo.CloseOrderID == "" && p.Ticker.Price > tpo.OpenPrice {
		oo.CloseOrderID = tpo.ID
		oo.ClosePrice = tpo.OpenPrice
		oo.CloseTime = h.Now13()
		oo.PL = h.NormalizeDouble(((oo.ClosePrice-oo.OpenPrice)*tpo.Qty)-oo.Commission-tpo.Commission, p.BP.PriceDigits)
		err := p.DB.UpdateOrder(*oo)
		if err != nil {
			h.Log(err)
			return
		}
		h.LogClosed(oo, tpo)

		tpo.CloseTime = oo.CloseTime
		err = p.DB.UpdateOrder(*tpo)
		if err != nil {
			h.Log("UpdateTPOrder", err)
		}
	}
}

func SyncSLLongOrders(p *AppParams) {
	slo := p.DB.GetHighestSLLongOrder(p.QO)
	if slo == nil {
		return
	}
	exo, err := p.Repo.GetOrder(*slo)
	if err != nil || exo == nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if exo.Status == t.OrderStatusNew {
		return
	}

	if slo.Status != exo.Status {
		slo.Status = exo.Status
		slo.UpdateTime = exo.UpdateTime
		err := p.DB.UpdateOrder(*slo)
		if err != nil {
			h.Log(err)
			return
		}

		if exo.Status == t.OrderStatusFilled {
			h.LogFilledF(slo)
		}

		if exo.Status == t.OrderStatusCanceled {
			h.LogCanceledF(slo)
		}
	}

	if p.Ticker.Price < slo.OpenPrice && slo.CloseTime == 0 {
		o := p.DB.GetOrderByID(slo.OpenOrderID)
		if o == nil {
			slo.CloseTime = h.Now13()
			err = p.DB.UpdateOrder(*slo)
			if err != nil {
				h.Log(err)
			}
			return
		}

		o.CloseOrderID = slo.ID
		o.ClosePrice = slo.OpenPrice
		o.CloseTime = h.Now13()
		o.PL = h.NormalizeDouble(((o.ClosePrice-o.OpenPrice)*slo.Qty)-o.Commission-slo.Commission, p.BP.PriceDigits)
		err := p.DB.UpdateOrder(*o)
		if err != nil {
			h.Log(err)
			return
		}

		slo.CloseTime = o.CloseTime
		err = p.DB.UpdateOrder(*slo)
		if err != nil {
			h.Log(err)
			return
		}

		h.LogClosedF(o, slo)
	}
}

func SyncSLShortOrders(p *AppParams) {
	slo := p.DB.GetLowestSLShortOrder(p.QO)
	if slo == nil {
		return
	}
	exo, err := p.Repo.GetOrder(*slo)
	if err != nil || exo == nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if exo.Status == t.OrderStatusNew {
		return
	}

	if slo.Status != exo.Status {
		slo.Status = exo.Status
		slo.UpdateTime = exo.UpdateTime
		err := p.DB.UpdateOrder(*slo)
		if err != nil {
			h.Log(err)
			return
		}

		if exo.Status == t.OrderStatusFilled {
			h.LogFilledF(slo)
		}

		if exo.Status == t.OrderStatusCanceled {
			h.LogCanceledF(slo)
		}
	}

	if p.Ticker.Price > slo.OpenPrice && slo.CloseTime == 0 {
		o := p.DB.GetOrderByID(slo.OpenOrderID)
		if o == nil {
			slo.CloseTime = h.Now13()
			err = p.DB.UpdateOrder(*slo)
			if err != nil {
				h.Log(err)
			}
			return
		}

		o.CloseOrderID = slo.ID
		o.ClosePrice = slo.OpenPrice
		o.CloseTime = h.Now13()
		o.PL = h.NormalizeDouble(((o.OpenPrice-o.ClosePrice)*slo.Qty)-o.Commission-slo.Commission, p.BP.PriceDigits)
		err := p.DB.UpdateOrder(*o)
		if err != nil {
			h.Log(err)
			return
		}

		slo.CloseTime = o.CloseTime
		err = p.DB.UpdateOrder(*slo)
		if err != nil {
			h.Log(err)
			return
		}

		h.LogClosedF(o, slo)
	}
}

func SyncTPLongOrders(p *AppParams) {
	tpo := p.DB.GetLowestTPLongOrder(p.QO)
	if tpo == nil {
		return
	}
	exo, err := p.Repo.GetOrder(*tpo)
	if err != nil || exo == nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if exo.Status == t.OrderStatusNew {
		return
	}

	if tpo.Status != exo.Status {
		tpo.Status = exo.Status
		tpo.UpdateTime = exo.UpdateTime
		err := p.DB.UpdateOrder(*tpo)
		if err != nil {
			h.Log(err)
			return
		}

		if exo.Status == t.OrderStatusFilled {
			h.LogFilledF(tpo)
		}

		if exo.Status == t.OrderStatusCanceled {
			h.LogCanceledF(tpo)
		}
	}

	if p.Ticker.Price > tpo.OpenPrice && tpo.CloseTime == 0 {
		o := p.DB.GetOrderByID(tpo.OpenOrderID)
		if o == nil {
			tpo.CloseTime = h.Now13()
			err = p.DB.UpdateOrder(*tpo)
			if err != nil {
				h.Log(err)
			}
			return
		}

		o.CloseOrderID = tpo.ID
		o.ClosePrice = tpo.OpenPrice
		o.CloseTime = h.Now13()
		o.PL = h.NormalizeDouble(((o.ClosePrice-o.OpenPrice)*tpo.Qty)-o.Commission-tpo.Commission, p.BP.PriceDigits)
		err := p.DB.UpdateOrder(*o)
		if err != nil {
			h.Log(err)
			return
		}

		tpo.CloseTime = o.CloseTime
		err = p.DB.UpdateOrder(*tpo)
		if err != nil {
			h.Log(err)
			return
		}

		h.LogClosedF(o, tpo)
	}
}

func SyncTPShortOrders(p *AppParams) {
	tpo := p.DB.GetHighestTPShortOrder(p.QO)
	if tpo == nil {
		return
	}
	exo, err := p.Repo.GetOrder(*tpo)
	if err != nil || exo == nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if exo.Status == t.OrderStatusNew {
		return
	}

	if tpo.Status != exo.Status {
		tpo.Status = exo.Status
		tpo.UpdateTime = exo.UpdateTime
		err := p.DB.UpdateOrder(*tpo)
		if err != nil {
			h.Log(err)
			return
		}

		if exo.Status == t.OrderStatusFilled {
			h.LogFilledF(tpo)
		}

		if exo.Status == t.OrderStatusCanceled {
			h.LogCanceledF(tpo)
		}
	}

	if p.Ticker.Price < tpo.OpenPrice && tpo.CloseTime == 0 {
		o := p.DB.GetOrderByID(tpo.OpenOrderID)
		if o == nil {
			tpo.CloseTime = h.Now13()
			err = p.DB.UpdateOrder(*tpo)
			if err != nil {
				h.Log(err)
			}
			return
		}

		o.CloseOrderID = tpo.ID
		o.ClosePrice = tpo.OpenPrice
		o.CloseTime = h.Now13()
		o.PL = h.NormalizeDouble(((o.OpenPrice-o.ClosePrice)*tpo.Qty)-o.Commission-tpo.Commission, p.BP.PriceDigits)
		err := p.DB.UpdateOrder(*o)
		if err != nil {
			h.Log(err)
			return
		}

		tpo.CloseTime = o.CloseTime
		err = p.DB.UpdateOrder(*tpo)
		if err != nil {
			h.Log(err)
			return
		}

		h.LogClosedF(o, tpo)
	}
}

func SyncLimitLongOrders(p *AppParams) {
	o := p.DB.GetHighestNewLongOrder(p.QO)
	if o == nil {
		return
	}
	exo, err := p.Repo.GetOrder(*o)
	if err != nil || exo == nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if exo.Status == t.OrderStatusNew {
		return
	}

	if o.Status != exo.Status {
		o.Status = exo.Status
		o.UpdateTime = exo.UpdateTime
		err := p.DB.UpdateOrder(*o)
		if err != nil {
			h.Log(err)
			return
		}
		if exo.Status == t.OrderStatusFilled {
			h.LogFilledF(o)
		}
		if exo.Status == t.OrderStatusCanceled {
			h.LogCanceledF(o)
		}
	}
}

func SyncLimitShortOrders(p *AppParams) {
	o := p.DB.GetLowestNewShortOrder(p.QO)
	if o == nil {
		return
	}
	exo, err := p.Repo.GetOrder(*o)
	if err != nil || exo == nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if exo.Status == t.OrderStatusNew {
		return
	}

	if o.Status != exo.Status {
		o.Status = exo.Status
		o.UpdateTime = exo.UpdateTime
		err := p.DB.UpdateOrder(*o)
		if err != nil {
			h.Log(err)
			return
		}
		if exo.Status == t.OrderStatusFilled {
			h.LogFilledF(o)
		}
		if exo.Status == t.OrderStatusCanceled {
			h.LogCanceledF(o)
		}
	}
}

func SyncClosedOrders(p *AppParams) {}
