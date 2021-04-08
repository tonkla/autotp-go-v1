package bitkub

import (
	"fmt"
	"log"

	"github.com/tidwall/gjson"

	"github.com/tonkla/autotp/common"
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
	return common.EXC_BITKUB
}

// GetTicker returns the latest ticker of the symbol
func (b Bitkub) GetTicker(symbol string) common.Ticker {
	return common.Ticker{}
}

// GetHistoricalPrices returns a list of k-lines/candlesticks of the symbol
func (b Bitkub) GetHistoricalPrices(symbol string, interval string, limit int) []common.HisPrice {
	return []common.HisPrice{}
}

// GetOrderBook returns an order book of the symbol
func (b Bitkub) GetOrderBook(symbol string, limit int) common.OrderBook {
	path := fmt.Sprintf(pathDepth, symbol, limit)
	url := fmt.Sprintf("%s%s", urlBase, path)

	data, err := common.Get(url)
	if err != nil {
		log.Println(err)
		return common.OrderBook{}
	}

	orders := gjson.Parse(string(data))

	var bids []common.Order
	for _, bid := range orders.Get("bids").Array() {
		b := bid.Array()
		ord := common.Order{
			Side:     "BUY",
			Price:    b[0].Float(),
			Quantity: b[1].Float()}
		bids = append(bids, ord)
	}

	var asks []common.Order
	for _, ask := range orders.Get("asks").Array() {
		a := ask.Array()
		ord := common.Order{
			Side:     "SELL",
			Price:    a[0].Float(),
			Quantity: a[1].Float()}
		asks = append(asks, ord)
	}

	return common.OrderBook{
		Exchange: common.Exchange{Name: common.EXC_BITKUB},
		Symbol:   symbol,
		Bids:     bids,
		Asks:     asks}
}
