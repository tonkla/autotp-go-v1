package exchange

import (
	t "github.com/tonkla/autotp/types"
)

type Repository interface {
	GetOrder(t.Order) (*t.Order, error)
	GetOrderBook(string, int) *t.OrderBook
	OpenLimitOrder(t.Order) (*t.Order, error)
	OpenMarketOrder(t.Order) (*t.Order, error)
	OpenStopOrder(t.Order) (*t.Order, error)
	CloseOrder(t.Order) (*t.Order, error)
	CancelOrder(t.Order) (*t.Order, error)
}
