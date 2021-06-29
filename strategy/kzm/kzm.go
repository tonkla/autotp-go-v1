package kzm

import (
	"github.com/tonkla/autotp/db"
	"github.com/tonkla/autotp/strategy/grid"
	"github.com/tonkla/autotp/strategy/trend"
	"github.com/tonkla/autotp/types"
)

type OnTickParams struct {
	Ticker    types.Ticker
	OrderBook types.OrderBook
	BotParams types.BotParams
	HPrices   []types.HistoricalPrice
	DB        db.DB
}

func OnTick(params OnTickParams) *types.TradeOrders {
	ticker := params.Ticker
	odbook := params.OrderBook
	p := params.BotParams
	prices := params.HPrices
	db := params.DB

	var openOrders, closeOrders []types.Order

	orders := grid.OnTick(grid.OnTickParams{Ticker: ticker, BotParams: p, DB: db})
	if orders != nil {
		openOrders = append(openOrders, orders.OpenOrders...)
		closeOrders = append(closeOrders, orders.CloseOrders...)
	}

	orders = trend.OnTick(trend.OnTickParams{
		Ticker: ticker, BotParams: p, OrderBook: odbook, HPrices: prices, DB: db})
	if orders != nil {
		openOrders = append(openOrders, orders.OpenOrders...)
		closeOrders = append(openOrders, orders.OpenOrders...)
	}

	return &types.TradeOrders{
		OpenOrders:  openOrders,
		CloseOrders: closeOrders,
	}
}
