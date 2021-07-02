package fusd

import (
	"github.com/tonkla/autotp/exchange/binance"
	"github.com/tonkla/autotp/types"
)

const (
	urlBase      = "https://fapi.binance.com/fapi/v1"
	pathDepth    = "/depth"
	pathKlines   = "/klines"
	pathNewOrder = "/order"
	pathTicker   = "/ticker/price"
)

// GetTicker returns the latest ticker
func GetTicker(symbol string) *types.Ticker {
	return binance.GetTicker(urlBase, pathTicker, symbol)
}

// GetOrderBook returns an order book (market depth)
func GetOrderBook(symbol string, limit int) *types.OrderBook {
	return binance.GetOrderBook(urlBase, pathDepth, symbol, limit)
}

func GetOrder(symbol string, id int) *types.Order {
	return nil
}

func GetOpenOrders(symbol string) []types.Order {
	return []types.Order{}
}

func GetOrderHistory(symbol string) []types.Order {
	return []types.Order{}
}

// GetHistoricalPrices returns a list of k-lines/candlesticks
func GetHistoricalPrices(symbol string, timeframe string, limit int) []types.HistoricalPrice {
	return binance.GetHistoricalPrices(urlBase, pathKlines, symbol, timeframe, limit)
}

// NewOrder sends an order to trade on the exchange
func NewOrder(order types.Order, apiKey string, secretKey string) *types.Order {
	return binance.NewOrder(urlBase, pathNewOrder, order, apiKey, secretKey)
}
