package satang

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
		baseURL: "https://satangcorp.com/api/v3",
	}
}

// GetTicker returns the latest ticker
func (c Client) GetTicker(symbol string) types.Ticker {
	var url strings.Builder
	fmt.Fprintf(&url, "%s/ticker/24hr?symbol=%s", c.baseURL, symbol)
	data, err := helper.Get(url.String())
	if err != nil {
		log.Println(err)
		return types.Ticker{}
	}
	r := gjson.ParseBytes(data)
	return types.Ticker{
		Symbol: symbol,
		Price:  r.Get("lastPrice").Float(),
		Qty:    r.Get("lastQty").Float(),
	}
}

// GetHistoricalPrices returns historical prices in a format of Klines/Candlesticks
func (c Client) GetHistoricalPrices(symbol string, interval string, limit int) []types.HistoricalPrice {
	var url strings.Builder
	fmt.Fprintf(&url, "%s/klines?symbol=%s&interval=%s&limit=%d", c.baseURL, symbol, interval, limit)
	data, err := helper.Get(url.String())
	if err != nil {
		log.Println(err)
		return nil
	}

	var hprices []types.HistoricalPrice
	for _, data := range gjson.ParseBytes(data).Array() {
		d := data.Array()
		p := types.HistoricalPrice{
			Symbol: symbol,
			Time:   d[0].Int() / 1000,
			Open:   d[1].Float(),
			High:   d[2].Float(),
			Low:    d[3].Float(),
			Close:  d[4].Float(),
		}
		hprices = append(hprices, p)
	}
	return hprices
}

// GetOrderBook returns an order book (market depth)
func (c Client) GetOrderBook(symbol string, limit int) types.OrderBook {
	var url strings.Builder
	fmt.Fprintf(&url, "%s/depth?symbol=%s&limit=%d", c.baseURL, symbol, limit)
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
			Qty:   b[1].Float()}
		bids = append(bids, ord)
	}

	var asks []types.ExOrder
	for _, ask := range orders.Get("asks").Array() {
		a := ask.Array()
		ord := types.ExOrder{
			Side:  types.OrderSideSell,
			Price: a[0].Float(),
			Qty:   a[1].Float()}
		asks = append(asks, ord)
	}

	return types.OrderBook{
		Exchange: types.ExcSatang,
		Symbol:   symbol,
		Bids:     bids,
		Asks:     asks}
}
