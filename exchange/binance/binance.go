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

type Client struct {
	apiKey    string
	secretKey string
	baseURL   string
}

// NewSpotClient returns Binance Spot client
func NewSpotClient(apiKey string, secretKey string) Client {
	return Client{
		baseURL:   "https://api.binance.com/api/v3",
		apiKey:    apiKey,
		secretKey: secretKey,
	}
}

// NewFuturesClient returns Binance USDⓈ-M Futures client
func NewFuturesClient(apiKey string, secretKey string) Client {
	return Client{
		baseURL:   "https://fapi.binance.com/fapi/v1",
		apiKey:    apiKey,
		secretKey: secretKey,
	}
}

// Sign signs a payload with a Binance API secret key
func Sign(payload string, secretKey string) string {
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func newHeader(apiKey string) http.Header {
	var header http.Header
	header.Set("X-MBX-APIKEY", apiKey)
	return header
}

func now() int64 {
	return time.Now().UnixNano() / 1e6
}

// Public APIs -----------------------------------------------------------------

// GetTicker returns the latest ticker
func (c Client) GetTicker(symbol string) *t.Ticker {
	var url strings.Builder
	fmt.Fprintf(&url, "%s/ticker/price?symbol=%s", c.baseURL, symbol)
	data, err := helper.Get(url.String())
	if err != nil {
		return nil
	}
	r := gjson.ParseBytes(data)
	return &t.Ticker{
		Exchange: t.ExcBinance,
		Symbol:   r.Get("symbol").String(),
		Price:    r.Get("price").Float(),
		Time:     r.Get("time").Int(),
	}
}

// GetOrderBook returns an order book (market depth)
func (c Client) GetOrderBook(symbol string, limit int) *t.OrderBook {
	var url strings.Builder
	fmt.Fprintf(&url, "%s/depth?symbol=%s&limit=%d", c.baseURL, symbol, limit)
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
		Exchange: t.ExcBinance,
		Symbol:   symbol,
		Bids:     bids,
		Asks:     asks,
	}
}

// GetHistoricalPrices returns historical prices in a format of k-lines/candlesticks
func (c Client) GetHistoricalPrices(symbol string, timeframe string, limit int) []t.HistoricalPrice {
	var url strings.Builder
	fmt.Fprintf(&url, "%s/klines?symbol=%s&interval=%s&limit=%d", c.baseURL, symbol, timeframe, limit)
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

// Private APIs ----------------------------------------------------------------

// GetOrder returns the order by its ID
func (c Client) GetOrder(symbol string, id int) *t.Order {
	var payload strings.Builder
	fmt.Fprintf(&payload,
		"symbol=%s&orderId=%d&origClientOrderId=%s&timestamp=%d", symbol, id, "", now())

	signature := Sign(payload.String(), c.secretKey)

	var url strings.Builder
	fmt.Fprintf(&url, "%s/order?%s&signature=%s", c.baseURL, payload.String(), signature)
	data, err := helper.Get(url.String())
	if err != nil {
		return nil
	}
	r := gjson.ParseBytes(data)
	return &t.Order{
		Symbol: r.Get("symbol").String(),
	}
}

// GetOpenOrders returns open orders
func (c Client) GetOpenOrders(symbol string) []t.Order {
	var payload strings.Builder
	fmt.Fprintf(&payload, "symbol=%s&timestamp=%d", symbol, now())

	signature := Sign(payload.String(), c.secretKey)

	var url strings.Builder
	fmt.Fprintf(&url, "%s/openOrders?%s&signature=%s", c.baseURL, payload.String(), signature)
	data, err := helper.Get(url.String())
	if err != nil {
		return nil
	}
	r := gjson.ParseBytes(data)

	var orders []t.Order
	order := t.Order{
		Symbol: r.Get("symbol").String(),
	}
	orders = append(orders, order)
	return orders
}

// GetHistoricalOrders returns historical orders
func (c Client) GetHistoricalOrders(symbol string, limit int) []t.Order {
	var payload strings.Builder
	fmt.Fprintf(&payload, "symbol=%s&limit=%d&timestamp=%d", symbol, limit, now())

	signature := Sign(payload.String(), c.secretKey)

	var url strings.Builder
	fmt.Fprintf(&url, "%s/allOrders?%s&signature=%s", c.baseURL, payload.String(), signature)
	data, err := helper.Get(url.String())
	if err != nil {
		return nil
	}
	r := gjson.ParseBytes(data)

	var orders []t.Order
	order := t.Order{
		Symbol: r.Get("symbol").String(),
	}
	orders = append(orders, order)
	return orders
}

// PlaceOrder sends an order to trade on the exchange
func (c Client) PlaceOrder(o t.Order) *t.Order {
	var payload strings.Builder
	fmt.Fprintf(&payload,
		"symbol=%s&side=%s&type=LIMIT&quantity=%f&price=%f&timestamp=%d",
		o.Symbol, o.Side, o.Qty, o.OpenPrice, now())

	signature := Sign(payload.String(), c.secretKey)

	var url strings.Builder
	fmt.Fprintf(&url, "%s/order?%s&signature=%s", c.baseURL, payload.String(), signature)
	url.WriteString(payload.String())
	resp, err := helper.Post(url.String(), newHeader(c.apiKey))
	if err != nil {
		return nil
	}
	fmt.Println("Response:", resp)
	return &o
}
