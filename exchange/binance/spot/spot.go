package spot

import (
	"fmt"
	"log"
	"strings"

	"github.com/tidwall/gjson"

	"github.com/tonkla/autotp/helper"
	"github.com/tonkla/autotp/types"
)

const (
	urlBase = "https://api.binance.com/api/v3"
	// urlTest   = "https://testnet.binance.vision/api/v3"
	pathDepth    = "/depth?symbol=%s&limit=%d"
	pathHisPrice = "/klines?symbol=%s&interval=%s&limit=%d"
	pathTicker   = "/ticker/24hr?symbol=%s"
)

type Binance struct {
}

func New() Binance {
	return Binance{}
}

func sanitizeSymbol(symbol string) string {
	s := strings.ToUpper(symbol)
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "_", "")
	return s
}

// GetName returns "BINANCE"
func (b Binance) GetName() string {
	return types.EXC_BINANCE
}

// GetTicker returns the latest ticker of the symbol
func (b Binance) GetTicker(symbol string) types.Ticker {
	_symbol := sanitizeSymbol(symbol)
	path := fmt.Sprintf(pathTicker, _symbol)
	url := fmt.Sprintf("%s%s", urlBase, path)
	data, err := helper.Get(url)
	if err != nil {
		log.Println(err)
		return types.Ticker{}
	}

	r := gjson.Parse(string(data))
	return types.Ticker{
		Symbol: _symbol,
		Price:  r.Get("lastPrice").Float(),
		Qty:    r.Get("lastQty").Float(),
	}
}

// GetHistoricalPrices returns a list of k-lines/candlesticks of the symbol
func (b Binance) GetHistoricalPrices(symbol string, interval string, limit int) []types.HisPrice {
	_symbol := sanitizeSymbol(symbol)
	path := fmt.Sprintf(pathHisPrice, _symbol, interval, limit)
	url := fmt.Sprintf("%s%s", urlBase, path)
	data, err := helper.Get(url)
	if err != nil {
		log.Println(err)
		return nil
	}

	var hPrices []types.HisPrice
	for _, data := range gjson.Parse(string(data)).Array() {
		d := data.Array()
		p := types.HisPrice{
			Symbol: _symbol,
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
func (b Binance) GetOrderBook(symbol string, limit int) types.OrderBook {
	_symbol := sanitizeSymbol(symbol)
	path := fmt.Sprintf(pathDepth, _symbol, limit)
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
		Exchange: types.Exchange{Name: types.EXC_BINANCE},
		Symbol:   _symbol,
		Bids:     bids,
		Asks:     asks}
}

func (b Binance) GetOpenOrders() []types.Order {
	return []types.Order{}
}

func (b Binance) GetOrderHistory() []types.Order {
	return []types.Order{}
}

func (b Binance) OpenOrder(order types.Order) *types.Order {
	url := ""
	data := ""
	helper.Post(url, data)
	return nil
}

func (b Binance) CloseOrder(order types.Order) *types.Order {
	url := ""
	data := ""
	helper.Post(url, data)
	return nil
}

func (b Binance) CloseOrderByID(id string) *types.Order {
	return nil
}
