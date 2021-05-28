package fusd

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/tidwall/gjson"
	"github.com/tonkla/autotp/helper"
	"github.com/tonkla/autotp/types"
)

const (
	urlBase    = "https://fapi.binance.com/fapi/v1"
	pathTicker = "/ticker/price"
	pathTrade  = "/order"
	pathKlines = "/klines"
)

// GetTicker returns the latest ticker
func GetTicker(symbol string) *types.Ticker {
	url := fmt.Sprintf("%s%s?symbol=%s", urlBase, pathTicker, symbol)
	data, err := helper.Get(url)
	if err != nil {
		return nil
	}
	r := gjson.Parse(string(data))
	return &types.Ticker{
		Exchange: types.EXC_BINANCE,
		Symbol:   r.Get("symbol").String(),
		Price:    r.Get("price").Float(),
		Time:     r.Get("time").Int(),
	}
}

func GetOpenOrders(symbol string) []types.Order {
	return []types.Order{}
}

func GetOrderHistory(symbol string) []types.Order {
	return []types.Order{}
}

// GetHistoricalPrices returns a list of k-lines/candlesticks
func GetHistoricalPrices(symbol string, interval string, limit int) []types.HistoricalPrice {
	url := fmt.Sprintf("%s%s?symbol=%s&interval=%s&limit=%d", urlBase, pathKlines, symbol, interval, limit)
	data, err := helper.Get(url)
	if err != nil {
		log.Println(err)
		return nil
	}

	var hPrices []types.HistoricalPrice
	for _, data := range gjson.Parse(string(data)).Array() {
		d := data.Array()
		p := types.HistoricalPrice{
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

func Trade(order *types.Order) *types.Order {
	url := fmt.Sprintf("%s%s", urlBase, pathTrade)
	data, err := json.Marshal(order)
	if err != nil {
		return nil
	}
	_, err = helper.Post(url, string(data))
	if err != nil {
		return nil
	}
	return order
}
