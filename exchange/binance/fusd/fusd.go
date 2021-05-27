package fusd

import (
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/tonkla/autotp/helper"
	"github.com/tonkla/autotp/types"
)

const (
	urlBase    = "https://fapi.binance.com/fapi/v1"
	pathTicker = "/ticker/price"
	pathTrade  = "/order"
)

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

func GetOpenOrders() {
}

func GetOrderHistory() {
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
