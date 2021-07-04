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
	baseURL   string
	apiKey    string
	secretKey string
}

// NewSpotClient returns Binance Spot client
func NewSpotClient(apiKey string, secretKey string) Client {
	return Client{
		baseURL:   "https://api.binance.com/api/v3",
		apiKey:    apiKey,
		secretKey: secretKey,
	}
}

// NewTestSpotClient returns Binance Test Network Spot client
func NewTestSpotClient(apiKey string, secretKey string) Client {
	return Client{
		baseURL:   "https://testnet.binance.vision/api/v3",
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
			Time:   d[0].Int(),
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
func (c Client) GetOrder(o t.Order) *t.Order {
	var payload, url strings.Builder
	fmt.Fprintf(&payload,
		"timestamp=%d&symbol=%s&orderId=%d&origClientOrderId=%s", now(), o.Symbol, o.RefID1, o.RefID2)

	signature := Sign(payload.String(), c.secretKey)

	fmt.Fprintf(&url, "%s/order?%s&signature=%s", c.baseURL, payload.String(), signature)
	data, err := helper.Get(url.String())
	if err != nil {
		return nil
	}
	r := gjson.ParseBytes(data)
	status := r.Get("status").String()
	if o.Status != status {
		o.Status = status
	}
	return &o
}

// GetOpenOrders returns open orders
func (c Client) GetOpenOrders(symbol string) []t.Order {
	var payload, url strings.Builder
	fmt.Fprintf(&payload, "symbol=%s&timestamp=%d", symbol, now())

	signature := Sign(payload.String(), c.secretKey)

	fmt.Fprintf(&url, "%s/openOrders?%s&signature=%s", c.baseURL, payload.String(), signature)
	data, err := helper.Get(url.String())
	if err != nil {
		return nil
	}

	var orders []t.Order
	for _, r := range gjson.ParseBytes(data).Array() {
		order := t.Order{
			Symbol:    symbol,
			RefID1:    r.Get("orderId").Int(),
			RefID2:    r.Get("clientOrderId").String(),
			Side:      r.Get("side").String(),
			Status:    r.Get("status").String(),
			Type:      r.Get("type").String(),
			OpenTime:  r.Get("time").Int(),
			Qty:       r.Get("origQty").Float(),
			OpenPrice: r.Get("price").Float(),
			IsWorking: r.Get("isWorking").Bool(),
		}
		orders = append(orders, order)
	}
	return orders
}

// GetAllOrders returns all account orders; active, canceled, or filled
func (c Client) GetAllOrders(symbol string, limit int, startTime int, endTime int) []t.Order {
	var payload, url strings.Builder
	fmt.Fprintf(&payload, "timestamp=%d&symbol=%s", now(), symbol)

	if limit > 0 {
		fmt.Fprintf(&payload, "&limit=%d", limit)
	}
	if startTime > 0 {
		fmt.Fprintf(&payload, "&startTime=%d", startTime)
	}
	if endTime > 0 {
		fmt.Fprintf(&payload, "&endTime=%d", endTime)
	}

	signature := Sign(payload.String(), c.secretKey)

	fmt.Fprintf(&url, "%s/allOrders?%s&signature=%s", c.baseURL, payload.String(), signature)
	data, err := helper.Get(url.String())
	if err != nil {
		return nil
	}

	var orders []t.Order
	for _, r := range gjson.ParseBytes(data).Array() {
		order := t.Order{
			Symbol:    symbol,
			RefID1:    r.Get("orderId").Int(),
			RefID2:    r.Get("clientOrderId").String(),
			Side:      r.Get("side").String(),
			Status:    r.Get("status").String(),
			Type:      r.Get("type").String(),
			OpenTime:  r.Get("time").Int(),
			Qty:       r.Get("origQty").Float(),
			OpenPrice: r.Get("price").Float(),
			IsWorking: r.Get("isWorking").Bool(),
		}
		orders = append(orders, order)
	}
	return orders
}

// PlaceOrder sends an order to the exchange
func (c Client) PlaceOrder(o t.Order) *t.Order {
	var payload, url strings.Builder
	fmt.Fprintf(&payload, "timestamp=%d&symbol=%s&side=%s&type=%s&quantity=%f",
		now(), o.Symbol, o.Side, o.Type, o.Qty)

	if o.Type == t.OrderTypeLimit || o.Type == t.OrderTypeTP || o.Type == t.OrderTypeSL {
		fmt.Fprintf(&payload, "&timeInForce=GTC")
	}

	if o.Type == t.OrderTypeLimit {
		fmt.Fprintf(&payload, "&price=%f", o.OpenPrice)
	} else if o.Type == t.OrderTypeTP {
		stopPrice := o.ClosePrice
		if o.Side == t.OrderSideBuy {
			stopPrice = o.ClosePrice - o.ClosePrice*0.002
		} else if o.Side == t.OrderSideSell {
			stopPrice = o.ClosePrice + o.ClosePrice*0.002
		}
		fmt.Fprintf(&payload, "&price=%f&stopPrice=%f", o.ClosePrice, stopPrice)
	} else if o.Type == t.OrderTypeSL {
		stopPrice := o.ClosePrice
		if o.Side == t.OrderSideBuy {
			stopPrice = o.ClosePrice + o.ClosePrice*0.002
		} else if o.Side == t.OrderSideSell {
			stopPrice = o.ClosePrice - o.ClosePrice*0.002
		}
		fmt.Fprintf(&payload, "&price=%f&stopPrice=%f", o.ClosePrice, stopPrice)
	}

	signature := Sign(payload.String(), c.secretKey)

	fmt.Fprintf(&url, "%s/order?%s&signature=%s", c.baseURL, payload.String(), signature)
	data, err := helper.Post(url.String(), newHeader(c.apiKey))
	if err != nil {
		return nil
	}
	r := gjson.ParseBytes(data)
	status := r.Get("status").String()
	if status != t.OrderStatusNew && status != t.OrderStatusFilled {
		return nil
	}
	o.Status = status
	o.RefID1 = r.Get("orderId").Int()
	o.RefID2 = r.Get("clientOrderId").String()
	o.OpenTime = r.Get("transactTime").Int()
	price := r.Get("price").Float()
	if o.OpenPrice != price && price > 0 {
		o.OpenPrice = price
	}
	return &o
}

// CancelOrder cancels an order on the exchange
func (c Client) CancelOrder(o t.Order) *t.Order {
	var payload, url strings.Builder
	fmt.Fprintf(&payload, "timestamp=%d&symbol=%s&orderId=%d&origClientOrderId=%s",
		now(), o.Symbol, o.RefID1, o.RefID2)

	signature := Sign(payload.String(), c.secretKey)

	fmt.Fprintf(&url, "%s/order?%s&signature=%s", c.baseURL, payload.String(), signature)
	data, err := helper.Delete(url.String(), newHeader(c.apiKey))
	if err != nil {
		return nil
	}
	status := gjson.ParseBytes(data).Get("status").String()
	if status != t.OrderStatusCanceled {
		return nil
	}
	o.Status = status
	return &o
}
