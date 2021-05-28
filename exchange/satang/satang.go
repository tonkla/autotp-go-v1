package satang

import (
	"fmt"
	"log"

	"github.com/tidwall/gjson"

	"github.com/tonkla/autotp/helper"
	"github.com/tonkla/autotp/types"
)

const (
	urlBase      = "https://satangcorp.com/api/v3"
	pathDepth    = "/depth?symbol=%s&limit=%d"
	pathHisPrice = "/klines?symbol=%s&interval=%s&limit=%d"
	pathTicker   = "/ticker/24hr?symbol=%s"
)

type Satang struct {
}

func New() *Satang {
	return &Satang{}
}

// GetName returns "SATANG"
func (s Satang) GetName() string {
	return types.EXC_SATANG
}

// GetTicker returns the latest ticker of the symbol
func (s Satang) GetTicker(symbol string) types.Ticker {
	path := fmt.Sprintf(pathTicker, symbol)
	url := fmt.Sprintf("%s%s", urlBase, path)
	data, err := helper.Get(url)
	if err != nil {
		log.Println(err)
		return types.Ticker{}
	}
	r := gjson.Parse(string(data))
	return types.Ticker{
		Symbol: symbol,
		Price:  r.Get("lastPrice").Float(),
		Qty:    r.Get("lastQty").Float(),
	}
}

// GetHistoricalPrices returns a list of k-lines/candlesticks of the symbol
func (s Satang) GetHistoricalPrices(symbol string, interval string, limit int) []types.HistoricalPrice {
	path := fmt.Sprintf(pathHisPrice, symbol, interval, limit)
	url := fmt.Sprintf("%s%s", urlBase, path)
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

// GetOrderBook returns an order book of the symbol
func (s Satang) GetOrderBook(symbol string, limit int) types.OrderBook {
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
		Exchange: types.EXC_SATANG,
		Symbol:   symbol,
		Bids:     bids,
		Asks:     asks}
}
