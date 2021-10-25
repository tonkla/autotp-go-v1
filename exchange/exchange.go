package exchange

import (
	"errors"

	bf "github.com/tonkla/autotp/exchange/binance/futures"
	bs "github.com/tonkla/autotp/exchange/binance/spot"
	t "github.com/tonkla/autotp/types"
)

type Repository interface {
	GetHistoricalPrices(string, string, int) []t.HistoricalPrice
	Get1wHistoricalPrices(string, int) []t.HistoricalPrice
	Get1dHistoricalPrices(string, int) []t.HistoricalPrice
	Get4hHistoricalPrices(string, int) []t.HistoricalPrice
	Get1hHistoricalPrices(string, int) []t.HistoricalPrice
	Get15mHistoricalPrices(string, int) []t.HistoricalPrice

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
			return bs.NewSpotClient(bp.ApiKey, bp.SecretKey), nil
		} else if bp.Product == t.ProductFutures {
			return bf.NewFuturesClient(bp.ApiKey, bp.SecretKey), nil
		}
	}
	return nil, errors.New("exchange not found")
}
