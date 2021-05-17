package bitkub

import (
	"fmt"
	"log"

	"github.com/tidwall/gjson"

	"github.com/tonkla/autotp/helper"
	"github.com/tonkla/autotp/types"
)

const (
	urlBase      = "https://api.bitkub.com/api"
	pathDepth    = "/market/depth?sym=%s&lmt=%d"
	pathHisPrice = "/market/tradingview?sym=%s&int=%d&frm=%d"
	pathTicker   = "/market/trades?sym=%s&lmt=%d"
)

type Bitkub struct {
}

func New() *Bitkub {
	return &Bitkub{}
}

// GetName returns "BITKUB"
func (b Bitkub) GetName() string {
	return types.EXC_BITKUB
}

// GetTicker returns the latest ticker of the symbol
func (b Bitkub) GetTicker(symbol string) types.Ticker {
	return types.Ticker{}
}

// GetHistoricalPrices returns a list of k-lines/candlesticks of the symbol
func (b Bitkub) GetHistoricalPrices(symbol string, interval string, limit int) []types.HisPrice {
	return []types.HisPrice{}
}

// GetOrderBook returns an order book of the symbol
func (b Bitkub) GetOrderBook(symbol string, limit int) types.OrderBook {
	path := fmt.Sprintf(pathDepth, symbol, limit)
	url := fmt.Sprintf("%s%s", urlBase, path)

	data, err := helper.Get(url)
	if err != nil {
		log.Println(err)
		return types.OrderBook{}
	}

	orders := gjson.Parse(string(data))

	var bids []types.Order
	for _, bid := range orders.Get("bids").Array() {
		b := bid.Array()
		ord := types.Order{
			Side:  "BUY",
			Price: b[0].Float(),
			Qty:   b[1].Float()}
		bids = append(bids, ord)
	}

	var asks []types.Order
	for _, ask := range orders.Get("asks").Array() {
		a := ask.Array()
		ord := types.Order{
			Side:  "SELL",
			Price: a[0].Float(),
			Qty:   a[1].Float()}
		asks = append(asks, ord)
	}

	return types.OrderBook{
		Exchange: types.Exchange{Name: types.EXC_BITKUB},
		Symbol:   symbol,
		Bids:     bids,
		Asks:     asks}
}
