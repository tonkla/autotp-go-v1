package fusd

import (
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/tonkla/autotp/helper"
	t "github.com/tonkla/autotp/types"
)

const (
	urlBase    = "https://fapi.binance.com/fapi/v1"
	pathTicker = "/ticker/price"
	pathDepth  = "/depth"
	pathKlines = "/klines"
	pathTrade  = "/order"
)

// GetTicker returns the latest ticker
func GetTicker(symbol string) *t.Ticker {
	url := fmt.Sprintf("%s%s?symbol=%s", urlBase, pathTicker, symbol)
	data, err := helper.Get(url)
	if err != nil {
		return nil
	}
	r := gjson.Parse(string(data))
	return &t.Ticker{
		Exchange: t.EXC_BINANCE,
		Symbol:   r.Get("symbol").String(),
		Price:    r.Get("price").Float(),
		Time:     r.Get("time").Int(),
	}
}

// GetOrderBook returns an order book (market depth)
func GetOrderBook(symbol string, limit int) *t.OrderBook {
	url := fmt.Sprintf("%s%s?symbol=%s&limit=%d", urlBase, pathDepth, symbol, limit)
	data, err := helper.Get(url)
	if err != nil {
		return nil
	}

	var bids, asks []t.ExOrder
	result := gjson.Parse(string(data))
	for _, bid := range result.Get("bids").Array() {
		b := bid.Array()
		bids = append(bids, t.ExOrder{
			Price: b[0].Float(),
			Qty:   b[1].Float(),
		})
	}
	for _, ask := range result.Get("asks").Array() {
		a := ask.Array()
		asks = append(asks, t.ExOrder{
			Price: a[0].Float(),
			Qty:   a[1].Float(),
		})
	}
	return &t.OrderBook{
		Exchange: t.EXC_BINANCE,
		Symbol:   symbol,
		Bids:     bids,
		Asks:     asks,
	}
}

func GetOpenOrders(symbol string) []t.Order {
	return []t.Order{}
}

func GetOrderHistory(symbol string) []t.Order {
	return []t.Order{}
}

// GetHistoricalPrices returns a list of k-lines/candlesticks
func GetHistoricalPrices(symbol string, timeframe string, limit int) []t.HistoricalPrice {
	url := fmt.Sprintf("%s%s?symbol=%s&interval=%s&limit=%d", urlBase, pathKlines, symbol, timeframe, limit)
	data, err := helper.Get(url)
	if err != nil {
		return nil
	}

	var hPrices []t.HistoricalPrice
	for _, data := range gjson.Parse(string(data)).Array() {
		d := data.Array()
		p := t.HistoricalPrice{
			Symbol: symbol,
			Time:   d[0].Int() / 1000,
			Open:   d[1].Float(),
			High:   d[2].Float(),
			Low:    d[3].Float(),
			Close:  d[4].Float(),
		}
		hPrices = append(hPrices, p)
	}
	return hPrices
}

func Trade(order t.Order) *t.Order {
	url := fmt.Sprintf("%s%s", urlBase, pathTrade)
	data, err := json.Marshal(order)
	if err != nil {
		return nil
	}
	_, err = helper.Post(url, string(data))
	if err != nil {
		return nil
	}
	return &order
}
