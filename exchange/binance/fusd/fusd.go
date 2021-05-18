package fusd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/tonkla/autotp/helper"
	"github.com/tonkla/autotp/types"
)

const (
	urlBase    = "https://fapi.binance.com/fapi/v1"
	pathTicker = "/ticker"
	pathTrade  = "/order"
)

func GetTicker(symbol string) types.Ticker {
	url := fmt.Sprintf("%s%s", urlBase, pathTicker)
	helper.Get(url)
	return types.Ticker{
		Exchange: types.Exchange{Name: types.EXC_BINANCE},
		Symbol:   symbol,
		Price:    0,
		Qty:      0,
	}
}

func GetOpenOrders() {
}

func GetOrderHistory() {
}

func Trade(order types.Order) *types.TradeResult {
	url := fmt.Sprintf("%s%s", urlBase, pathTrade)
	data, e := json.Marshal(order)
	if e != nil {
		return nil
	}
	isSucceeded := helper.Post(url, string(data))
	if isSucceeded {
		return &types.TradeResult{Time: time.Now().Unix()}
	}
	return nil
}
