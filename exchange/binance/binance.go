package binance

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/tidwall/gjson"

	"github.com/tonkla/autotp/helper"
	t "github.com/tonkla/autotp/types"
)

// Sign signs a payload with a Binance API secret key
func Sign(payload string, secretKey string) string {
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

// GetTicker returns the latest ticker
func GetTicker(base string, path string, symbol string) *t.Ticker {
	var url strings.Builder
	fmt.Fprintf(&url, "%s%s?symbol=%s", base, path, symbol)
	data, err := helper.Get(url.String())
	if err != nil {
		return nil
	}
	r := gjson.ParseBytes(data)
	return &t.Ticker{
		Exchange: t.EXC_BINANCE,
		Symbol:   r.Get("symbol").String(),
		Price:    r.Get("price").Float(),
		Time:     r.Get("time").Int(),
	}
}

// GetOrderBook returns an order book (market depth)
func GetOrderBook(base string, path string, symbol string, limit int) *t.OrderBook {
	var url strings.Builder
	fmt.Fprintf(&url, "%s%s?symbol=%s&limit=%d", base, path, symbol, limit)
	data, err := helper.Get(url.String())
	if err != nil {
		return nil
	}

	var bids, asks []t.ExOrder
	result := gjson.ParseBytes(data)
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

func GetOrder(base string, path string, symbol string, id int) *t.Order {
	return nil
}

func GetOpenOrders(base string, path string, symbol string) []t.Order {
	return []t.Order{}
}

func GetOrderHistory(base string, path string, symbol string) []t.Order {
	return []t.Order{}
}

// GetHistoricalPrices returns a list of k-lines/candlesticks
func GetHistoricalPrices(base string, path string, symbol string, timeframe string, limit int) []t.HistoricalPrice {
	var url strings.Builder
	fmt.Fprintf(&url, "%s%s?symbol=%s&interval=%s&limit=%d", base, path, symbol, timeframe, limit)
	data, err := helper.Get(url.String())
	if err != nil {
		return nil
	}

	var hPrices []t.HistoricalPrice
	for _, data := range gjson.ParseBytes(data).Array() {
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

// NewOrder sends an order to trade on the exchange
func NewOrder(base string, path string, o t.Order, apiKey string, secretKey string) *t.Order {
	var payload strings.Builder
	fmt.Fprintf(&payload,
		"symbol=%s&side=%s&type=LIMIT&quantity=%f&price=%f&timestamp=%d",
		o.Symbol, o.Side, o.Qty, o.OpenPrice, time.Now().Unix())

	signature := Sign(payload.String(), secretKey)

	var url strings.Builder
	fmt.Fprintf(&url, "%s%s", base, path)
	fmt.Fprintf(&url, "?%s&signature=%s", payload, signature)
	url.WriteString(payload.String())
	resp, err := helper.Post(url.String(), newHeader(apiKey))
	if err != nil {
		return nil
	}
	fmt.Println("Response:", resp)
	return &o
}

func newHeader(apiKey string) http.Header {
	var header http.Header
	header.Set("X-MBX-APIKEY", apiKey)
	return header
}
