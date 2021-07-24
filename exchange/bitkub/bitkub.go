package bitkub

import (
	"fmt"
	"log"
	"strings"

	"github.com/tidwall/gjson"

	"github.com/tonkla/autotp/helper"
	"github.com/tonkla/autotp/types"
)

type Client struct {
	baseURL string
}

func NewClient() Client {
	return Client{
		baseURL: "https://api.bitkub.com/api",
	}
}

// GetTicker returns the latest ticker
func (c Client) GetTicker(symbol string) *types.Ticker {
	var url strings.Builder
	fmt.Fprintf(&url, "%s/market/ticker?sym=%s", c.baseURL, symbol)
	data, err := helper.Get(url.String())
	if err != nil {
		return nil
	}
	ticker := gjson.ParseBytes(data).Get(strings.ToUpper(symbol))
	return &types.Ticker{
		Symbol: symbol,
		Price:  ticker.Get("last").Float(),
	}
}

// GetHistoricalPrices returns historical prices in a format of Klines/Candlesticks
func (c Client) GetHistoricalPrices(symbol string, interval int) []types.HistoricalPrice {
	return []types.HistoricalPrice{}
}

// GetOrderBook returns an order book (market depth)
func (c Client) GetOrderBook(symbol string, limit int) types.OrderBook {
	var url strings.Builder
	fmt.Fprintf(&url, "%s/market/depth?sym=%s&lmt=%d", c.baseURL, symbol, limit)
	data, err := helper.Get(url.String())
	if err != nil {
		log.Println(err)
		return types.OrderBook{}
	}

	orders := gjson.ParseBytes(data)

	var bids []types.ExOrder
	for _, bid := range orders.Get("bids").Array() {
		b := bid.Array()
		ord := types.ExOrder{
			Side:  types.OrderSideBuy,
			Price: b[0].Float(),
			Qty:   b[1].Float(),
		}
		bids = append(bids, ord)
	}

	var asks []types.ExOrder
	for _, ask := range orders.Get("asks").Array() {
		a := ask.Array()
		ord := types.ExOrder{
			Side:  types.OrderSideSell,
			Price: a[0].Float(),
			Qty:   a[1].Float(),
		}
		asks = append(asks, ord)
	}

	return types.OrderBook{
		Symbol: symbol,
		Bids:   bids,
		Asks:   asks,
	}
}
