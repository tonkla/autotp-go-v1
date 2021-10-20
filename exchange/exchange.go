package exchange

import (
	"errors"

	bs "github.com/tonkla/autotp/exchange/binance"
	bf "github.com/tonkla/autotp/exchange/binance/futures"
	t "github.com/tonkla/autotp/types"
)

type Repository interface {
	GetHistoricalPrices(string, string, int) []t.HistoricalPrice
	GetOrder(t.Order) (*t.Order, error)
	GetOrderBook(string, int) *t.OrderBook
	GetTicker(string) *t.Ticker
	OpenLimitOrder(t.Order) (*t.Order, error)
	OpenMarketOrder(t.Order) (*t.Order, error)
	OpenStopOrder(t.Order) (*t.Order, error)
	CancelOrder(t.Order) (*t.Order, error)
	CloseOrder(t.Order) (*t.Order, error)
}

func New(bp *t.BotParams) (Repository, error) {
	if bp.Exchange == t.ExcBinance {
		if bp.Product == t.ProductSpot {
			return bs.NewSpotClient("", ""), nil
		} else if bp.Product == t.ProductFutures {
			return bf.NewFuturesClient("", ""), nil
		}
	}
	return nil, errors.New("exchange not found")
}
