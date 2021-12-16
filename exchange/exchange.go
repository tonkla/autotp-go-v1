package exchange

import (
	"errors"

	bf "github.com/tonkla/autotp/exchange/binance/futures"
	bs "github.com/tonkla/autotp/exchange/binance/spot"
	t "github.com/tonkla/autotp/types"
)

type Repository interface {
	GetHistoricalPrices(symbol string, timeframe string, limit int) []t.HistoricalPrice
	Get1wHistoricalPrices(symbol string, limit int) []t.HistoricalPrice
	Get1dHistoricalPrices(symbol string, limit int) []t.HistoricalPrice
	Get4hHistoricalPrices(symbol string, limit int) []t.HistoricalPrice
	Get1hHistoricalPrices(symbol string, limit int) []t.HistoricalPrice
	Get15mHistoricalPrices(symbol string, limit int) []t.HistoricalPrice

	CountOpenOrders(symbol string) (int, error)
	GetOrder(t.Order) (*t.Order, error)
	GetOpenOrders(symbol string) []t.Order
	GetTradeList(symbol string, limit int, startTime int, endTime int) ([]t.TradeOrder, error)
	GetAllOrders(symbol string, limit int, startTime int, endTime int) []t.Order
	GetCommission(symbol string, orderRefID string) *float64
	GetOrderBook(symbol string, limit int) *t.OrderBook
	GetTicker(symbol string) *t.Ticker
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
